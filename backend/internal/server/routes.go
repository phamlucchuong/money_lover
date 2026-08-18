package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (s *Server) SetupRoutes() {
	s.echo.GET("/health", s.healthCheck)

	v1 := s.echo.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.POST("/register", s.userHandler.CreateUser)

	userGroup := v1.Group("/users")
	userGroup.GET("", s.userHandler.GetAllUsers)
	userGroup.GET("/:id", s.userHandler.GetUserByID)
	userGroup.PUT("/:id", s.userHandler.UpdateUser)
	userGroup.DELETE("/:id", s.userHandler.DeleteUser)
}

func (s *Server) healthCheck(c *echo.Context) error {
	// ctx := c.Request().Context()

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})

}
