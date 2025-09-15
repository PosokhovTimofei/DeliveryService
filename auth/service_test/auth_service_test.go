package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/auth/models"
	"github.com/maksroxx/DeliveryService/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userID string, updateFields map[string]any) error {
	args := m.Called(ctx, userID, updateFields)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockTelegramer struct {
	mock.Mock
}

func (m *MockTelegramer) Save(code string, userID string, ttl time.Duration) error {
	args := m.Called(code, userID, ttl)
	return args.Error(0)
}

func (m *MockTelegramer) FindUserIDByCode(code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTgRepo := new(MockTelegramer)
	authService := service.NewAuthService(mockUserRepo, mockTgRepo, "test-secret")

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				// Исправлено: возвращаем nil для пользователя и ошибку
				mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("user not found"))
				mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func() {
				existingUser := &models.User{ID: "existing-user", Email: "existing@example.com"}
				mockUserRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: models.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, token, err := authService.Register(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, "user", user.Role)
				err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(tt.password))
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RegisterModerator(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTgRepo := new(MockTelegramer)
	authService := service.NewAuthService(mockUserRepo, mockTgRepo, "test-secret")

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "successful moderator registration",
			email:    "moderator@example.com",
			password: "password123",
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, "moderator@example.com").Return(nil, errors.New("user not found"))
				mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, token, err := authService.RegisterModerator(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, "moderator", user.Role)
				err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(tt.password))
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTgRepo := new(MockTelegramer)
	authService := service.NewAuthService(mockUserRepo, mockTgRepo, "test-secret")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:                "test-user",
		Email:             "test@example.com",
		EncryptedPassword: string(hashedPassword),
		Role:              "user",
	}

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "correctpassword",
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
			},
			expectedError: nil,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
			expectedError: models.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
			},
			expectedError: models.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, token, err := authService.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.email, user.Email)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
