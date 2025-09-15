package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoRepository_Create(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := repo.Create(ctx, pkg)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-package-1", result.PackageID)
	assert.Equal(t, "test-user", result.UserID)
}

func TestMongoRepository_Create_Duplicate(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := repo.Create(ctx, pkg)
	assert.NoError(t, err)

	duplicatePkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user-2",
		Weight:         15.0,
		Length:         25,
		Width:          20,
		Height:         15,
		From:           "Chicago",
		To:             "Miami",
		Address:        "456 Test Ave",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           75.0,
		EstimatedHours: 72,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := repo.Create(ctx, duplicatePkg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "package has already exists")
}

func TestMongoRepository_GetByID(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := db.Collection("packages").InsertOne(ctx, pkg)
	assert.NoError(t, err)

	result, err := repo.GetByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-package-1", result.PackageID)
	assert.Equal(t, "test-user", result.UserID)
	assert.Equal(t, "Created", result.Status)
}

func TestMongoRepository_GetByID_NotFound(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	result, err := repo.GetByID(ctx, "non-existent-package")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "route not found")
}

func TestMongoRepository_GetAllPackages(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	packages := []interface{}{
		models.Package{
			PackageID:      "package-1",
			UserID:         "user-1",
			Weight:         10.0,
			Length:         20,
			Width:          15,
			Height:         10,
			From:           "New York",
			To:             "Los Angeles",
			Address:        "123 Test St",
			PaymentStatus:  "PENDING",
			Status:         "Created",
			Cost:           50.0,
			EstimatedHours: 48,
			Currency:       "USD",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		models.Package{
			PackageID:      "package-2",
			UserID:         "user-1",
			Weight:         15.0,
			Length:         25,
			Width:          20,
			Height:         15,
			From:           "Chicago",
			To:             "Miami",
			Address:        "456 Test Ave",
			PaymentStatus:  "PAID",
			Status:         "Processing",
			Cost:           75.0,
			EstimatedHours: 72,
			Currency:       "USD",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		models.Package{
			PackageID:      "package-3",
			UserID:         "user-2",
			Weight:         20.0,
			Length:         30,
			Width:          25,
			Height:         20,
			From:           "Seattle",
			To:             "San Francisco",
			Address:        "789 Test Blvd",
			PaymentStatus:  "PENDING",
			Status:         "Created",
			Cost:           100.0,
			EstimatedHours: 96,
			Currency:       "USD",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	_, err := db.Collection("packages").InsertMany(ctx, packages)
	assert.NoError(t, err)

	filter := models.PackageFilter{
		UserID: "user-1",
		Limit:  10,
		Offset: 0,
	}

	result, err := repo.GetAllPackages(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	for _, pkg := range result {
		assert.Equal(t, "user-1", pkg.UserID)
	}
}

func TestMongoRepository_UpdatePackage(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Вставляем пакет
	_, err := db.Collection("packages").InsertOne(ctx, pkg)
	assert.NoError(t, err)

	update := models.PackageUpdate{
		Status:        "Processing",
		PaymentStatus: "PAID",
	}

	result, err := repo.UpdatePackage(ctx, "test-package-1", update)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Processing", result.Status)
	assert.Equal(t, "PAID", result.PaymentStatus)
}

func TestMongoRepository_DeletePackage(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := db.Collection("packages").InsertOne(ctx, pkg)
	assert.NoError(t, err)

	err = repo.DeletePackage(ctx, "test-package-1")
	assert.NoError(t, err)

	result, err := repo.GetByID(ctx, "test-package-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "route not found")
}

func TestMongoRepository_MarkAsExpiredByID(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := db.Collection("packages").InsertOne(ctx, pkg)
	assert.NoError(t, err)

	result, err := repo.MarkAsExpiredByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "In pick-up point", result.Status)
}

func TestMongoRepository_GetExpiredPackages(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	expiredTime := time.Now().AddDate(0, 0, -61)
	packages := []interface{}{
		bson.M{
			"package_id":      "expired-package-1",
			"user_id":         "test-user",
			"weight":          10.0,
			"length":          20,
			"width":           15,
			"height":          10,
			"from":            "New York",
			"to":              "Los Angeles",
			"address":         "123 Test St",
			"payment_status":  "PENDING",
			"status":          "In pick-up point",
			"cost":            50.0,
			"estimated_hours": 48,
			"currency":        "USD",
			"created_at":      expiredTime,
			"updated_at":      expiredTime,
		},
		bson.M{
			"package_id":      "expired-package-2",
			"user_id":         "test-user",
			"weight":          15.0,
			"length":          25,
			"width":           20,
			"height":          15,
			"from":            "Chicago",
			"to":              "Miami",
			"address":         "456 Test Ave",
			"payment_status":  "PAID",
			"status":          "In pick-up point",
			"cost":            75.0,
			"estimated_hours": 72,
			"currency":        "USD",
			"created_at":      expiredTime,
			"updated_at":      expiredTime,
		},
	}

	_, err := db.Collection("packages").InsertMany(ctx, packages)
	assert.NoError(t, err)

	result, err := repo.GetExpiredPackages(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	for _, pkg := range result {
		assert.Equal(t, "In pick-up point", pkg.Status)
	}
}

func TestMongoRepository_Ping(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	err := repo.Ping(ctx)
	assert.NoError(t, err)
}

func TestMongoRepository_AlreadyCreatedToday(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	for i := 1; i <= 3; i++ {
		pkg.PackageID = fmt.Sprintf("test-package-%d", i)
		_, err := repo.Create(ctx, pkg)
		assert.NoError(t, err)
	}

	pkg.PackageID = "test-package-4"
	result, err := repo.Create(ctx, pkg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "limit: only 3 identical packages allowed per day")
}

func TestMongoRepository_CalculateRemainingHours(t *testing.T) {
	ctx, db, cleanup := setupDatabaseTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "packages")

	pkg := &models.Package{
		PackageID:      "test-package-1",
		UserID:         "test-user",
		Weight:         10.0,
		Length:         20,
		Width:          15,
		Height:         10,
		From:           "New York",
		To:             "Los Angeles",
		Address:        "123 Test St",
		PaymentStatus:  "PENDING",
		Status:         "Created",
		Cost:           50.0,
		EstimatedHours: 48,
		Currency:       "USD",
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		UpdatedAt:      time.Now().Add(-24 * time.Hour),
	}

	// Вставляем пакет
	_, err := db.Collection("packages").InsertOne(ctx, pkg)
	assert.NoError(t, err)

	result, err := repo.GetByID(ctx, "test-package-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 24, result.RemainingHours)
}

func setupDatabaseTestEnvironment(t *testing.T) (context.Context, *mongo.Database, func()) {
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

	db := client.Database("test_database_repository")

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
