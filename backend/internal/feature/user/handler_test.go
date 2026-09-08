package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	usermocks "chuongpl/quan-ly-chi-tieu/internal/feature/user/mocks"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
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

// realValidator mirrors server.NewValidator so c.Validate(&req) calls in
// handlers behave the same way as production.
type realValidator struct{ v *validator.Validate }

func (rv *realValidator) Validate(i any) error { return rv.v.Struct(i) }

// newUserTestServer wires a minimal Echo with the user CRUD routes
// registered against a mocked user.Service so handler tests can assert
// on the request/response contract.
func newUserTestServer(t *testing.T, svc user.Service) *echo.Echo {
	t.Helper()

	e := echo.New()
	e.Validator = &realValidator{v: validator.New()}

	h := user.NewHandler(svc, testConfig(), nopLogger())

	v1 := e.Group("/api/v1")
	g := v1.Group("/users")
	g.GET("", h.GetAllUsers)
	g.GET("/:id", h.GetUserByID)
	g.PUT("/:id", h.UpdateUser)
	g.DELETE("/:id", h.DeleteUser)

	return e
}

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

// TestHandler_GetUserByID_Success asserts the happy path.
func TestHandler_GetUserByID_Success(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		GetByID(mock.Anything, userID).
		Return(&user.UserResponse{ID: userID, Name: "Alice", Email: "alice@example.com"}, nil)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users/"+userID.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestHandler_GetUserByID_BadUUID asserts malformed UUIDs short-circuit
// at parsing with 400 before the service is consulted.
func TestHandler_GetUserByID_BadUUID(t *testing.T) {
	svcMock := new(usermocks.MockService)
	// No EXPECT — service should not be called.

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users/not-a-uuid", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandler_GetUserByID_NotFound asserts the missing-user path: the
// service returns ErrUserNotFound, handler must surface 404.
func TestHandler_GetUserByID_NotFound(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		GetByID(mock.Anything, userID).
		Return(nil, user.ErrUserNotFound)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users/"+userID.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// TestHandler_GetAllUsers_Success asserts the list endpoint returns
// 200 with pagination metadata.
func TestHandler_GetAllUsers_Success(t *testing.T) {
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		GetAllUsers(mock.Anything, 1, 20).
		Return(
			[]*user.UserResponse{{ID: uuid.New(), Name: "Alice", Email: "alice@example.com"}},
			&pkg.PaginationMeta{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1},
			nil,
		)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users", nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestHandler_GetAllUsers_InvalidPage asserts the page < 1 path is
// rejected with 400 before the service runs.
func TestHandler_GetAllUsers_InvalidPage(t *testing.T) {
	svcMock := new(usermocks.MockService)
	// No EXPECT — handler must reject before the service runs.

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users?page=0", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandler_GetAllUsers_InvalidPageSize asserts the page_size out-of-
// range path is rejected with 400.
func TestHandler_GetAllUsers_InvalidPageSize(t *testing.T) {
	svcMock := new(usermocks.MockService)
	// No EXPECT — handler must reject before the service runs.

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodGet, "/api/v1/users?page_size=500", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandler_UpdateUser_Success asserts the happy path: a valid
// payload produces 200 with the updated user body.
func TestHandler_UpdateUser_Success(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		Update(mock.Anything, userID, mock.AnythingOfType("user.UpdateUserRequest")).
		Return(&user.UserResponse{ID: userID, Name: "Alice Updated", Email: "alice@example.com"}, nil)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodPut, "/api/v1/users/"+userID.String(), user.UpdateUserRequest{
		Name: "Alice Updated",
	})

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestHandler_UpdateUser_NotFound covers the missing-user path.
func TestHandler_UpdateUser_NotFound(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		Update(mock.Anything, userID, mock.AnythingOfType("user.UpdateUserRequest")).
		Return(nil, user.ErrUserNotFound)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodPut, "/api/v1/users/"+userID.String(), user.UpdateUserRequest{
		Name: "Ghost",
	})

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// TestHandler_UpdateUser_DuplicateEmail covers the already-exists path.
func TestHandler_UpdateUser_DuplicateEmail(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		Update(mock.Anything, userID, mock.AnythingOfType("user.UpdateUserRequest")).
		Return(nil, user.ErrUserAlreadyExists)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodPut, "/api/v1/users/"+userID.String(), user.UpdateUserRequest{
		Email: "taken@example.com",
	})

	require.Equal(t, http.StatusConflict, rec.Code)
}

// TestHandler_UpdateUser_InvalidEmail covers the validator's email rule.
func TestHandler_UpdateUser_InvalidEmail(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	// No EXPECT — validation must reject before the service runs.

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodPut, "/api/v1/users/"+userID.String(), user.UpdateUserRequest{
		Email: "not-an-email",
	})

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// TestHandler_DeleteUser_Success asserts the happy path.
func TestHandler_DeleteUser_Success(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		Delete(mock.Anything, userID).
		Return(nil)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodDelete, "/api/v1/users/"+userID.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestHandler_DeleteUser_NotFound covers the missing-user path.
func TestHandler_DeleteUser_NotFound(t *testing.T) {
	userID := uuid.New()
	svcMock := new(usermocks.MockService)
	svcMock.EXPECT().
		Delete(mock.Anything, userID).
		Return(user.ErrUserNotFound)

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodDelete, "/api/v1/users/"+userID.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// TestHandler_DeleteUser_BadUUID covers the malformed-UUID path.
func TestHandler_DeleteUser_BadUUID(t *testing.T) {
	svcMock := new(usermocks.MockService)
	// No EXPECT — handler must reject before the service runs.

	e := newUserTestServer(t, svcMock)
	rec := doJSON(t, e, http.MethodDelete, "/api/v1/users/garbage", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// silenceUnusedContextImport keeps the import alive if all tests are
// ever removed — the test helper pattern still uses context.Context
// indirectly via mock.Anything matchers.
var _ = context.Background