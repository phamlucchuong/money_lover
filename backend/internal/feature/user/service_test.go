package user_test

import (
	"context"
	"testing"

	"chuongpl/quan-ly-chi-tieu/internal/feature/user"
	usermocks "chuongpl/quan-ly-chi-tieu/internal/feature/user/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	validReq := user.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	testCases := []struct {
		name          string
		req           user.CreateUserRequest
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
					return u.Email == validReq.Email &&
						u.Name == validReq.Name &&
						u.Password == validReq.Password
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
