package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (s *Server) SetupRoutes() {
	s.echo.GET("/health", s.healthCheck)

}

func (s *Server) healthCheck(c *echo.Context) error {
	// ctx := c.Request().Context()

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})

}
