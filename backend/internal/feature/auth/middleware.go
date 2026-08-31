package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// Context keys set by JWTAuth middleware.
const (
	CtxKeyUserID = "user_id"
	CtxKeyJTI    = "jti"
	CtxKeyExp    = "exp"
)

// BlacklistChecker reports whether a JWT JTI has been revoked.
type BlacklistChecker func(ctx context.Context, jti string) (bool, error)

// extractToken reads JWT from access_token cookie first, then Authorization header (mobile fallback).
func extractToken(c *echo.Context) string {
	if cookie, err := c.Cookie("access_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

// JWTAuth returns an Echo middleware that validates Bearer JWTs and checks the blacklist.
func JWTAuth(cfg *config.Config, log *slog.Logger, checker BlacklistChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw := extractToken(c)
			if raw == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication token")
			}

			token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
				}
				return []byte(cfg.JWTAccessSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			userID, _ := claims["sub"].(string)
			jti, _ := claims["jti"].(string)

			if userID == "" || jti == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			// Check blacklist.
			blacklisted, err := checker(c.Request().Context(), jti)
			if err != nil {
				log.Error("blacklist check failed", "jti", jti, "err", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
			}
			if blacklisted {
				return echo.NewHTTPError(http.StatusUnauthorized, "token has been revoked")
			}

			// Parse exp for downstream use (Logout needs it).
			var expTime time.Time
			if expVal, ok := claims["exp"].(float64); ok {
				expTime = time.Unix(int64(expVal), 0)
			}

			c.Set(CtxKeyUserID, userID)
			c.Set(CtxKeyJTI, jti)
			c.Set(CtxKeyExp, expTime)

			return next(c)
		}
	}
}

// OptionalJWTAuth returns an Echo middleware that extracts user info from a
// Bearer JWT if present, but allows unauthenticated requests to proceed.
// Sets CtxKeyUserID in context only if a valid token is found.
func OptionalJWTAuth(cfg *config.Config, checker BlacklistChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw := extractToken(c)
			if raw == "" {
				return next(c)
			}

			token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(cfg.JWTAccessSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !token.Valid {
				// Invalid token — proceed as anonymous
				return next(c)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return next(c)
			}

			userID, _ := claims["sub"].(string)
			jti, _ := claims["jti"].(string)

			if userID == "" || jti == "" {
				return next(c)
			}

			// Check blacklist — if error or blacklisted, just proceed as anonymous.
			blacklisted, err := checker(c.Request().Context(), jti)
			if err != nil || blacklisted {
				return next(c)
			}

			c.Set(CtxKeyUserID, userID)

			return next(c)
		}
	}
}
