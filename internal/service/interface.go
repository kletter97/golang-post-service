package service

import "post-service/internal/repository"


type TestService interface {
	GetGreeting() (string, error)
}

type testService struct {
	repo repository.TestRepository
}

func NewTestService(repo repository.TestRepository) TestService {
	return &testService{
		repo: repo,
	}
}

func (s *testService) GetGreeting() (string, error) {
	return s.repo.GetGreeting()
}
