// Package db wires up the GORM Postgres connection.
package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects to Postgres via GORM, configures the underlying connection
// pool, and pings to verify connectivity. Schema is owned by golang-migrate, so
// AutoMigrate is intentionally not called here.
func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("db: get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return gdb, nil
}

// Close releases the underlying connection pool.
func Close(gdb *gorm.DB) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("db: get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
