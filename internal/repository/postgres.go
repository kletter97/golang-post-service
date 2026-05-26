package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        int64     `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(
	db *pgxpool.Pool,
) TestRepository {

	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) GetGreeting(
	ctx context.Context,
) (string, error) {

	return "Hello from PostgreSQL!", nil
}

func (r *postgresRepository) SaveMessage(
	ctx context.Context,
	message string,
) error {

	query := `
		INSERT INTO messages (message)
		VALUES ($1)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		message,
	)

	log.Printf("Something posted")

	return err
}

func (r *postgresRepository) GetMessages(
	ctx context.Context,
) (string, error) {
	query := `SELECT id, message, created_at FROM messages ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	messages, err := pgx.CollectRows(rows, pgx.RowToStructByName[Message])
	if err != nil {
		return "", fmt.Errorf("collect rows failed: %w", err)
	}

	jsonData, err := json.Marshal(messages)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	return string(jsonData), nil
}
