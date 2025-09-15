package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestBidRepository_PlaceBid(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewBidRepository(db, "bids")

	bid := &models.Bid{
		PackageID: "test-package-1",
		UserID:    "test-user",
		Amount:    100.0,
	}

	err := repo.PlaceBid(ctx, bid)
	assert.NoError(t, err)

	var result models.Bid
	err = db.Collection("bids").FindOne(ctx, bson.M{"package_id": "test-package-1"}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", result.UserID)
	assert.Equal(t, 100.0, result.Amount)
	assert.False(t, result.Timestamp.IsZero())
}

func TestBidRepository_GetBidsByPackage(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewBidRepository(db, "bids")

	bids := []interface{}{
		models.Bid{PackageID: "test-package-1", UserID: "user1", Amount: 100, Timestamp: time.Now()},
		models.Bid{PackageID: "test-package-1", UserID: "user2", Amount: 150, Timestamp: time.Now()},
		models.Bid{PackageID: "test-package-2", UserID: "user3", Amount: 200, Timestamp: time.Now()},
	}

	_, err := db.Collection("bids").InsertMany(ctx, bids)
	assert.NoError(t, err)

	result, err := repo.GetBidsByPackage(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, "user1", result[0].UserID)
	assert.Equal(t, 100.0, result[0].Amount)

	assert.Equal(t, "user2", result[1].UserID)
	assert.Equal(t, 150.0, result[1].Amount)
}

func TestBidRepository_GetTopBidByPackage(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewBidRepository(db, "bids")

	bids := []interface{}{
		models.Bid{PackageID: "test-package-1", UserID: "user1", Amount: 100, Timestamp: time.Now()},
		models.Bid{PackageID: "test-package-1", UserID: "user2", Amount: 150, Timestamp: time.Now()},
		models.Bid{PackageID: "test-package-1", UserID: "user3", Amount: 120, Timestamp: time.Now()},
	}

	_, err := db.Collection("bids").InsertMany(ctx, bids)
	assert.NoError(t, err)

	result, err := repo.GetTopBidByPackage(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.Equal(t, "user2", result.UserID)
	assert.Equal(t, 150.0, result.Amount)
}

func TestPackageRepository_CreateAndFind(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewPackageRepository(db, "packages")

	pkg := &models.Package{
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

	_, err := repo.Create(ctx, pkg)
	assert.NoError(t, err)

	result, err := repo.FindByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.Equal(t, "Waiting", result.Status)
	assert.Equal(t, 10.0, result.Weight)
	assert.Equal(t, 50.0, result.Cost)
}

func TestPackageRepository_Update(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewPackageRepository(db, "packages")

	pkg := &models.Package{
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

	_, err := repo.Create(ctx, pkg)
	assert.NoError(t, err)

	pkg.Status = "Auctioning"
	err = repo.Update(ctx, pkg)
	assert.NoError(t, err)

	result, err := repo.FindByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.Equal(t, "Auctioning", result.Status)
}

func TestPackageRepository_FindByStatus(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewPackageRepository(db, "packages")

	packages := []interface{}{
		models.Package{PackageID: "pkg-1", Status: "Waiting", From: "A", To: "B", Weight: 10.0, Cost: 50.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		models.Package{PackageID: "pkg-2", Status: "Auctioning", From: "A", To: "B", Weight: 15.0, Cost: 75.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		models.Package{PackageID: "pkg-3", Status: "Waiting", From: "A", To: "B", Weight: 20.0, Cost: 100.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		models.Package{PackageID: "pkg-4", Status: "Auction-failed", From: "A", To: "B", Weight: 25.0, Cost: 125.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	_, err := db.Collection("packages").InsertMany(ctx, packages)
	assert.NoError(t, err)

	waitingPkgs, err := repo.FindByWaitingStatus(ctx)
	assert.NoError(t, err)
	assert.Len(t, waitingPkgs, 2)

	auctioningPkgs, err := repo.FindByAuctioningStatus(ctx)
	assert.NoError(t, err)
	assert.Len(t, auctioningPkgs, 1)
	assert.Equal(t, "pkg-2", auctioningPkgs[0].PackageID)

	failedPkgs, err := repo.FindByFailedStatus(ctx)
	assert.NoError(t, err)
	assert.Len(t, failedPkgs, 1)
	assert.Equal(t, "pkg-4", failedPkgs[0].PackageID)
}

func TestPackageRepository_FindUserPackages(t *testing.T) {
	ctx, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	repo := repository.NewPackageRepository(db, "packages")

	packages := []interface{}{
		models.Package{PackageID: "pkg-1", UserID: "user-1", Status: "Waiting", From: "A", To: "B", Weight: 10.0, Cost: 50.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		models.Package{PackageID: "pkg-2", UserID: "user-2", Status: "Auctioning", From: "A", To: "B", Weight: 15.0, Cost: 75.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		models.Package{PackageID: "pkg-3", UserID: "user-1", Status: "Finished", From: "A", To: "B", Weight: 20.0, Cost: 100.0, Currency: "USD", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	_, err := db.Collection("packages").InsertMany(ctx, packages)
	assert.NoError(t, err)

	userPkgs, err := repo.FindUserPackages(ctx, "user-1")
	assert.NoError(t, err)
	assert.Len(t, userPkgs, 2)

	pkgIDs := make([]string, len(userPkgs))
	for i, pkg := range userPkgs {
		pkgIDs[i] = pkg.PackageID
	}
	assert.Contains(t, pkgIDs, "pkg-1")
	assert.Contains(t, pkgIDs, "pkg-3")
}

func setupTestEnvironment(t *testing.T) (context.Context, *mongo.Database, func()) {
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

	host, err := mongoContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := mongoContainer.MappedPort(ctx, "27017")
	assert.NoError(t, err)

	uri := "mongodb://" + host + ":" + port.Port()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)

	db := client.Database("test_auction_repository")

	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err == nil {
		for _, coll := range collections {
			db.Collection(coll).Drop(ctx)
		}
	}

	cleanup := func() {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
	}

	return ctx, db, cleanup
}
