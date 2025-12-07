package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Rasulikus/chat/internal/config"
	"github.com/Rasulikus/chat/internal/migrator"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var (
	testDB      *bun.DB
	testDSN     string
	truncateSQL = `
	TRUNCATE TABLE
		rooms,
	    messages
	RESTART IDENTITY CASCADE;
	`
)

func DB() *bun.DB {
	if testDB == nil {
		cfg := config.DBConfig{
			User: "admin",
			Pass: "mypassword",
			Host: "localhost",
			Port: "5432",
			Name: "chat_test",
		}

		var err error
		testDB, err = newClient(&cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
	return testDB
}

func newClient(cfg *config.DBConfig) (*bun.DB, error) {
	EnsureTestDatabase(cfg)

	testDSN = cfg.PostgresURL()
	// Open a PostgreSQL database
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(testDSN)))
	err := sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("cant connect to database: %w", err)
	}
	// Open a PostgreSQL database
	db := bun.NewDB(sqlDB, pgdialect.New())
	// Print all queries to stdout
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	return db, nil
}

func EnsureTestDatabase(cfg *config.DBConfig) {
	adminCfg := *cfg
	adminCfg.Name = "postgres"

	adminDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminCfg.PostgresURL())))
	if err := adminDB.Ping(); err != nil {
		log.Fatalf("ping admin db: %v", err)
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		log.Fatalf("ping admin db: %v", err)
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);`
	if err := adminDB.QueryRow(query, cfg.Name).Scan(&exists); err != nil {
		log.Fatalf("check database exists: %v", err)
	}

	if exists {
		return
	}

	createSQL := fmt.Sprintf(`CREATE DATABASE %s`, cfg.Name)
	if _, err := adminDB.Exec(createSQL); err != nil {
		log.Fatalf("create database %s: %v", cfg.Name, err)
	}
}

func CloseDB() {
	if testDB != nil {
		err := testDB.Close()
		if err != nil {
			log.Print(err)
			return
		}
		testDB = nil
	}
}

func RecreateTables() {
	if testDB == nil {
		DB()
	}

	opts := migrator.NewOptions(testDSN, "")

	err := migrator.Down(*opts)
	if err != nil {
		log.Fatal(err)
	}

	err = migrator.Up(*opts)
	if err != nil {
		log.Fatal(err)
	}

	CleanDB(context.Background())
}

func CleanDB(ctx context.Context) {
	_, err := testDB.ExecContext(ctx, truncateSQL)
	if err != nil {
		log.Fatalf("error clean db: %v", err)
	}
}
