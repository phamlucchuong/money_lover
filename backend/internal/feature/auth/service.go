package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"chuongpl/quan-ly-chi-tieu/internal/platform/cache"
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
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, jti string, ttl time.Duration) error
}

type service struct {
	userSrevice user.Service
	redisCache  cache.Cache
	cfg         *config.Config
	log         *slog.Logger
}

func NewService(userService user.Service, redisCache cache.Cache, cfg *config.Config, log *slog.Logger) Service {
	return &service{
		userSrevice: userService,
		redisCache:  redisCache,
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
		AccessToken:   token,
		RefreshToken:  refreshToken,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid || claims.Type != string(RefreshToken) {
		return nil, errors.New("invalid refresh token")
	}

	// Reuse detection: a refresh token that already shows up in the
	// blacklist has been rotated before. Either the client retried or
	// the chain was stolen — reject either way.
	blacklisted, err := s.redisCache.Exists(ctx, blacklistKey(claims.ID))
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, errors.New("refresh token has been revoked")
	}

	// Rotate: blacklist the old refresh jti with its remaining TTL so it
	// cannot be reused. Done before issuing new tokens so a failure here
	// blocks replay attempts at the cost of denying a valid refresh.
	if claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			if err := s.redisCache.Set(ctx, blacklistKey(claims.ID), "true", remaining); err != nil {
				return nil, err
			}
		}
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := generatedToken(s.cfg, userID, AccessToken)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generatedToken(s.cfg, userID, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Authenticated: true,
		AccessToken:   newAccessToken,
		RefreshToken:  newRefreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, jti string, ttl time.Duration) error {
	blacklistKey := "blacklist:" + jti
	if err := s.redisCache.Set(ctx, blacklistKey, "true", ttl); err != nil {
		return err
	}

	return nil
}
