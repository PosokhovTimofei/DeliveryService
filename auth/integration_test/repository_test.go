package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/maksroxx/DeliveryService/auth/models"
	"github.com/maksroxx/DeliveryService/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoRepository_CreateUser(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err)

	var result models.User
	err = db.Collection("users").FindOne(ctx, bson.M{"user_id": "test-user-1"}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "user", result.Role)
}

func TestMongoRepository_CreateUser_Duplicate(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user1 := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	err := repo.CreateUser(ctx, user1)
	assert.NoError(t, err)

	user2 := &models.User{
		ID:                "test-user-1",
		Email:             "test2@example.com",
		EncryptedPassword: "hashedpassword2",
		Role:              "user",
	}

	err = repo.CreateUser(ctx, user2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user has already exists")
}

func TestMongoRepository_GetByID(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	_, err := db.Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	// Получаем пользователя по ID
	result, err := repo.GetByID(ctx, "test-user-1")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "user", result.Role)
}

func TestMongoRepository_GetByID_NotFound(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	result, err := repo.GetByID(ctx, "non-existent-user")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestMongoRepository_GetByEmail(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	_, err := db.Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	result, err := repo.GetByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "test-user-1", result.ID)
	assert.Equal(t, "user", result.Role)
}

func TestMongoRepository_UpdateUser(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	_, err := db.Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	updateFields := map[string]any{
		"email": "updated@example.com",
		"role":  "admin",
	}

	err = repo.UpdateUser(ctx, "test-user-1", updateFields)
	assert.NoError(t, err)

	result, err := repo.GetByID(ctx, "test-user-1")
	assert.NoError(t, err)
	assert.Equal(t, "updated@example.com", result.Email)
	assert.Equal(t, "admin", result.Role)
}

func TestMongoRepository_DeleteUser(t *testing.T) {
	ctx, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewMongoRepository(db, "users")

	user := &models.User{
		ID:                "test-user-1",
		Email:             "test@example.com",
		EncryptedPassword: "hashedpassword",
		Role:              "user",
	}

	_, err := db.Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	err = repo.DeleteUser(ctx, "test-user-1")
	assert.NoError(t, err)

	result, err := repo.GetByID(ctx, "test-user-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestTelegramAuthRepo_SaveAndFind(t *testing.T) {
	_, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTelegramAuthRepo(db, "telegram_auth_codes")

	err := repo.Save("123456", "test-user-1", 5*time.Minute)
	assert.NoError(t, err)

	userID, err := repo.FindUserIDByCode("123456")
	assert.NoError(t, err)
	assert.Equal(t, "test-user-1", userID)
}

func TestTelegramAuthRepo_FindExpiredCode(t *testing.T) {
	_, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTelegramAuthRepo(db, "telegram_auth_codes")

	err := repo.Save("123456", "test-user-1", 1*time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(2 * time.Millisecond)

	userID, err := repo.FindUserIDByCode("123456")
	assert.Error(t, err)
	assert.Empty(t, userID)
}

func TestTelegramAuthRepo_FindNonExistentCode(t *testing.T) {
	_, db, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTelegramAuthRepo(db, "telegram_auth_codes")

	userID, err := repo.FindUserIDByCode("non-existent-code")
	assert.Error(t, err)
	assert.Empty(t, userID)
}

func setupAuthTestEnvironment(t *testing.T) (context.Context, *mongo.Database, func()) {
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

	db := client.Database("test_auth_repository")

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
