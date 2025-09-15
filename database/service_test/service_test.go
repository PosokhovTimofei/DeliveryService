package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/service"
	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRouteRepository struct {
	mock.Mock
}

func (m *MockRouteRepository) GetByID(ctx context.Context, id string) (*models.Package, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockRouteRepository) GetAllPackages(ctx context.Context, filter models.PackageFilter) ([]*models.Package, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Package), args.Error(1)
}

func (m *MockRouteRepository) GetExpiredPackages(ctx context.Context) ([]*models.Package, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Package), args.Error(1)
}

func (m *MockRouteRepository) MarkAsExpiredByID(ctx context.Context, packageID string) (*models.Package, error) {
	args := m.Called(ctx, packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockRouteRepository) Create(ctx context.Context, route *models.Package) (*models.Package, error) {
	args := m.Called(ctx, route)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockRouteRepository) UpdatePackage(ctx context.Context, id string, update models.PackageUpdate) (*models.Package, error) {
	args := m.Called(ctx, id, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockRouteRepository) DeletePackage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRouteRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockCalculator struct {
	mock.Mock
}

func (m *MockCalculator) Calculate(weight float64, userID, from, to, address string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	args := m.Called(weight, userID, from, to, address, length, width, height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*calculatorpb.CalculateDeliveryCostResponse), args.Error(1)
}

func (m *MockCalculator) CalculateByTariff(weight float64, userID, from, to, address, tariffCode string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	args := m.Called(weight, userID, from, to, address, tariffCode, length, width, height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*calculatorpb.CalculateDeliveryCostResponse), args.Error(1)
}

type MockPaymentProducer struct {
	mock.Mock
}

func (m *MockPaymentProducer) SendPaymentEvent(payment models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentProducer) SendExpiredPackageEvent(pkg models.Package) error {
	args := m.Called(pkg)
	return args.Error(0)
}

func TestPackageService_GetPackageByID(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockCalc := new(MockCalculator)
	mockProducer := new(MockPaymentProducer)
	logger := logrus.New()

	packageService := service.NewPackageService(mockRepo, mockCalc, mockProducer, logger)

	tests := []struct {
		name           string
		packageID      string
		setupMocks     func()
		expectedError  error
		expectedResult *models.Package
	}{
		{
			name:      "successful get",
			packageID: "test-package-1",
			setupMocks: func() {
				testPackage := &models.Package{PackageID: "test-package-1", UserID: "test-user"}
				mockRepo.On("GetByID", mock.Anything, "test-package-1").Return(testPackage, nil)
			},
			expectedError:  nil,
			expectedResult: &models.Package{PackageID: "test-package-1", UserID: "test-user"},
		},
		{
			name:      "package not found",
			packageID: "non-existent-package",
			setupMocks: func() {
				mockRepo.On("GetByID", mock.Anything, "non-existent-package").Return(nil, errors.New("not found"))
			},
			expectedError:  errors.New("not found"),
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			result, err := packageService.GetPackageByID(context.Background(), tt.packageID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.PackageID, result.PackageID)
				assert.Equal(t, tt.expectedResult.UserID, result.UserID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPackageService_CreatePackage(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockCalc := new(MockCalculator)
	mockProducer := new(MockPaymentProducer)
	logger := logrus.New()

	packageService := service.NewPackageService(mockRepo, mockCalc, mockProducer, logger)

	testPackage := &models.Package{
		PackageID: "test-package-1",
		UserID:    "test-user",
		Weight:    10.0,
		From:      "New York",
		To:        "Los Angeles",
	}

	tests := []struct {
		name          string
		pkg           *models.Package
		setupMocks    func()
		expectedError error
	}{
		{
			name: "create error",
			pkg:  testPackage,
			setupMocks: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Package")).Return(nil, errors.New("create failed"))
			},
			expectedError: errors.New("create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			result, err := packageService.CreatePackage(context.Background(), tt.pkg)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.pkg.PackageID, result.PackageID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPackageService_CancelPackage(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockCalc := new(MockCalculator)
	mockProducer := new(MockPaymentProducer)
	logger := logrus.New()

	packageService := service.NewPackageService(mockRepo, mockCalc, mockProducer, logger)

	tests := []struct {
		name          string
		packageID     string
		setupMocks    func()
		expectedError error
	}{
		{
			name:      "successful cancel",
			packageID: "test-package-1",
			setupMocks: func() {
				testPackage := &models.Package{
					PackageID: "test-package-1",
					UserID:    "test-user",
					Status:    "Created",
				}
				mockRepo.On("GetByID", mock.Anything, "test-package-1").Return(testPackage, nil)

				updatedPackage := &models.Package{
					PackageID: "test-package-1",
					UserID:    "test-user",
					Status:    "Сanceled",
				}
				update := models.PackageUpdate{Status: "Сanceled"}
				mockRepo.On("UpdatePackage", mock.Anything, "test-package-1", update).Return(updatedPackage, nil)
			},
			expectedError: nil,
		},
		{
			name:      "package not found",
			packageID: "non-existent-package",
			setupMocks: func() {
				mockRepo.On("GetByID", mock.Anything, "non-existent-package").Return(nil, errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
		{
			name:      "already delivered",
			packageID: "delivered-package",
			setupMocks: func() {
				testPackage := &models.Package{
					PackageID: "delivered-package",
					UserID:    "test-user",
					Status:    "Delivered",
				}
				mockRepo.On("GetByID", mock.Anything, "delivered-package").Return(testPackage, nil)
			},
			expectedError: errors.New("cannot cancel a delivered package"),
		},
		{
			name:      "already canceled",
			packageID: "canceled-package",
			setupMocks: func() {
				testPackage := &models.Package{
					PackageID: "canceled-package",
					UserID:    "test-user",
					Status:    "Сanceled",
				}
				mockRepo.On("GetByID", mock.Anything, "canceled-package").Return(testPackage, nil)
			},
			expectedError: errors.New("package already canceled"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			result, err := packageService.CancelPackage(context.Background(), tt.packageID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Сanceled", result.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
