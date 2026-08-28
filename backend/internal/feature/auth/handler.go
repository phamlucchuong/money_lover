package auth

import (
	"errors"
	"log/slog"
	"net/http"

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

	return pkg.JSONOK(c, resp)
}
