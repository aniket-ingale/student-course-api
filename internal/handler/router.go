package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// HealthChecker reports whether dependencies (e.g. the DB) are reachable.
type HealthChecker func(ctx context.Context) error

// NewRouter wires the routes onto a ServeMux and applies global middleware.
func NewRouter(students *StudentHandler, health HealthChecker) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students/{id}", students.Get)
	mux.HandleFunc("GET /students", students.List)
	mux.HandleFunc("GET /healthz", healthHandler(health))

	return recoverMW(logMW(mux))
}

// healthHandler returns 200 when the health check passes, 503 otherwise.
func healthHandler(check HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := check(ctx); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// statusRecorder captures the response status for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logMW logs each request with method, path, status, and duration.
func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}

// recoverMW turns a panic into a 500 instead of crashing the server.
func recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(r.Context(), "panic recovered",
					"method", r.Method, "path", r.URL.Path, "panic", rec)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
