// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds all runtime settings for the service.
type Config struct {
	// DatabaseURL is the Postgres DSN used by GORM.
	DatabaseURL string
	// HTTPPort is the port the HTTP server listens on.
	HTTPPort string
	// LogLevel is one of debug, info, warn, error.
	LogLevel string
}

// Load reads configuration from the environment. It prefers DATABASE_URL and
// falls back to building a DSN from the discrete DB_* variables. It returns an
// error if required values are missing so the process can fail fast.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPPort:    getEnvDefault("HTTP_PORT", "8080"),
		LogLevel:    getEnvDefault("LOG_LEVEL", "info"),
	}

	if cfg.DatabaseURL == "" {
		dsn, err := dsnFromParts()
		if err != nil {
			return Config{}, err
		}
		cfg.DatabaseURL = dsn
	}

	return cfg, nil
}

// dsnFromParts assembles a Postgres URL from the discrete DB_* variables.
func dsnFromParts() (string, error) {
	host := os.Getenv("DB_HOST")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")

	var missing []string
	if host == "" {
		missing = append(missing, "DB_HOST")
	}
	if name == "" {
		missing = append(missing, "DB_NAME")
	}
	if user == "" {
		missing = append(missing, "DB_USER")
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("config: set DATABASE_URL or all of %s", strings.Join(missing, ", "))
	}

	port := getEnvDefault("DB_PORT", "5432")
	sslmode := getEnvDefault("DB_SSLMODE", "disable")
	password := os.Getenv("DB_PASSWORD")

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
