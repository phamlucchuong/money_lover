package main

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/logger"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"
	"chuongpl/quan-ly-chi-tieu/internal/server"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("fail to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.Environment)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gormDB, err := db.NewGormConfig(cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresName, cfg.PostgresPort)
	if err != nil {
		slog.Error("fail to create gorm config", slog.Any("error", err))
		os.Exit(1)
	}

	srv := server.NewServer(cfg, log, gormDB)

	log.Info("shutting down server", slog.String("port", cfg.Port))
	if err := srv.Start(ctx); err != nil {
		log.Error("fail to start server", slog.Any("error", err))
		os.Exit(1)
	}
}
