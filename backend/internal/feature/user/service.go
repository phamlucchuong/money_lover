package user

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Service interface {
	GetByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*UserResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) ([]*UserResponse, *pkg.PaginationMeta, error)
	Update(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, userID uuid.UUID) error
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

func (s *service) GetByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmailAndDeletedAtIsNull(ctx, email)
}

func (s *service) Create(ctx context.Context, user *User) (*UserResponse, error) {
	if err := s.repo.Create(ctx, user); err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
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

func (s *service) Update(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID) error {
	err := s.repo.Delete(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	return err
}
