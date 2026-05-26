package service

import (
	"context"

	"post-service/internal/repository"
)

type TestService interface {
	GetGreeting(
		ctx context.Context,
	) (string, error)

	GetMessages(
		ctx context.Context,
	) (string, error)

	SaveMessage(
		ctx context.Context,
		message string,
	) error
}

type testService struct {
	repo repository.TestRepository
}

func NewTestService(
	repo repository.TestRepository,
) TestService {

	return &testService{
		repo: repo,
	}
}

func (s *testService) GetGreeting(
	ctx context.Context,
) (string, error) {

	return s.repo.GetGreeting(ctx)
}

func (s *testService) GetMessages(
	ctx context.Context,
) (string, error) {

	return s.repo.GetMessages(ctx)
}

func (s *testService) SaveMessage(
	ctx context.Context,
	message string,
) error {

	return s.repo.SaveMessage(
		ctx,
		message,
	)
}
