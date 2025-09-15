package integration

import (
	"context"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestPaymentMongoRepository_CreatePayment(t *testing.T) {
	ctx, base, cleanup := setupPaymentTestEnvironment(t)
	defer cleanup()

	repo := db.NewPaymentMongoRepository(base, "payments")

	payment := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Cost:      50.0,
		Currency:  "USD",
		Status:    models.PaymentStatusPending,
	}

	err := repo.CreatePayment(ctx, payment)
	assert.NoError(t, err)

	var result bson.M
	err = base.Collection("payments").FindOne(ctx, bson.M{"user_id": "test-user", "package_id": "test-package-1"}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", result["user_id"])
	assert.Equal(t, "test-package-1", result["package_id"])
	assert.Equal(t, 50.0, result["cost"])
	assert.Equal(t, "USD", result["currency"])
	assert.Equal(t, "PENDING", result["status"])
	assert.NotNil(t, result["created_at"])
	assert.NotNil(t, result["updated_at"])
}

func TestPaymentMongoRepository_CreatePayment_Duplicate(t *testing.T) {
	ctx, base, cleanup := setupPaymentTestEnvironment(t)
	defer cleanup()

	repo := db.NewPaymentMongoRepository(base, "payments")

	payment := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Cost:      50.0,
		Currency:  "USD",
		Status:    models.PaymentStatusPending,
	}

	err := repo.CreatePayment(ctx, payment)
	assert.NoError(t, err)

	duplicatePayment := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Cost:      60.0,
		Currency:  "EUR",
		Status:    models.PaymentStatusPending,
	}

	err = repo.CreatePayment(ctx, duplicatePayment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment has already exists")
}

func TestPaymentMongoRepository_UpdatePayment(t *testing.T) {
	ctx, base, cleanup := setupPaymentTestEnvironment(t)
	defer cleanup()

	repo := db.NewPaymentMongoRepository(base, "payments")

	payment := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Cost:      50.0,
		Currency:  "USD",
		Status:    models.PaymentStatusPending,
	}

	doc := bson.M{
		"user_id":    payment.UserID,
		"package_id": payment.PackageID,
		"cost":       payment.Cost,
		"currency":   payment.Currency,
		"status":     payment.Status,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	_, err := base.Collection("payments").InsertOne(ctx, doc)
	assert.NoError(t, err)

	update := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Status:    models.PaymentStatusPaid,
	}

	result, err := repo.UpdatePayment(ctx, update)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.PaymentStatusPaid, result.Status)

	var updatedDoc bson.M
	err = base.Collection("payments").FindOne(ctx, bson.M{"user_id": "test-user", "package_id": "test-package-1"}).Decode(&updatedDoc)
	assert.NoError(t, err)
	assert.Equal(t, "PAID", updatedDoc["status"])
}

func TestPaymentMongoRepository_UpdatePayment_AlreadyPaid(t *testing.T) {
	ctx, base, cleanup := setupPaymentTestEnvironment(t)
	defer cleanup()

	repo := db.NewPaymentMongoRepository(base, "payments")

	doc := bson.M{
		"user_id":    "test-user",
		"package_id": "test-package-1",
		"cost":       50.0,
		"currency":   "USD",
		"status":     models.PaymentStatusPaid,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	_, err := base.Collection("payments").InsertOne(ctx, doc)
	assert.NoError(t, err)

	update := models.Payment{
		UserID:    "test-user",
		PackageID: "test-package-1",
		Status:    models.PaymentStatusCancelled,
	}

	result, err := repo.UpdatePayment(ctx, update)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "payment already confirmed")
}

func TestPaymentMongoRepository_UpdatePayment_NotFound(t *testing.T) {
	ctx, base, cleanup := setupPaymentTestEnvironment(t)
	defer cleanup()

	repo := db.NewPaymentMongoRepository(base, "payments")

	update := models.Payment{
		UserID:    "non-existent-user",
		PackageID: "non-existent-package",
		Status:    models.PaymentStatusPaid,
	}

	result, err := repo.UpdatePayment(ctx, update)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "payment already confirmed")
}

func setupPaymentTestEnvironment(t *testing.T) (context.Context, *mongo.Database, func()) {
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

	db := client.Database("test_payment_repository")

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
