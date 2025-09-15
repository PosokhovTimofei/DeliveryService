package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionFlow_ShortDuration(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.4",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}

	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer mongoContainer.Terminate(ctx)

	host, err := mongoContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := mongoContainer.MappedPort(ctx, "27017")
	assert.NoError(t, err)

	uri := "mongodb://" + host + ":" + port.Port()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)
	defer client.Disconnect(ctx)

	db := client.Database("test_auction_integration")

	packageRepo := repository.NewPackageRepository(db, "packages")
	bidRepo := repository.NewBidRepository(db, "bids")

	mockProducer := &MockAuctionPublisher{}
	logger := logrus.New()
	mockProducer.On("PublishPayment", mock.Anything, mock.AnythingOfType("*models.AuctionResult")).Return(nil)
	mockProducer.On("PublishNotification", mock.Anything, mock.AnythingOfType("*models.Notification")).Return(nil)

	auctionService := service.NewAuctionService(bidRepo, packageRepo, mockProducer, logger)
	auctionService.SetAuctionDuration(1 * time.Second)

	testPkg := &models.Package{
		PackageID: "test-package-1",
		Status:    "Waiting",
		From:      "Location A",
		To:        "Location B",
		Weight:    10.0,
		Cost:      50.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = packageRepo.Create(ctx, testPkg)
	assert.NoError(t, err)

	err = auctionService.StartWaitingAuctions(ctx)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	bid1 := &models.Bid{PackageID: "test-package-1", UserID: "user1", Amount: 60.0}
	err = auctionService.PlaceBid(ctx, bid1)
	assert.NoError(t, err)

	bid2 := &models.Bid{PackageID: "test-package-1", UserID: "user2", Amount: 70.0}
	err = auctionService.PlaceBid(ctx, bid2)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	updatedPkg, err := packageRepo.FindByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.Equal(t, "Finished", updatedPkg.Status)
	assert.Equal(t, "user2", updatedPkg.UserID)
	assert.Equal(t, 70.0, updatedPkg.Cost)
	mockProducer.AssertCalled(t, "PublishPayment", mock.Anything, mock.AnythingOfType("*models.AuctionResult"))
	mockProducer.AssertCalled(t, "PublishNotification", mock.Anything, mock.AnythingOfType("*models.Notification"))
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
