package repository

import "context"

type TestRepository interface {
	GetGreeting(ctx context.Context) (string, error)

	GetMessages(ctx context.Context) (string, error)

	SaveMessage(
		ctx context.Context,
		message string,
	) error
}
