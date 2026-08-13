package server

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type Server struct {
	echo        *echo.Echo
	cfg         *config.Config
	log         *slog.Logger
	gormDB      *gorm.DB
	userHandler *user.Handler
	userService *user.Service
}

func NewServer(cfg *config.Config, log *slog.Logger, gormDB *gorm.DB) *Server {
	e := echo.New()
	s := &Server{
		echo:   e,
		cfg:    cfg,
		log:    log,
		gormDB: gormDB,
	}

	userRepo := user.NewRepository(gormDB)
	s.userService = user.NewService(userRepo, cfg, log)
	s.userHandler = user.NewHandler(s.userService, cfg, log)

	s.echo.Validator = NewValidator()
	s.SetupRoutes()

	return s
}

func (s *Server) Start(ctx context.Context) error {
	s.log.Info("server starting", slog.String("port", s.cfg.Port))

	sc := echo.StartConfig{
		Address:         ":" + s.cfg.Port,
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}

	return sc.Start(ctx, s.echo)
}
