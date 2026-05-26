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

// ===== User Service =====

type UserService interface {
	Register(ctx context.Context, email, password string) (*repository.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUserByID(ctx context.Context, id int64) (*repository.User, error)
}

type userService struct {
	userRepo     repository.UserRepository
	jwtSecret    string
	passwordHash PasswordHasher
	tokenGenerator TokenGenerator
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

type TokenGenerator interface {
	GenerateToken(userID int64, email string) (string, error)
}

func NewUserService(
	userRepo repository.UserRepository,
	jwtSecret string,
	passwordHash PasswordHasher,
	tokenGenerator TokenGenerator,
) UserService {
	return &userService{
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		passwordHash: passwordHash,
		tokenGenerator: tokenGenerator,
	}
}

func (s *userService) Register(ctx context.Context, email, password string) (*repository.User, error) {
	hash, err := s.passwordHash.Hash(password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, email, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !s.passwordHash.Verify(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenGenerator.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*repository.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

var ErrInvalidCredentials = &serviceError{"invalid credentials"}

type serviceError struct {
	message string
}

func (e *serviceError) Error() string {
	return e.message
}
