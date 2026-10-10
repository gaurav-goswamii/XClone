package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// ConnectPostgres connects to PostgreSQL and verifies the connection
func ConnectPostgres(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Migrate executes the initial database schema migrations
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(64) PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		birth_date VARCHAR(20) NOT NULL,
		avatar_url TEXT DEFAULT '',
		bio TEXT DEFAULT '',
		is_verified BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));
	CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));

	CREATE TABLE IF NOT EXISTS verification_records (
		email VARCHAR(255) PRIMARY KEY,
		code VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		attempts INT DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS posts (
		id VARCHAR(64) PRIMARY KEY,
		user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		content TEXT NOT NULL,
		media_url TEXT DEFAULT '',
		likes INT DEFAULT 0,
		retweets INT DEFAULT 0,
		replies INT DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);
	CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts (created_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
