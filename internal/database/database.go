package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arlinggacr/BtechDevCases/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseConfig config.DatabaseConfig) (*sql.DB, error) {
	if databaseConfig.URL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	database, err := sql.Open("pgx", databaseConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return database, nil
}

func InitializeSchema(database *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`
	if _, err := database.Exec(query); err != nil {
		return fmt.Errorf("initialize users table: %w", err)
	}
	return nil
}
