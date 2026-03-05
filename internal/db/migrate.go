package db

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
)

// RunMigrations applies all pending SQL migration files from the given directory.
// It is safe to call on every startup — already-applied migrations are skipped.
// migrationsPath example: "file://migrations"
func RunMigrations(databaseURL, migrationsPath string) error {
	migrateURL, err := toMigrateURL(databaseURL)
	if err != nil {
		return fmt.Errorf("migrate init: build url: %w", err)
	}

	m, err := migrate.New(migrationsPath, migrateURL)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migrations: no new migrations to apply")
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}

	log.Println("migrations: all migrations applied successfully")
	return nil
}

// toMigrateURL converts any pgx-accepted connection string (DSN key=value or
// postgres:// URL) into the pgx5:// URL format required by golang-migrate.
func toMigrateURL(databaseURL string) (string, error) {
	cfg, err := pgconn.ParseConfig(databaseURL)
	if err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme: "pgx5",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   "/" + cfg.Database,
	}

	q := url.Values{}
	if sslmode, ok := cfg.RuntimeParams["sslmode"]; ok {
		q.Set("sslmode", sslmode)
	} else {
		q.Set("sslmode", "require")
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
