package server

import (
	"io"
	"log/slog"
	"testing"

	"chuongpl/quan-ly-chi-tieu/internal/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// nopLogger returns a logger that discards output. Tests don't assert on log
// content — they just need *Server construction to succeed.
func nopLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// newTestConfig returns a *config.Config with deterministic JWT secrets so
// token issuance in tests is reproducible. Other fields are zero-valued,
// which is fine because the handlers under test don't read them.
func newTestConfig() *config.Config {
	return &config.Config{
		Environment:            "test",
		AllowedOrigins:         []string{"http://localhost:3000"},
		JWTAccessSecret:        "test-access-secret-do-not-use-in-prod",
		JWTRefreshSecret:       "test-refresh-secret-do-not-use-in-prod",
		AccessTokenExpiration:  5,    // minutes
		RefreshTokenExpiration: 1440, // minutes (24h)
	}
}

// newMockDB returns a *gorm.DB backed by a sqlmock-driven *sql.DB plus the
// sqlmock controller so callers can stage expected SQL. Use this anywhere a
// handler touches the database — including the /health ping.
//
// The DSN passed to the postgres driver is empty because we inject the
// underlying *sql.DB directly via postgres.Config{Conn: sqlDB}, which skips
// driver-level connection entirely.
//
// MonitorPingsOption is required: GORM pings the underlying *sql.DB on
// every connection acquisition when MaxIdleConns > 0, and without
// monitoring, sqlmock silently drops ExpectPing and the /health failure
// path can't be exercised.
func newMockDB(t testing.TB) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	// gorm.Open pings the underlying *sql.DB once on init to verify the
	// connection — stage that expectation here so callers only have to
	// add their own expectations for the test scenario.
	mock.ExpectPing()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return gormDB, mock
}

// newMockRedis returns a *redis.Client pointed at an in-process miniredis
// instance, plus the miniredis handle so callers can read/manipulate keys
// directly when needed (e.g. to seed the JWT blacklist).
func newMockRedis(t testing.TB) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	return rdb, mr
}

// newTestServer wires a Server using sqlmock + miniredis, returning the
// server, the sqlmock controller, and the miniredis handle. Use this in
// handler tests instead of constructing *Server by hand so the real
// NewServer wiring (including auth/user service setup) runs.
//
// newMockDB already stages the gorm.Open ping expectation, so /health
// handlers that ping the DB just need to stage one more ExpectPing.
func newTestServer(t testing.TB) (*Server, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()

	gormDB, mock := newMockDB(t)
	rdb, mr := newMockRedis(t)

	srv := NewServer(newTestConfig(), nopLogger(), gormDB, rdb)
	return srv, mock, mr
}

// expectPingSuccess stages the sqlmock expectation for a healthy /health
// response: a single successful ping, no rows returned.
func expectPingSuccess(mock sqlmock.Sqlmock) {
	mock.ExpectPing()
}

// expectPingFailure stages the sqlmock expectation for a /health response
// that should surface a 503: a ping that returns a generic driver error.
func expectPingFailure(mock sqlmock.Sqlmock) {
	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)
}

// pingExpectations captures which /health scenario we want to set up.
// Avoids boolean-flag overload at call sites.
type pingExpectations int

const (
	pingHealthy pingExpectations = iota
	pingUnhealthy
)

// setupHealthExpectations is a tiny convenience wrapper around the helpers
// above — call it once at the top of each /health test.
func setupHealthExpectations(mock sqlmock.Sqlmock, kind pingExpectations) {
	switch kind {
	case pingHealthy:
		expectPingSuccess(mock)
	case pingUnhealthy:
		expectPingFailure(mock)
	default:
		panic("unknown pingExpectations")
	}
}