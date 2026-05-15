package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

	return err
}