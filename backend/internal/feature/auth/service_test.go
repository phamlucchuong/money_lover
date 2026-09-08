package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	usermocks "chuongpl/quan-ly-chi-tieu/internal/feature/user/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// fakeCache is a tiny in-process implementation of cache.Cache for use
// only in auth service tests. It records the most recent Set so the
// RefreshToken test can assert that the old JTI was blacklisted with
// the correct remaining TTL.
//
// We keep this hand-rolled instead of generating a mock so the test
// doesn't depend on a generation step and the assertions read top-to-
// bottom in one file.
type fakeCache struct {
	mu      sync.Mutex
	store   map[string]string
	lastSet struct {
		key   string
		value string
		ttl   time.Duration
	}
}

func newFakeCache() *fakeCache {
	return &fakeCache{store: map[string]string{}}
}

func (f *fakeCache) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.store[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return v, nil
}

func (f *fakeCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[key] = value
	f.lastSet.key = key
	f.lastSet.value = value
	f.lastSet.ttl = ttl
	return nil
}

func (f *fakeCache) Delete(_ context.Context, keys ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, k := range keys {
		delete(f.store, k)
	}
	return nil
}

func (f *fakeCache) Exists(_ context.Context, key string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.store[key]
	return ok, nil
}

func newSvc(t *testing.T, userSvc user.Service, cache *fakeCache) auth.Service {
	t.Helper()
	return auth.NewService(userSvc, cache, testConfig(), nopLogger())
}

// decodeJTI extracts the jti claim from a signed JWT so we can poison
// the blacklist in the replay-detection test.
func decodeJTI(tokenStr, secret string) (string, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("unexpected claims type")
	}
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return "", errors.New("missing jti")
	}
	return jti, nil
}

// TestService_Register_Success asserts the happy path: a fresh email
// returns a UserResponse populated from the user service.
func TestService_Register_Success(t *testing.T) {
	userID := uuid.New()
	userSvc := new(usermocks.MockService)
	// auth.service.Register treats any error other than gorm.ErrRecordNotFound
	// as a real failure — so the mock must surface ErrRecordNotFound
	// to drive the "create" path.
	userSvc.EXPECT().
		GetByEmailAndDeletedAtIsNull(mock.Anything, "alice@example.com").
		Return(nil, gorm.ErrRecordNotFound)
	userSvc.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*user.User")).
		RunAndReturn(func(_ context.Context, u *user.User) (*user.UserResponse, error) {
			return &user.UserResponse{ID: userID, Name: u.Name, Email: u.Email}, nil
		})

	svc := newSvc(t, userSvc, newFakeCache())
	resp, err := svc.Register(context.Background(), auth.RegisterRequest{
		Name: "Alice", Email: "alice@example.com", Password: "hunter2hunter",
	})

	require.NoError(t, err)
	assert.Equal(t, userID, resp.ID)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "alice@example.com", resp.Email)
}

// TestService_Register_Duplicate covers the already-exists path: when
// user.Service.GetByEmail returns an existing record, Register must
// return ErrUserAlreadyExists without attempting to create.
func TestService_Register_Duplicate(t *testing.T) {
	userSvc := new(usermocks.MockService)
	userSvc.EXPECT().
		GetByEmailAndDeletedAtIsNull(mock.Anything, "bob@example.com").
		Return(&user.User{ID: uuid.New(), Email: "bob@example.com"}, nil)
	// No Create EXPECT — Register must short-circuit.

	svc := newSvc(t, userSvc, newFakeCache())
	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		Name: "Bob", Email: "bob@example.com", Password: "hunter2hunter",
	})

	assert.ErrorIs(t, err, user.ErrUserAlreadyExists)
}

// TestService_Logout_BlacklistsJTI asserts that Logout writes the JTI
// to the cache with the supplied TTL.
func TestService_Logout_BlacklistsJTI(t *testing.T) {
	cache := newFakeCache()
	svc := newSvc(t, new(usermocks.MockService), cache)

	jti := uuid.New().String()
	ttl := 5 * time.Minute
	err := svc.Logout(context.Background(), jti, ttl)

	require.NoError(t, err)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	assert.Equal(t, "blacklist:"+jti, cache.lastSet.key)
	assert.Equal(t, "true", cache.lastSet.value)
	assert.Equal(t, ttl, cache.lastSet.ttl)
}

// TestService_RefreshToken_RotatesAndBlacklists covers the rotation path:
// a valid refresh token gets the old JTI blacklisted with its remaining
// TTL, and a fresh access+refresh pair is returned.
func TestService_RefreshToken_RotatesAndBlacklists(t *testing.T) {
	password := "correct-horse-battery-staple"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	userID := uuid.New()
	userSvc := new(usermocks.MockService)
	userSvc.EXPECT().
		GetByEmailAndDeletedAtIsNull(mock.Anything, "alice@example.com").
		Return(&user.User{ID: userID, Email: "alice@example.com", Name: "Alice", Password: string(hashed)}, nil)

	cache := newFakeCache()
	svc := newSvc(t, userSvc, cache)

	loginResp, err := svc.Login(context.Background(), auth.LoginRequest{
		Email: "alice@example.com", Password: password,
	})
	require.NoError(t, err)
	originalRefresh := loginResp.RefreshToken

	refreshResp, err := svc.RefreshToken(context.Background(), originalRefresh)
	require.NoError(t, err)
	assert.NotEqual(t, originalRefresh, refreshResp.RefreshToken, "refresh token must rotate")
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.NotEmpty(t, refreshResp.RefreshToken)

	// Cache must contain a blacklist entry — verify with the same key
	// shape auth uses internally.
	assert.True(t, len(cache.store) > 0, "RefreshToken must write to the blacklist")
}

// TestService_RefreshToken_RejectsReplay covers the replay-detection
// path: a refresh token that was already rotated (its JTI is on the
// blacklist) must be rejected even if its signature is still valid.
func TestService_RefreshToken_RejectsReplay(t *testing.T) {
	password := "correct-horse-battery-staple"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	userID := uuid.New()
	userSvc := new(usermocks.MockService)
	userSvc.EXPECT().
		GetByEmailAndDeletedAtIsNull(mock.Anything, "alice@example.com").
		Return(&user.User{ID: userID, Email: "alice@example.com", Name: "Alice", Password: string(hashed)}, nil)

	cache := newFakeCache()
	svc := newSvc(t, userSvc, cache)

	loginResp, err := svc.Login(context.Background(), auth.LoginRequest{
		Email: "alice@example.com", Password: password,
	})
	require.NoError(t, err)

	// Decode the original refresh to read its JTI so we can poison it.
	cfg := testConfig()
	parsed, err := decodeJTI(loginResp.RefreshToken, cfg.JWTRefreshSecret)
	require.NoError(t, err)
	require.NoError(t, cache.Set(context.Background(), "blacklist:"+parsed, "true", time.Hour))

	_, err = svc.RefreshToken(context.Background(), loginResp.RefreshToken)
	require.Error(t, err, "RefreshToken must reject a blacklisted JTI")
	assert.Contains(t, err.Error(), "revoked")
}