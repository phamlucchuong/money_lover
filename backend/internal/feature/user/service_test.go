package user_test

import (
	"context"
	"fmt"
	"testing"

	"chuongpl/quan-ly-chi-tieu/internal/feature/auth"
	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	usermocks "chuongpl/quan-ly-chi-tieu/internal/feature/user/mocks"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newTestService(repo *usermocks.MockRepository) user.Service {
	return user.NewService(repo, nil, nil)
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// 1. Tạo sẵn mock ID cố định
	mockUserID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	validReq := auth.RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	testCases := []struct {
		name          string
		req           auth.RegisterRequest
		mockSetup     func(repo *usermocks.MockRepository)
		expectedResp  *user.UserResponse
		expectedError error
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func(repo *usermocks.MockRepository) {
				repo.EXPECT().GetByEmailAndDeletedAtIsNull(mock.Anything, validReq.Email).
					Return(nil, gorm.ErrRecordNotFound)

				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(validReq.Password))
					return u.Email == validReq.Email &&
						u.Name == validReq.Name &&
						err == nil
				})).
					RunAndReturn(func(ctx context.Context, u *user.User) error {
						u.ID = mockUserID
						return nil
					})

			},
			expectedResp: &user.UserResponse{
				ID:    mockUserID,
				Name:  validReq.Name,
				Email: validReq.Email,
			},
			expectedError: nil,
		},
		{
			name: "unique violation race condition",
			req:  validReq,
			mockSetup: func(repo *usermocks.MockRepository) {
				repo.EXPECT().GetByEmailAndDeletedAtIsNull(mock.Anything, validReq.Email).
					Return(nil, gorm.ErrRecordNotFound)

				pgErr := &pgconn.PgError{Code: "23505"}
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(fmt.Errorf("create user: %w", pgErr))
			},
			expectedResp:  nil,
			expectedError: user.ErrUserAlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(usermocks.MockRepository)

			tc.mockSetup(mockRepo)

			svc := newTestService(mockRepo)

			resp, err := svc.Create(ctx, tc.req)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expectedResp.ID, resp.ID)
				assert.Equal(t, tc.expectedResp.Name, resp.Name)
				assert.Equal(t, tc.expectedResp.Email, resp.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	mockUserID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	testCases := []struct {
		name          string
		req           uuid.UUID
		mockSetup     func(repo *usermocks.MockRepository)
		expectedResp  *user.UserResponse
		expectedError error
	}{
		{
			name: "success",
			req:  mockUserID,
			mockSetup: func(repo *usermocks.MockRepository) {
				repo.EXPECT().GetByID(mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&user.User{
						ID:    mockUserID,
						Name:  "Test User",
						Email: "test@gmail.com",
					}, nil)
			},
			expectedResp: &user.UserResponse{
				ID:    mockUserID,
				Name:  "Test User",
				Email: "test@gmail.com",
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			req:  mockUserID,
			mockSetup: func(repo *usermocks.MockRepository) {
				repo.EXPECT().GetByID(mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectedResp:  nil,
			expectedError: user.ErrUserNotFound,
		},
		{
			name: "internal error",
			req:  mockUserID,
			mockSetup: func(repo *usermocks.MockRepository) {
				repo.EXPECT().GetByID(mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, gorm.ErrInvalidData)
			},
			expectedResp:  nil,
			expectedError: gorm.ErrInvalidData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockrepo := new(usermocks.MockRepository)
			tc.mockSetup(mockrepo)

			svc := user.NewService(mockrepo, nil, nil)
			resp, err := svc.GetByID(ctx, tc.req)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tc.expectedResp.ID, resp.ID)
				assert.Equal(t, tc.expectedResp.Name, resp.Name)
				assert.Equal(t, tc.expectedResp.Email, resp.Email)
			}

			mockrepo.AssertExpectations(t)
		})
	}
}

func TestService_GetAllUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name         string
		page         int
		pageSize     int
		mockSetup    func(repo *usermocks.MockRepository, offset, limit int)
		expectedLen  int
		expectedMeta *pkg.PaginationMeta
		expectedErr  error
	}{
		{
			name:     "page 1 returns first page",
			page:     1,
			pageSize: 20,
			mockSetup: func(repo *usermocks.MockRepository, offset, limit int) {
				repo.EXPECT().GetAll(mock.Anything, offset, limit).
					Return([]*user.User{
						{ID: uuid.New(), Name: "Alice", Email: "alice@example.com"},
						{ID: uuid.New(), Name: "Bob", Email: "bob@example.com"},
					}, int64(2), nil)
			},
			expectedLen: 2,
			expectedMeta: &pkg.PaginationMeta{
				Page:       1,
				PageSize:   20,
				TotalItems: 2,
				TotalPages: 1,
			},
		},
		{
			name:     "page 2 uses offset 20",
			page:     2,
			pageSize: 20,
			mockSetup: func(repo *usermocks.MockRepository, offset, limit int) {
				repo.EXPECT().GetAll(mock.Anything, offset, limit).
					Return([]*user.User{}, int64(45), nil)
			},
			expectedLen: 0,
			expectedMeta: &pkg.PaginationMeta{
				Page:       2,
				PageSize:   20,
				TotalItems: 45,
				TotalPages: 3,
			},
		},
		{
			name:     "empty list returns total_pages 0",
			page:     1,
			pageSize: 20,
			mockSetup: func(repo *usermocks.MockRepository, offset, limit int) {
				repo.EXPECT().GetAll(mock.Anything, offset, limit).
					Return([]*user.User{}, int64(0), nil)
			},
			expectedLen: 0,
			expectedMeta: &pkg.PaginationMeta{
				Page:       1,
				PageSize:   20,
				TotalItems: 0,
				TotalPages: 0,
			},
		},
		{
			name:     "repo error propagated",
			page:     1,
			pageSize: 20,
			mockSetup: func(repo *usermocks.MockRepository, offset, limit int) {
				repo.EXPECT().GetAll(mock.Anything, offset, limit).
					Return(nil, int64(0), gorm.ErrInvalidDB)
			},
			expectedLen: 0,
			expectedErr: gorm.ErrInvalidDB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockrepo := new(usermocks.MockRepository)
			offset := (tc.page - 1) * tc.pageSize
			tc.mockSetup(mockrepo, offset, tc.pageSize)

			svc := user.NewService(mockrepo, nil, nil)
			users, meta, err := svc.GetAllUsers(ctx, tc.page, tc.pageSize)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, users)
				assert.Nil(t, meta)
			} else {
				require.NoError(t, err)
				assert.Len(t, users, tc.expectedLen)
				require.NotNil(t, meta)
				assert.Equal(t, tc.expectedMeta.Page, meta.Page)
				assert.Equal(t, tc.expectedMeta.PageSize, meta.PageSize)
				assert.Equal(t, tc.expectedMeta.TotalItems, meta.TotalItems)
				assert.Equal(t, tc.expectedMeta.TotalPages, meta.TotalPages)
			}

			mockrepo.AssertExpectations(t)
		})
	}
}
