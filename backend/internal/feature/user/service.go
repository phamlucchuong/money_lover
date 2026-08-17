package user

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"
	"context"
	"errors"
	"log/slog"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Service interface {
	Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) ([]*UserResponse, *pkg.PaginationMeta, error)
}

type service struct {
	repo Repository
	cfg  *config.Config
	log  *slog.Logger
}

func NewService(repo Repository, cfg *config.Config, log *slog.Logger) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
		log:  log,
	}
}

func (s *service) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	existing, err := s.repo.GetByEmailAndDeletedAtIsNull(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	user := &User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	if err = s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *service) GetByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}
	return &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *service) GetAllUsers(ctx context.Context, page, pageSize int) ([]*UserResponse, *pkg.PaginationMeta, error) {
	offset := (page - 1) * pageSize
	users, total, err := s.repo.GetAll(ctx, offset, pageSize)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, &UserResponse{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	meta := &pkg.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: int(total),
		TotalPages: totalPages,
	}

	return responses, meta, nil
}
