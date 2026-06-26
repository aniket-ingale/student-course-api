// Command server is the entry point for the student-course-api HTTP service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aniket/student-course-api/internal/config"
	"github.com/aniket/student-course-api/internal/db"
	"github.com/aniket/student-course-api/internal/handler"
	"github.com/aniket/student-course-api/internal/repository"
	"github.com/aniket/student-course-api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	setupLogger(cfg.LogLevel)

	// Bound startup (DB connect/ping) so a dead DB doesn't hang forever.
	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gdb, err := db.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := db.Close(gdb); cerr != nil {
			slog.Error("close db", "error", cerr)
		}
	}()

	repo := repository.NewStudentRepository(gdb)
	svc := service.NewStudentService(repo)
	studentHandler := handler.NewStudentHandler(svc)

	health := func(ctx context.Context) error {
		sqlDB, err := gdb.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}

	router := handler.NewRouter(studentHandler, health)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run the server until a signal arrives, then shut down gracefully.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		slog.Info("server stopped cleanly")
	}

	return nil
}

func setupLogger(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}
