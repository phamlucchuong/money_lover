package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheck_Returns200 covers the happy path: /health pings the DB,
// the ping succeeds, and the response is a uniform {"status":"healthy"}
// JSON envelope with Content-Type application/json.
func TestHealthCheck_Returns200(t *testing.T) {
	s, mock, _ := newTestServer(t)
	setupHealthExpectations(mock, pingHealthy)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, map[string]string{"status": "healthy"}, body)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestHealthCheck_DBUnreachable_Returns503 covers the failure path: when
// the DB ping errors, /health must surface 503 (not 200 and not crash)
// and include the error message in the response body.
func TestHealthCheck_DBUnreachable_Returns503(t *testing.T) {
	s, mock, _ := newTestServer(t)
	setupHealthExpectations(mock, pingUnhealthy)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "unhealthy", body["status"])
	assert.NotEmpty(t, body["database"], "DB error must be surfaced for ops debugging")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestHealthCheck_UnknownRoute_Returns404 sanity-checks that SetupRoutes
// registered /health and didn't accidentally swallow other paths.
func TestHealthCheck_UnknownRoute_Returns404(t *testing.T) {
	s, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestHealthCheck_WrongMethod_Returns405 guards against accidentally
// loosening the method list on /health later. POST must be rejected by
// Echo's router with 405, not silently routed to GET.
func TestHealthCheck_WrongMethod_Returns405(t *testing.T) {
	s, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}