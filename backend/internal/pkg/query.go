package pkg

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

// ParseIntQuery reads an integer query parameter from the Echo context.
// Returns defaultValue when:
//   - the parameter is missing (empty string)
//   - the value cannot be parsed as an integer
//
// Range validation should be done by the caller after parsing.
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