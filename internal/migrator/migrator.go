package migrator

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Options struct {
	DSN string

	MigrationsDir string
}

func NewOptions(dsn string, migrationDir string) *Options {
	return &Options{
		DSN:           dsn,
		MigrationsDir: migrationDir,
	}
}

func defaultMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(root, "migrations")
}

func Up(opts Options) error {
	if opts.MigrationsDir == "" {
		opts.MigrationsDir = defaultMigrationsDir()
	}
	if opts.DSN == "" {
		return fmt.Errorf("migrator: DSN is empty")
	}

	sourceURL := "file://" + opts.MigrationsDir

	m, err := migrate.New(sourceURL, opts.DSN)
	if err != nil {
		return fmt.Errorf("migrator: create migrate: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrator: up: %w", err)
	}

	log.Println("migrator: up done")
	return nil
}

func Down(opts Options) error {
	if opts.MigrationsDir == "" {
		opts.MigrationsDir = defaultMigrationsDir()
	}
	if opts.DSN == "" {
		return fmt.Errorf("migrator: DSN is empty")
	}

	sourceURL := "file://" + opts.MigrationsDir

	m, err := migrate.New(sourceURL, opts.DSN)
	if err != nil {
		return fmt.Errorf("migrator: create migrate: %w", err)
	}
	defer m.Close()

	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrator: down: %w", err)
	}

	log.Println("migrator: down done")
	return nil
}
