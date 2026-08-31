package server

import (
	"net/http"

	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"

	"github.com/labstack/echo/v5"
)

func (s *Server) SetupRoutes() {
	s.echo.GET("/health", s.healthCheck)

	v1 := s.echo.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.POST("/register", s.authHandler.Register)
	authGroup.POST("/login", s.authHandler.Login)
	authGroup.POST("/logout", s.authHandler.Logout, auth.JWTAuth(s.cfg, s.log, s.blacklistChecker))

	userGroup := v1.Group("/users", auth.JWTAuth(s.cfg, s.log, s.blacklistChecker))
	userGroup.GET("", s.userHandler.GetAllUsers)
	userGroup.GET("/:id", s.userHandler.GetUserByID)
	userGroup.PUT("/:id", s.userHandler.UpdateUser)
	userGroup.DELETE("/:id", s.userHandler.DeleteUser)
}

func (s *Server) healthCheck(c *echo.Context) error {
	ctx := c.Request().Context()

	sqlDB, err := s.gormDB.DB()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":   "unhealthy",
			"database": err.Error(),
		})
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":   "unhealthy",
			"database": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}
