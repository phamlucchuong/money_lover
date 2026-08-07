package logger

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	var handler slog.Handler

	if env == "development" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}

	return slog.New(handler)
}
