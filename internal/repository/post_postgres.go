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
	err := r.db.QueryRow(ctx, query, author_id, content).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.Status,
		&post.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (r *postgresPostRepository) GetPostsByAuthor(ctx context.Context, author_id int64) ([]Post, error) {
	query := `
        SELECT id, author_id, content, status, created_at
        FROM posts
        WHERE author_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(ctx, query, author_id)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Content,
			&p.Status,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return posts, nil
}
