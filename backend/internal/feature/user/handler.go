package user

import (
	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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

func (h *Handler) CreateUser(c *echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid request data")
	}
	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, "invalid request data")
	}

	resp, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			return pkg.JSONError(c, http.StatusConflict, pkg.CodeConflict, "user already exists")
		default:
			h.log.Error("create user failed", slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")

		}
	}

	return pkg.JSONCreated(c, resp)
}

func (h *Handler) GetUserByID(c *echo.Context) error {
	strID := c.Param("id")
	userID, err := uuid.Parse(strID)
	if err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid user ID")
	}

	resp, err := h.svc.GetByID(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return pkg.JSONError(c, http.StatusNotFound, pkg.CodeNotFound, "user not found")
		default:
			h.log.Error("get user by ID failed", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")
		}
	}

	return pkg.JSONOK(c, resp)
}

func (h *Handler) GetAllUsers(c *echo.Context) error {
	page := pkg.ParseIntQuery(c, "page", 1)
	pageSize := pkg.ParseIntQuery(c, "page_size", 20)

	if page < 1 {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "page must be >= 1")
	}
	if pageSize < 1 || pageSize > 100 {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "page_size must be between 1 and 100")
	}

	users, meta, err := h.svc.GetAllUsers(c.Request().Context(), page, pageSize)
	if err != nil {
		h.log.Error("get all users failed", slog.String("error", err.Error()))
		return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")
	}

	return pkg.JSONList(c, users, *meta)
}

func (h *Handler) UpdateUser(c *echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid user ID")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid request data")
	}

	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, "invalid request data")
	}

	resp, err := h.svc.Update(c.Request().Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return pkg.JSONError(c, http.StatusNotFound, pkg.CodeNotFound, "user not found")
		case errors.Is(err, ErrUserAlreadyExists):
			return pkg.JSONError(c, http.StatusConflict, pkg.CodeConflict, "user already exists")
		default:
			h.log.Error("update user failed", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")
		}
	}

	return pkg.JSONOK(c, resp)
}

func (h *Handler) DeleteUser(c *echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, "invalid user ID")
	}

	err = h.svc.Delete(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return pkg.JSONError(c, http.StatusNotFound, pkg.CodeNotFound, "user not found")
		default:
			h.log.Error("delete user failed", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			return pkg.JSONError(c, http.StatusInternalServerError, pkg.CodeInternalServerError, "internal server error")
		}
	}

	return pkg.JSONOK(c, map[string]interface{}{"message": "user deleted successfully"})
}
