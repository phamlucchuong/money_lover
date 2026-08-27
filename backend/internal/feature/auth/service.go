package auth

import (
	"context"
	"errors"
	"log/slog"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*user.UserResponse, error)
}

type service struct {
	userSrevice user.Service
	cfg         *config.Config
	log         *slog.Logger
}

func NewService(userService user.Service, cfg *config.Config, log *slog.Logger) Service {
	return &service{
		userSrevice: userService,
		cfg:         cfg,
		log:         log,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*user.UserResponse, error) {
	existing, err := s.userSrevice.GetByEmailAndDeletedAtIsNull(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, user.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userResp, err := s.userSrevice.Create(ctx, &user.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, user.ErrUserAlreadyExists
		}
		return nil, err
	}

	return &user.UserResponse{
		ID:    userResp.ID,
		Name:  userResp.Name,
		Email: userResp.Email,
	}, nil
}

// func (s *service) Login(ctx context.Context, email, password string) (AuthResponse, error) {
// }
