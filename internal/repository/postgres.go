package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Rasulikus/chat/internal/config"
	"github.com/Rasulikus/chat/internal/model"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type DB struct {
	DB *bun.DB
}

func NewClient(cfg *config.Config) (*DB, error) {
	dsn := cfg.DB.PostgresURL()
	// Open a PostgreSQL database
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	err := sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("cant connect to database: %w", err)
	}
	// Open a PostgreSQL database
	db := bun.NewDB(sqlDB, pgdialect.New())
	// Print all queries to stdout
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	return &DB{
		DB: db,
	}, nil
}

func IsNoRowsError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNotFound
	}
	return err
}
