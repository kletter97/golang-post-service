package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitTables(
	ctx context.Context,
	db *pgxpool.Pool,
) error {

	query := `
	CREATE TYPE post_status AS ENUM ('draft', 'pending', 'published', 'rejected');

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
	);

	CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		author_id INTEGER REFERENCES users (id),
		content post_status NOT NULL DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT NOW()
	);
	`

	_, err := db.Exec(ctx, query)

	return err
}
