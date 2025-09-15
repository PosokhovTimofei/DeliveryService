package integration

import (
	"context"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/telegram/internal/models"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserLinkRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	client    *mongo.Client
	database  *mongo.Database
	repo      *repository.UserLinkRepository
	ctx       context.Context
}

func (s *UserLinkRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:6.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = container

	host, err := container.Host(s.ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(s.ctx, "27017")
	s.Require().NoError(err)

	uri := "mongodb://" + host + ":" + port.Port()
	client, err := mongo.Connect(s.ctx, options.Client().ApplyURI(uri))
	s.Require().NoError(err)

	s.client = client
	s.database = client.Database("test_telegram_integration")

	s.repo = repository.NewUserLinkRepository(s.database, "user_links")
}

func (s *UserLinkRepositoryTestSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Disconnect(s.ctx)
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *UserLinkRepositoryTestSuite) SetupTest() {
	err := s.database.Collection("user_links").Drop(s.ctx)
	if err != nil && err.Error() != "ns not found" {
		s.Require().NoError(err)
	}
}

func TestUserLinkRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserLinkRepositoryTestSuite))
}

func (s *UserLinkRepositoryTestSuite) TestSaveLink_NewLink() {
	telegramID := int64(123456789)
	userID := "user-123"

	err := s.repo.SaveLink(s.ctx, telegramID, userID)

	s.NoError(err)

	var result models.TelegramUserLink
	err = s.database.Collection("user_links").
		FindOne(s.ctx, bson.M{"telegram_id": telegramID}).
		Decode(&result)

	s.NoError(err)
	s.Equal(telegramID, result.TelegramID)
	s.Equal(userID, result.UserID)
	s.WithinDuration(time.Now(), result.LinkedAt, 5*time.Second)
}

func (s *UserLinkRepositoryTestSuite) TestSaveLink_UpdateExistingLink() {
	telegramID := int64(123456789)
	initialUserID := "user-123"
	newUserID := "user-456"

	err := s.repo.SaveLink(s.ctx, telegramID, initialUserID)
	s.NoError(err)

	err = s.repo.SaveLink(s.ctx, telegramID, newUserID)

	s.NoError(err)

	var result models.TelegramUserLink
	err = s.database.Collection("user_links").
		FindOne(s.ctx, bson.M{"telegram_id": telegramID}).
		Decode(&result)

	s.NoError(err)
	s.Equal(telegramID, result.TelegramID)
	s.Equal(newUserID, result.UserID)
	s.WithinDuration(time.Now(), result.LinkedAt, 5*time.Second)
}

func (s *UserLinkRepositoryTestSuite) TestGetUserIDByTelegramID_Exists() {
	telegramID := int64(123456789)
	userID := "user-123"

	err := s.repo.SaveLink(s.ctx, telegramID, userID)
	s.NoError(err)

	result, err := s.repo.GetUserIDByTelegramID(s.ctx, telegramID)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *UserLinkRepositoryTestSuite) TestGetUserIDByTelegramID_NotExists() {
	// Arrange
	telegramID := int64(999999999)

	result, err := s.repo.GetUserIDByTelegramID(s.ctx, telegramID)

	s.Error(err)
	s.Empty(result)
	s.Equal(mongo.ErrNoDocuments, err)
}

func (s *UserLinkRepositoryTestSuite) TestGetTelegramIDByUserID_Exists() {
	telegramID := int64(123456789)
	userID := "user-123"

	err := s.repo.SaveLink(s.ctx, telegramID, userID)
	s.NoError(err)

	result, err := s.repo.GetTelegramIDByUserID(s.ctx, userID)

	s.NoError(err)
	s.Equal(telegramID, result)
}

func (s *UserLinkRepositoryTestSuite) TestGetTelegramIDByUserID_NotExists() {
	userID := "non-existent-user"

	result, err := s.repo.GetTelegramIDByUserID(s.ctx, userID)

	s.Error(err)
	s.Zero(result)
	s.Equal(mongo.ErrNoDocuments, err)
}

func (s *UserLinkRepositoryTestSuite) TestConcurrentOperations() {
	telegramID := int64(123456789)
	userID := "user-123"

	s.Run("ConcurrentSaveAndRead", func() {
		s.T().Parallel()

		err := s.repo.SaveLink(s.ctx, telegramID, userID)
		s.NoError(err)

		result, err := s.repo.GetUserIDByTelegramID(s.ctx, telegramID)
		s.NoError(err)
		s.Equal(userID, result)
	})

	s.Run("ConcurrentReadNonExistent", func() {
		s.T().Parallel()

		_, err := s.repo.GetTelegramIDByUserID(s.ctx, "non-existent")
		s.Error(err)
	})
}
