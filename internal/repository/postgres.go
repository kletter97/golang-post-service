package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Message struct {
	ID        int64     `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) TestRepository {
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

	_, err := r.db.ExecContext(ctx, query, message)

	log.Printf("Something posted")

	return err
}

func (r *postgresRepository) GetMessages(
	ctx context.Context,
) (string, error) {
	query := `SELECT id, message, created_at FROM messages ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Message, &msg.CreatedAt); err != nil {
			return "", fmt.Errorf("scan failed: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("rows error: %w", err)
	}

	jsonData, err := json.Marshal(messages)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	return string(jsonData), nil
}
