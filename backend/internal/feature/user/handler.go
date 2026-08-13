package user

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc ServiceInterface
	cfg *config.Config
	log *slog.Logger
}

func NewHandler(svc ServiceInterface, cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		cfg: cfg,
		log: log,
	}
}

func (h *Handler) CreateUser(c *echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadeRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, err.Error())
	}

	resp, err := h.svc.CreateUser(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			return pkg.JSONError(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		default:
			h.log.Error("create user failed", slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, err.Error())
		}
	}

	return pkg.JSONCreated(c, resp)
}
