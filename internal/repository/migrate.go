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
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)
	`

	_, err := db.Exec(ctx, query)

	return err
}