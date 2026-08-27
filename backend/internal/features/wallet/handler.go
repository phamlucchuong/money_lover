package wallet

import (
	"log/slog"
	"net/http"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc service
	cfg *config.Config
	log *slog.Logger
}

func NewHandler(svc service, cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		cfg: cfg,
		log: log,
	}
}

/*
this func have not implement yet,
because we need to implement the auth service first,
once the auth service is implemented, we can get userID from
request's access token, then store it in the context
*/
func (h *Handler) CreateWallet(c *echo.Context) error {
	var req CreateWalletRequest
	if err := c.Bind(&req); err != nil {
		return pkg.JSONError(c, http.StatusBadRequest, pkg.CodeBadRequest, pkg.ErrInvalidInput.Error())
	}

	if err := c.Validate(&req); err != nil {
		return pkg.JSONError(c, http.StatusUnprocessableEntity, pkg.CodeValidationError, pkg.ErrInvalidInput.Error())
	}

	return nil
}
