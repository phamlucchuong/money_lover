package auth

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"log/slog"
)

type Service interface {
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
