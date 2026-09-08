package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signToken mints a HS256 access token with the given claims. Used by
// middleware tests to drive the JWTAuth path.
func signToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

// protectedHandler returns a no-op echo.HandlerFunc that records
// whether the middleware let it through.
func protectedHandler(reached *bool) echo.HandlerFunc {
	return func(c *echo.Context) error {
		*reached = true
		return c.NoContent(http.StatusOK)
	}
}

// TestJWTAuth_MissingToken asserts that without any token (no cookie,
// no Authorization header) the middleware blocks the request with
// 401 — the protected handler must never run.
func TestJWTAuth_MissingToken(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.JWTAuth(cfg, nopLogger(), func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, reached, "protected handler must not run without a token")
}

// TestJWTAuth_ValidBearerHeader asserts the Authorization: Bearer
// fallback path: when there's no cookie, a Bearer header is honoured.
func TestJWTAuth_ValidBearerHeader(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.JWTAuth(cfg, nopLogger(), func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	uid := uuid.New().String()
	exp := time.Now().Add(time.Hour)
	tok := signToken(t, cfg.JWTAccessSecret, jwt.MapClaims{
		"sub": uid, "jti": uuid.New().String(), "exp": exp.Unix(),
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, reached)
}

// TestJWTAuth_BlacklistedJTI asserts the blacklist path: a token with
// a valid signature but a JTI that the checker reports as revoked
// must be rejected with 401, even though the JWT itself is valid.
func TestJWTAuth_BlacklistedJTI(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.JWTAuth(cfg, nopLogger(), func(_ context.Context, _ string) (bool, error) {
		return true, nil // JTI is revoked
	})
	e.GET("/protected", protectedHandler(&reached), mw)

	tok := signToken(t, cfg.JWTAccessSecret, jwt.MapClaims{
		"sub": uuid.New().String(), "jti": uuid.New().String(), "exp": time.Now().Add(time.Hour).Unix(),
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, reached)
}

// TestJWTAuth_ExpiredToken asserts the expired-signature path: a token
// whose exp claim is in the past is rejected with 401.
func TestJWTAuth_ExpiredToken(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.JWTAuth(cfg, nopLogger(), func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	tok := signToken(t, cfg.JWTAccessSecret, jwt.MapClaims{
		"sub": uuid.New().String(), "jti": uuid.New().String(), "exp": time.Now().Add(-time.Hour).Unix(),
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, reached)
}

// TestJWTAuth_BadSignature asserts the wrong-secret path: a token signed
// with a different secret is rejected with 401 — secret isolation between
// access and refresh tokens is part of the security model.
func TestJWTAuth_BadSignature(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.JWTAuth(cfg, nopLogger(), func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	tok := signToken(t, "completely-different-secret", jwt.MapClaims{
		"sub": uuid.New().String(), "jti": uuid.New().String(), "exp": time.Now().Add(time.Hour).Unix(),
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, reached)
}

// TestOptionalJWTAuth_MissingTokenAllowsAnonymous asserts the optional
// middleware's defining behaviour: no token, no problem — the handler
// still runs.
func TestOptionalJWTAuth_MissingTokenAllowsAnonymous(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.OptionalJWTAuth(cfg, func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, reached, "OptionalJWTAuth must allow anonymous requests through")
}

// TestOptionalJWTAuth_InvalidTokenFallsBackToAnonymous asserts that a
// present-but-invalid token does NOT abort the request — the user is
// treated as anonymous instead. This is intentional: optional auth is
// for endpoints that adapt to logged-in users without requiring one.
func TestOptionalJWTAuth_InvalidTokenFallsBackToAnonymous(t *testing.T) {
	var reached bool
	cfg := testConfig()
	e := echo.New()
	mw := auth.OptionalJWTAuth(cfg, func(_ context.Context, _ string) (bool, error) { return false, nil })
	e.GET("/protected", protectedHandler(&reached), mw)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer not.a.valid.jwt")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, reached)
}