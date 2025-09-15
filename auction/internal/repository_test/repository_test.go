package repositorytest

import (
	"context"
	"testing"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) *mongo.Database {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	return client.Database("test_auction")
}

func cleanupTestDB(t *testing.T, db *mongo.Database) {
	err := db.Drop(context.TODO())
	assert.NoError(t, err)
}

func TestBidRepository_PlaceBid(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := repository.NewBidRepository(db, "bids")
	bid := &models.Bid{
		PackageID: "test-package",
		UserID:    "test-user",
		Amount:    100.0,
	}

	err := repo.PlaceBid(context.TODO(), bid)
	assert.NoError(t, err)

	var result models.Bid
	err = db.Collection("bids").FindOne(context.TODO(), bson.M{"package_id": "test-package"}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", result.UserID)
	assert.Equal(t, 100.0, result.Amount)
}

func TestBidRepository_GetBidsByPackage(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := repository.NewBidRepository(db, "bids")

	bids := []interface{}{
		models.Bid{PackageID: "test-package", UserID: "user1", Amount: 100},
		models.Bid{PackageID: "test-package", UserID: "user2", Amount: 150},
	}
	_, err := db.Collection("bids").InsertMany(context.TODO(), bids)
	assert.NoError(t, err)

	result, err := repo.GetBidsByPackage(context.TODO(), "test-package")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
