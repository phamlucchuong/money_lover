package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (s *Server) setupMiddleware() {
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestID())

	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: s.cfg.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodOptions,
		},

		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, echo.HeaderAccept},
		AllowCredentials: true,
	}))

	s.echo.Use(s.requestLogger())
}

func (s *Server) requestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			status := 0
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = resp.Status
			}

			s.log.Info(
				"request",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.Int("status", status),
				slog.Duration("latency", time.Since(start)),
				slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			)

			return err
		}
	}
}
