//go:build integration

package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDB starts a throwaway Postgres container, applies the migrations, and
// returns a connected GORM DB plus a cleanup func.
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("student_course"),
		tcpostgres.WithUsername("student"),
		tcpostgres.WithPassword("student"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = pgContainer.Terminate(context.Background())
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	gdb, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	applyMigrations(t, gdb)
	return gdb
}

// applyMigrations runs every *.up.sql in the migrations dir in name order.
func applyMigrations(t *testing.T, gdb *gorm.DB) {
	t.Helper()
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var ups []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".sql" && hasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if err := gdb.Exec(string(sqlBytes)).Error; err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestStudentRepository_Integration(t *testing.T) {
	gdb := setupDB(t)
	repo := NewStudentRepository(gdb)
	ctx := context.Background()

	t.Run("List returns seeded students", func(t *testing.T) {
		students, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(students) != 5 {
			t.Fatalf("len = %d, want 5 (seeded)", len(students))
		}
		// Ordered by student_id ascending.
		for i := 1; i < len(students); i++ {
			if students[i-1].StudentID > students[i].StudentID {
				t.Fatalf("not ordered by student_id: %+v", students)
			}
		}
	})

	t.Run("GetByID returns a known student", func(t *testing.T) {
		// Seed assigns ids starting at 1; the first row is Ada Lovelace.
		s, err := repo.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if s.Name != "Ada Lovelace" {
			t.Fatalf("name = %q, want Ada Lovelace", s.Name)
		}
	})

	t.Run("GetByID missing returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
