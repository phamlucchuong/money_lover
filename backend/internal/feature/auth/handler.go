package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc Service
	cfg *config.Config
	log *slog.Logger
}

func NewHandler(svc Service, cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		cfg: cfg,
		log: log,
	}
}

func (h *Handler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid request data")
	}
	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, "invalid request data")
	}

	resp, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserAlreadyExists):
			return pkg.JSONError(c, http.StatusConflict, pkg.CodeConflict, "user already exists")
		default:
			h.log.Error("create user failed", slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")

		}
	}

	return pkg.JSONCreated(c, resp)
}

func (h *Handler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid request data")
	}
	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, "invalid request data")
	}

	resp, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserAlreadyExists):
			return pkg.JSONError(c, http.StatusConflict, pkg.CodeConflict, "user already exists")
		default:
			h.log.Error("login failed", slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")

		}
	}

	setAuthCookies(c, resp.AccessToken, resp.RefreshToken, h.cfg)
	if h.cfg.Environment == "production" {
		c.Response().Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}

	return pkg.JSONOK(c, resp)
}

func (h *Handler) Logout(c *echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return pkg.JSONError(c, http.StatusUnauthorized, pkg.CodeUnauthorized, "missing authorization header")
	}

	err := h.svc.Logout(c.Request().Context(), tokenString)
	if err != nil {
		h.log.Error("logout failed", slog.String("error", err.Error()))
		return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")
	}

	ClearAuthCookies(c)

	return pkg.JSONOK(c, map[string]string{"message": "logged out successfully"})
}

func setAuthCookies(c *echo.Context, accessToken, refreshToken string, cfg *config.Config) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(cfg.AccessTokenExpiration) * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(cfg.RefreshTokenExpiration) * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(refreshCookie)
}

func ClearAuthCookies(c *echo.Context) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}
