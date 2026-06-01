package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPostRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPostRepository(db *pgxpool.Pool) PostRepository {
	return &postgresPostRepository{
		db: db,
	}
}

func (r *postgresPostRepository) Create(ctx context.Context, author_id int64, content string) (*Post, error) {
	query := `
		INSERT INTO posts (author_id, content)
		VALUES ($1, $2)
		RETURNING id, author_id, content
	`

	post := &Post{}
	err := r.db.QueryRow(ctx, query, author_id, content).Scan(&post.ID, &post.AuthorID, &post.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return post, nil
}
