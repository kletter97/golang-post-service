package repository

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func InitTables(
	ctx context.Context,
	db *sql.DB,
) error {

	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)
	`

	_, err := db.ExecContext(ctx, query)

	return err
}