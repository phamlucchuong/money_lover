package server

import (
	"context"
	"log/slog"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"chuongpl/quan-ly-chi-tieu/internal/platform/cache"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	echo        *echo.Echo
	cfg         *config.Config
	log         *slog.Logger
	gormDB      *gorm.DB
	rdb         *redis.Client
	userHandler *user.Handler
	userService user.Service
	authHandler *auth.Handler
	authService auth.Service
}

func NewServer(cfg *config.Config, log *slog.Logger, gormDB *gorm.DB, rdb *redis.Client) *Server {
	e := echo.New()
	s := &Server{
		echo:   e,
		cfg:    cfg,
		log:    log,
		gormDB: gormDB,
		rdb:    rdb,
	}

	redisCache := cache.NewRedisCache(rdb)

	userRepo := user.NewRepository(gormDB)
	s.userService = user.NewService(userRepo, cfg, log)
	s.userHandler = user.NewHandler(s.userService, cfg, log)

	s.authService = auth.NewService(s.userService, redisCache, cfg, log)
	s.authHandler = auth.NewHandler(s.authService, cfg, log)

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
