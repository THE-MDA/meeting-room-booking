package migrator

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	migrationsPath string
	databaseURL    string
}

func New(migrationsPath, databaseURL string) *Migrator {
	return &Migrator{
		migrationsPath: migrationsPath,
		databaseURL:    databaseURL,
	}
}

func (m *Migrator) Up() error {
	slog.Info("Running migrations",
		"path", m.migrationsPath,
		"database_url", maskPassword(m.databaseURL),
	)

	mInstance, err := migrate.New(
		"file://"+m.migrationsPath,
		m.databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer mInstance.Close()

	if err := mInstance.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

func (m *Migrator) Down() error {
	mInstance, err := migrate.New(
		"file://"+m.migrationsPath,
		m.databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer mInstance.Close()

	if err := mInstance.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	slog.Info("Migrations rolled back successfully")
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	mInstance, err := migrate.New(
		"file://"+m.migrationsPath,
		m.databaseURL,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer mInstance.Close()

	version, dirty, err := mInstance.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

func maskPassword(url string) string {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' && i+1 < len(url) {
			for j := i + 1; j < len(url); j++ {
				if url[j] == '@' {
					return url[:i+1] + "***" + url[j:]
				}
			}
		}
	}
	return url
}
