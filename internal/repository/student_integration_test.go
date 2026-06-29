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

	"github.com/aniket/student-course-api/internal/model"
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

func TestStudentRepository_Create(t *testing.T) {
	gdb := setupDB(t)
	repo := NewStudentRepository(gdb)
	ctx := context.Background()

	t.Run("Create assigns a new studentId and persists the row", func(t *testing.T) {
		in := model.Student{Name: "Margaret Hamilton", Address: "1 Apollo Way, Boston", Grade: 12}

		got, err := repo.Create(ctx, in)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if got.StudentID == 0 {
			t.Fatalf("StudentID = 0, want DB-assigned non-zero id")
		}
		if got.Name != in.Name || got.Address != in.Address || got.Grade != in.Grade {
			t.Fatalf("returned student = %+v, want fields to match input %+v", got, in)
		}

		// Round-trip: the row is actually readable by the assigned id.
		fetched, err := repo.GetByID(ctx, got.StudentID)
		if err != nil {
			t.Fatalf("GetByID after Create: %v", err)
		}
		if fetched.Name != in.Name {
			t.Fatalf("fetched name = %q, want %q", fetched.Name, in.Name)
		}
	})

	t.Run("Create ignores any caller-supplied StudentID", func(t *testing.T) {
		// student_id is GENERATED ALWAYS AS IDENTITY; a supplied id must not win.
		in := model.Student{StudentID: 99999, Name: "Hedy Lamarr", Address: "9 Spread Spectrum Rd", Grade: 11}

		got, err := repo.Create(ctx, in)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if got.StudentID == 99999 {
			t.Fatalf("StudentID = 99999, want a DB-assigned id (caller id must be ignored)")
		}
	})
}

func TestStudentRepository_Update(t *testing.T) {
	gdb := setupDB(t)
	repo := NewStudentRepository(gdb)
	ctx := context.Background()

	t.Run("Update overwrites mutable fields", func(t *testing.T) {
		// Seed row id 1 is Ada Lovelace, grade 10.
		in := model.Student{StudentID: 1, Name: "Ada King", Address: "99 Countess Court", Grade: 0}

		got, err := repo.Update(ctx, in)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.Name != "Ada King" || got.Address != "99 Countess Court" || got.Grade != 0 {
			t.Fatalf("returned student = %+v, want updated fields", got)
		}

		// Grade 0 must actually persist (zero-value field must not be skipped).
		fetched, err := repo.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID after Update: %v", err)
		}
		if fetched.Grade != 0 {
			t.Fatalf("persisted grade = %d, want 0 (zero value must be written)", fetched.Grade)
		}
		if fetched.Name != "Ada King" {
			t.Fatalf("persisted name = %q, want Ada King", fetched.Name)
		}
	})

	t.Run("Update of a missing id returns ErrNotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, model.Student{StudentID: 99999, Name: "Nobody", Address: "x", Grade: 1})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

func TestStudentRepository_Delete(t *testing.T) {
	gdb := setupDB(t)
	repo := NewStudentRepository(gdb)
	ctx := context.Background()

	t.Run("Delete removes an existing row", func(t *testing.T) {
		// Seed row id 2 is Alan Turing.
		if err := repo.Delete(ctx, 2); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 2)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("GetByID after Delete: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete of a missing id returns ErrNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, 99999)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
