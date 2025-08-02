package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // migrate filesystem driver
	_ "github.com/lib/pq"                                // postgres driver
)

// Migrate - helps migrate database schema using migration files in the directory
func (s *PostgresStorage) Migrate(
	_ context.Context, // context for request
	directory string, // directory containing migration files
) error {
	// create a driver instance
	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("database migration initilisation failed: %w", err)
	}

	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("failed to get file info for migration directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("migration directory %s is not a directory", directory)
	}

	// create a migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+directory, // Use full path for DB migration files directory
		"postgres",          // we are using postgres driver
		driver,              // initialised driver
	)
	if err != nil {
		return fmt.Errorf("failed to create a migration DB instance: %w", err)
	}

	// run the migrations
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	s.log.Info("Database migration completed successfully", "directory", directory)
	return nil
}
