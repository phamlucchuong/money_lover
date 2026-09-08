package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"
	authmocks "chuongpl/quan-ly-chi-tieu/internal/feature/auth/mocks"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// nopLogger returns a logger that discards output so handler tests stay
// readable. We don't assert on log content.
func nopLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func testConfig() *config.Config {
	return &config.Config{
		Environment:            "test",
		AllowedOrigins:         []string{"http://localhost:3000"},
		JWTAccessSecret:        "test-access-secret-do-not-use-in-prod",
		JWTRefreshSecret:       "test-refresh-secret-do-not-use-in-prod",
		AccessTokenExpiration:  5,
		RefreshTokenExpiration: 1440,
	}
}

// realValidator mirrors production server.NewValidator — go-playground/validator
// with struct-tag rules. We need it so handler's c.Validate(&req) calls work.
type realValidator struct{ v *validator.Validate }

func (rv *realValidator) Validate(i any) error { return rv.v.Struct(i) }

// newAuthTestServer wires a minimal Echo with the auth handler routes
// registered. Unlike the production NewServer, this lets us pass a
// pre-mocked auth.Service — the handler test cares about the request/
// response contract, not the service-layer implementation.
//
// The blacklist checker always returns (false, nil) so the JWT middleware
// lets requests through without consulting Redis.
func newAuthTestServer(t *testing.T, svc auth.Service) *echo.Echo {
	t.Helper()

	e := echo.New()
	e.Validator = &realValidator{v: validator.New()}

	cfg := testConfig()
	h := auth.NewHandler(svc, cfg, nopLogger())

	noBlacklist := auth.BlacklistChecker(func(_ context.Context, _ string) (bool, error) {
		return false, nil
	})

	v1 := e.Group("/api/v1")
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.POST("/refresh", h.RefreshToken)
	authGroup.POST("/logout", h.Logout, auth.JWTAuth(cfg, nopLogger(), noBlacklist))

	return e
}

// doJSON marshals body and sends a request through the Echo instance,
// returning the recorder for assertions.
func doJSON(t *testing.T, e *echo.Echo, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestHandler_Register_Success asserts the happy path: a valid payload
// produces a 201 Created with the user response body.
func TestHandler_Register_Success(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	expectedID := uuid.New()
	svcMock.EXPECT().
		Register(mock.Anything, mock.AnythingOfType("auth.RegisterRequest")).
		Return(&user.UserResponse{ID: expectedID, Name: "Alice", Email: "alice@example.com"}, nil)

	e := newAuthTestServer(t, svcMock)

	rec := doJSON(t, e, http.MethodPost, "/api/v1/auth/register", auth.RegisterRequest{
		Name: "Alice", Email: "alice@example.com", Password: "hunter2hunter",
	})

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data user.UserResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, expectedID, resp.Data.ID)
	assert.Equal(t, "Alice", resp.Data.Name)
	assert.Equal(t, "alice@example.com", resp.Data.Email)
}

// TestHandler_Register_Duplicate covers the already-exists path: the
// service returns ErrUserAlreadyExists, handler must surface 409.
func TestHandler_Register_Duplicate(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	svcMock.EXPECT().
		Register(mock.Anything, mock.AnythingOfType("auth.RegisterRequest")).
		Return(nil, user.ErrUserAlreadyExists)

	e := newAuthTestServer(t, svcMock)

	rec := doJSON(t, e, http.MethodPost, "/api/v1/auth/register", auth.RegisterRequest{
		Name: "Bob", Email: "bob@example.com", Password: "hunter2hunter",
	})

	require.Equal(t, http.StatusConflict, rec.Code)
}

// TestHandler_Register_InvalidBody asserts malformed JSON short-circuits
// at binding with 400 Bad Request before reaching the service.
func TestHandler_Register_InvalidBody(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	// No EXPECT() — service should never be called when binding fails.

	e := newAuthTestServer(t, svcMock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("not json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandler_Register_ValidationFailure asserts invalid input (password
// too short) trips go-playground/validator with 422 before the service
// is consulted.
func TestHandler_Register_ValidationFailure(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	// No EXPECT() — validation should reject before the service runs.

	e := newAuthTestServer(t, svcMock)

	rec := doJSON(t, e, http.MethodPost, "/api/v1/auth/register", auth.RegisterRequest{
		Name: "Alice", Email: "alice@example.com", Password: "short", // min=8
	})

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// TestHandler_Login_Success covers the happy path: a valid login sets the
// auth cookies and returns 200 with the AuthResponse body.
func TestHandler_Login_Success(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	svcMock.EXPECT().
		Login(mock.Anything, mock.AnythingOfType("auth.LoginRequest")).
		Return(&auth.AuthResponse{
			Authenticated: true,
			AccessToken:   "access.jwt.token",
			RefreshToken:  "refresh.jwt.token",
		}, nil)

	e := newAuthTestServer(t, svcMock)

	rec := doJSON(t, e, http.MethodPost, "/api/v1/auth/login", auth.LoginRequest{
		Email: "alice@example.com", Password: "hunter2hunter",
	})

	require.Equal(t, http.StatusOK, rec.Code)

	// Both cookies should be set with HttpOnly + Secure + SameSite=Strict.
	cookies := rec.Result().Cookies()
	var accessCookie, refreshCookie *http.Cookie
	for _, c := range cookies {
		switch c.Name {
		case "access_token":
			accessCookie = c
		case "refresh_token":
			refreshCookie = c
		}
	}
	require.NotNil(t, accessCookie, "access_token cookie must be set")
	require.NotNil(t, refreshCookie, "refresh_token cookie must be set")
	assert.Equal(t, "access.jwt.token", accessCookie.Value)
	assert.Equal(t, "refresh.jwt.token", refreshCookie.Value)
	assert.True(t, accessCookie.HttpOnly)
	assert.True(t, accessCookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, accessCookie.SameSite)
	assert.True(t, accessCookie.Expires.After(time.Now()), "access cookie should expire in the future")
}

// TestHandler_Login_BadCredentials asserts that any login failure surfaces
// as a uniform 401 with a generic message — the handler must not leak
// whether the email exists.
func TestHandler_Login_BadCredentials(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	svcMock.EXPECT().
		Login(mock.Anything, mock.AnythingOfType("auth.LoginRequest")).
		Return(nil, errors.New("invalid credentials"))

	e := newAuthTestServer(t, svcMock)

	rec := doJSON(t, e, http.MethodPost, "/api/v1/auth/login", auth.LoginRequest{
		Email: "alice@example.com", Password: "wrong",
	})

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestHandler_Refresh_MissingCookie asserts the no-cookie path: when the
// client sends no refresh_token cookie, the handler returns 401 without
// consulting the service.
func TestHandler_Refresh_MissingCookie(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	// No EXPECT() — service should not be invoked.

	e := newAuthTestServer(t, svcMock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestHandler_Refresh_InvalidToken asserts the bad-cookie path: cookie
// is present but the service rejects the token with a generic error,
// and the handler responds 401.
func TestHandler_Refresh_InvalidToken(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	svcMock.EXPECT().
		RefreshToken(mock.Anything, "bad.token").
		Return(nil, errors.New("invalid refresh token"))

	e := newAuthTestServer(t, svcMock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad.token"})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestHandler_Logout_RequiresJWT asserts that /logout without a token
// short-circuits at the JWT middleware with 401 — the logout handler
// must never run for anonymous callers.
func TestHandler_Logout_RequiresJWT(t *testing.T) {
	svcMock := authmocks.NewMockService(t)
	// No EXPECT() — service must not run because the middleware blocks.

	e := newAuthTestServer(t, svcMock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}