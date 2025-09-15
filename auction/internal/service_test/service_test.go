package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

type MockBidRepository struct {
	mock.Mock
}

func (m *MockBidRepository) PlaceBid(ctx context.Context, bid *models.Bid) error {
	args := m.Called(ctx, bid)
	return args.Error(0)
}

func (m *MockBidRepository) GetBidsByPackage(ctx context.Context, packageID string) ([]*models.Bid, error) {
	args := m.Called(ctx, packageID)
	return args.Get(0).([]*models.Bid), args.Error(1)
}

func (m *MockBidRepository) WatchBidsByPackage(ctx context.Context, packageID string) (*mongo.ChangeStream, error) {
	args := m.Called(ctx, packageID)
	return args.Get(0).(*mongo.ChangeStream), args.Error(1)
}

func (m *MockBidRepository) GetTopBidByPackage(ctx context.Context, packageID string) (*models.Bid, error) {
	args := m.Called(ctx, packageID)
	return args.Get(0).(*models.Bid), args.Error(1)
}

type MockPackageRepository struct {
	mock.Mock
}

func (m *MockPackageRepository) Create(ctx context.Context, pkg *models.Package) (*models.Package, error) {
	args := m.Called(ctx, pkg)
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockPackageRepository) Update(ctx context.Context, pkg *models.Package) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}

func (m *MockPackageRepository) FindByID(ctx context.Context, packageID string) (*models.Package, error) {
	args := m.Called(ctx, packageID)
	return args.Get(0).(*models.Package), args.Error(1)
}

func (m *MockPackageRepository) FindUserPackages(ctx context.Context, userId string) ([]*models.Package, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]*models.Package), args.Error(1)
}

func (m *MockPackageRepository) FindByFailedStatus(ctx context.Context) ([]*models.Package, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Package), args.Error(1)
}

func (m *MockPackageRepository) FindByAuctioningStatus(ctx context.Context) ([]*models.Package, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Package), args.Error(1)
}

func (m *MockPackageRepository) FindByWaitingStatus(ctx context.Context) ([]*models.Package, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Package), args.Error(1)
}

type MockAuctionPublisher struct {
	mock.Mock
}

func (m *MockAuctionPublisher) PublishPayment(ctx context.Context, result *models.AuctionResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockAuctionPublisher) PublishNotification(ctx context.Context, note *models.Notification) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *MockAuctionPublisher) PublishDeliveryInit(ctx context.Context, init *models.DeliveryInit) error {
	args := m.Called(ctx, init)
	return args.Error(0)
}

func (m *MockAuctionPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockConsumerGroupSession struct {
	mock.Mock
}

func (m *MockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	m.Called(msg, metadata)
}

func (m *MockConsumerGroupSession) Commit() {
	m.Called()
}

func (m *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	m.Called(topic, partition, offset, metadata)
}

func (m *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	m.Called(topic, partition, offset, metadata)
}

func (m *MockConsumerGroupSession) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func TestAuctionService_PlaceBid(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockPkgRepo := new(MockPackageRepository)
	mockProducer := new(MockAuctionPublisher)
	logger := logrus.New()

	auctionService := service.NewAuctionService(mockBidRepo, mockPkgRepo, mockProducer, logger)

	tests := []struct {
		name          string
		bid           *models.Bid
		setupMocks    func()
		expectedError error
	}{
		{
			name: "successful bid",
			bid:  &models.Bid{PackageID: "test-pkg", UserID: "user1", Amount: 100},
			setupMocks: func() {
				mockPkgRepo.On("FindByID", mock.Anything, "test-pkg").Return(&models.Package{
					Status:    "Auctioning",
					UpdatedAt: time.Now().Add(-time.Minute),
				}, nil)
				mockBidRepo.On("GetTopBidByPackage", mock.Anything, "test-pkg").Return(&models.Bid{Amount: 50}, nil)
				mockBidRepo.On("PlaceBid", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "bid too low",
			bid:  &models.Bid{PackageID: "test-pkg", UserID: "user1", Amount: 40},
			setupMocks: func() {
				mockPkgRepo.On("FindByID", mock.Anything, "test-pkg").Return(&models.Package{
					Status:    "Auctioning",
					UpdatedAt: time.Now().Add(-time.Minute),
				}, nil)
				mockBidRepo.On("GetTopBidByPackage", mock.Anything, "test-pkg").Return(&models.Bid{Amount: 50}, nil)
			},
			expectedError: errors.New("bid must be greater than current highest"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := auctionService.PlaceBid(context.TODO(), tt.bid)
			assert.Equal(t, tt.expectedError, err)
			mockBidRepo.AssertExpectations(t)
			mockPkgRepo.AssertExpectations(t)
		})
	}
}
