package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type CustomClaims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*user.UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	// GeneratedToken(userID uuid.UUID, tokenType TokenType) (string, error)
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

func generatedToken(cfg *config.Config, userID uuid.UUID, tokenType TokenType) (string, error) {
	var expirationTime time.Duration
	var secret string
	switch tokenType {
	case AccessToken:
		secret = cfg.JWTAccessSecret
		expirationTime = time.Duration(cfg.AccessTokenExpiration) * time.Minute
	case RefreshToken:
		secret = cfg.JWTRefreshSecret
		expirationTime = time.Duration(cfg.RefreshTokenExpiration) * time.Minute
	default:
		return "", errors.New("invalid token type")
	}

	now := time.Now()
	claims := CustomClaims{
		Type: string(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "chuongpl/money_lover/backend/authentication",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expirationTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		slog.Error("Error generating token", "error", err)
		return "", err
	}
	return signedToken, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userSrevice.GetByEmailAndDeletedAtIsNull(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := generatedToken(s.cfg, user.ID, AccessToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generatedToken(s.cfg, user.ID, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Authenticated: true,
		Token:         token,
		RefreshToken:  refreshToken,
	}, nil
}
