package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectDB() error {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return err
	}

	// Adjust pool config for standard web api deployment
	config.MaxConns = 25
	config.MaxConnIdleTime = 30 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	return Pool.Ping(context.Background())
}

func Migrate() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			credits INT DEFAULT 10,
			stripe_customer_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS checks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			check_type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS violations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			check_id UUID NOT NULL REFERENCES checks(id) ON DELETE CASCADE,
			issue TEXT NOT NULL,
			severity VARCHAR(50) NOT NULL,
			fix TEXT NOT NULL,
			rule_type VARCHAR(50) NOT NULL
		);`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS credits INT DEFAULT 10;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);`,
	}

	for _, query := range queries {
		_, err := Pool.Exec(context.Background(), query)
		if err != nil {
			log.Printf("Migration failed on query: %s", query)
			return err
		}
	}
	return nil
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
