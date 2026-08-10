package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// nopLogger returns a logger that discards output. We don't assert on log
// content in these tests — we just need *Server construction to succeed.
func nopLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// nilDB is acceptable for these tests because the /health handler does not
// touch the database. If a future route depends on gormDB, the constructor
// will panic — that's the signal to introduce a real DB or interface.
var nilDB *gorm.DB

func TestHealthCheck_Returns200(t *testing.T) {
	e := echo.New()
	s := &Server{echo: e, log: nopLogger(), gormDB: nilDB}
	s.SetupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, map[string]string{"status": "healthy"}, body)
}

func TestHealthCheck_UnknownRoute_Returns404(t *testing.T) {
	// Sanity check that SetupRoutes did register /health and didn't
	// accidentally swallow other paths.
	e := echo.New()
	s := &Server{echo: e, log: nopLogger(), gormDB: nilDB}
	s.SetupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHealthCheck_WrongMethod_Returns405(t *testing.T) {
	// /health is registered with GET only. POST must be rejected by Echo's
	// router — this guards against accidentally loosening the method later.
	e := echo.New()
	s := &Server{echo: e, log: nopLogger(), gormDB: nilDB}
	s.SetupRoutes()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
