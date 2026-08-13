package user

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type ServiceInterface interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
}

type Service struct {
	repo RepositoryInterface
	cfg  *config.Config
	log  *slog.Logger
}

func NewService(repo RepositoryInterface, cfg *config.Config, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
		log:  log,
	}
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	existing, err := s.repo.GetUserByEmailAndDeletedAtIsNull(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	user := &User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	if err = s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// func (s *Service) GetAll() ([]*User, error) {

// }
