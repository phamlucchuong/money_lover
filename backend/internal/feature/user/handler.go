package user

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"log/slog"
)

type Hander struct {
	cfg *config.Config
	log *slog.Logger
}
