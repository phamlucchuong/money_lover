package wallet

import (
	"log/slog"

	"chuongpl/quan-ly-chi-tieu/internal/config"
)

type Service interface{}

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
