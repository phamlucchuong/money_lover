package pkg

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

func ParseIntQuery(c *echo.Context, key string, defaultValue int) int {
	val := c.QueryParam(key)
	if val == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return n
}
