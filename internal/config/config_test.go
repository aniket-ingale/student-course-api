package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("prefers DATABASE_URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db?sslmode=disable")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DatabaseURL != "postgres://u:p@h:5432/db?sslmode=disable" {
			t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
		}
		if cfg.HTTPPort != "8080" {
			t.Fatalf("HTTPPort default = %q, want 8080", cfg.HTTPPort)
		}
	})

	t.Run("builds DSN from parts", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_NAME", "student_course")
		t.Setenv("DB_USER", "student")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_PORT", "5433")
		t.Setenv("DB_SSLMODE", "require")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, want := range []string{"student:secret@", "localhost:5433", "/student_course", "sslmode=require"} {
			if !strings.Contains(cfg.DatabaseURL, want) {
				t.Fatalf("DSN %q missing %q", cfg.DatabaseURL, want)
			}
		}
	})

	t.Run("errors when required parts missing", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("DB_HOST", "")
		t.Setenv("DB_NAME", "")
		t.Setenv("DB_USER", "")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
