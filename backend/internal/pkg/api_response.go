package pkg

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type APIResponse struct {
	Data      interface{}     `json:"data,omitempty"`
	Error     *APIError       `json:"error,omitempty"`
	Meta      *PaginationMeta `json:"meta,omitempty"`
	TimeStamp string          `json:"timestamp"`
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func JSONOK(c *echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Data:      data,
		TimeStamp: now(),
	})
}

func JSONList(c *echo.Context, data interface{}, meta PaginationMeta) error {
	return c.JSON(http.StatusOK, APIResponse{
		Data:      data,
		Meta:      &meta,
		TimeStamp: now(),
	})
}

func JSONCreated(c *echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Data:      data,
		TimeStamp: now(),
	})
}

func JSONError(c *echo.Context, status int, code int, message string) error {
	return c.JSON(status, APIResponse{
		Error: &APIError{
			Code:    code,
			Message: message,
			Status:  status,
		},
		TimeStamp: now(),
	})
}
