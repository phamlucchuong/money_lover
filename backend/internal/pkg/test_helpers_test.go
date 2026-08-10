package pkg

import (
	"net/http/httptest"

	"github.com/labstack/echo/v5"
)

// newRecorder returns a fresh ResponseRecorder. Each test must use its own
// because Echo v5 writes to it directly and reusing would leak state.
func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// newTestContext constructs an *echo.Context suitable for unit-testing the
// pkg helpers (which take *echo.Context) without spinning up a real server.
// Echo v5's Context.JSON uses the embedded *Echo for JSON serialization, so
// we must pass a real *Echo as the third option to NewContext — otherwise
// the serializer is nil and c.JSON panics.
func newTestContext(rec *httptest.ResponseRecorder) *echo.Context {
	req := httptest.NewRequest("GET", "/", nil)
	return echo.NewContext(req, rec, echo.New())
}
