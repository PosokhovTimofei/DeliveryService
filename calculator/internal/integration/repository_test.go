package integration

import (
	"context"
	"testing"

	"github.com/maksroxx/DeliveryService/calculator/internal/repository"
	"github.com/maksroxx/DeliveryService/calculator/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoCityRepo_GetCoordinates(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	countries := []interface{}{
		bson.M{
			"name": "United States",
			"location": bson.M{
				"type":        "Point",
				"coordinates": []float64{-98.5795, 39.8283},
			},
			"zone": "America/New_York",
		},
		bson.M{
			"name": "Germany",
			"location": bson.M{
				"type":        "Point",
				"coordinates": []float64{10.4515, 51.1657},
			},
			"zone": "Europe/Berlin",
		},
	}

	_, err := db.Collection("countries").InsertMany(ctx, countries)
	assert.NoError(t, err)

	repo := repository.NewCityMongoRepository(db, "countries")

	tests := []struct {
		name          string
		country       string
		expectedName  string
		expectedLat   float64
		expectedLon   float64
		expectedZone  string
		expectedError bool
	}{
		{
			name:          "existing country",
			country:       "United States",
			expectedName:  "United States",
			expectedLat:   39.8283,
			expectedLon:   -98.5795,
			expectedZone:  "America/New_York",
			expectedError: false,
		},
		{
			name:          "another existing country",
			country:       "Germany",
			expectedName:  "Germany",
			expectedLat:   51.1657,
			expectedLon:   10.4515,
			expectedZone:  "Europe/Berlin",
			expectedError: false,
		},
		{
			name:          "non-existent country",
			country:       "NonExistentCountry",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coordinates, err := repo.GetCoordinates(ctx, tt.country)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, coordinates)
				assert.Contains(t, err.Error(), "country not found in DB")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, coordinates)
				assert.Equal(t, tt.expectedName, coordinates.Name)
				assert.Equal(t, tt.expectedLat, coordinates.Latitude)
				assert.Equal(t, tt.expectedLon, coordinates.Longitude)
				assert.Equal(t, tt.expectedZone, coordinates.Zone)
			}
		})
	}
}

func TestMongoTariffRepo_GetAll(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	tariffs := []interface{}{
		models.Tariff{
			Code:              "STANDARD",
			Name:              "Standard Delivery",
			BaseRate:          10.0,
			PricePerKm:        0.5,
			PricePerKg:        1.0,
			Currency:          "USD",
			VolumetricDivider: 5000,
			SpeedKmph:         60,
		},
		models.Tariff{
			Code:              "EXPRESS",
			Name:              "Express Delivery",
			BaseRate:          20.0,
			PricePerKm:        0.8,
			PricePerKg:        1.5,
			Currency:          "USD",
			VolumetricDivider: 5000,
			SpeedKmph:         100,
		},
	}

	_, err := db.Collection("tariffs").InsertMany(ctx, tariffs)
	assert.NoError(t, err)

	repo := repository.NewTariffMongoRepository(db, "tariffs")

	result, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, "STANDARD", result[0].Code)
	assert.Equal(t, "Standard Delivery", result[0].Name)
	assert.Equal(t, 10.0, result[0].BaseRate)

	assert.Equal(t, "EXPRESS", result[1].Code)
	assert.Equal(t, "Express Delivery", result[1].Name)
	assert.Equal(t, 20.0, result[1].BaseRate)
}

func TestMongoTariffRepo_GetByCode(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	tariff := models.Tariff{
		Code:              "STANDARD",
		Name:              "Standard Delivery",
		BaseRate:          10.0,
		PricePerKm:        0.5,
		PricePerKg:        1.0,
		Currency:          "USD",
		VolumetricDivider: 5000,
		SpeedKmph:         60,
	}

	_, err := db.Collection("tariffs").InsertOne(ctx, tariff)
	assert.NoError(t, err)

	repo := repository.NewTariffMongoRepository(db, "tariffs")

	tests := []struct {
		name          string
		code          string
		expectedName  string
		expectedError bool
	}{
		{
			name:         "existing tariff",
			code:         "STANDARD",
			expectedName: "Standard Delivery",
		},
		{
			name:          "non-existent tariff",
			code:          "NON_EXISTENT",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByCode(ctx, tt.code)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "tariff not found")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.code, result.Code)
				assert.Equal(t, tt.expectedName, result.Name)
			}
		})
	}
}

func TestMongoTariffRepo_CreateTariff(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTariffMongoRepository(db, "tariffs")

	tariff := &models.Tariff{
		Code:              "STANDARD",
		Name:              "Standard Delivery",
		BaseRate:          10.0,
		PricePerKm:        0.5,
		PricePerKg:        1.0,
		Currency:          "USD",
		VolumetricDivider: 5000,
		SpeedKmph:         60,
	}

	result, err := repo.CreateTariff(ctx, tariff)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "STANDARD", result.Code)

	created, err := repo.GetByCode(ctx, "STANDARD")
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "Standard Delivery", created.Name)
	assert.Equal(t, 10.0, created.BaseRate)
}

func TestMongoTariffRepo_DeleteTariff(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTariffMongoRepository(db, "tariffs")

	tariff := &models.Tariff{
		Code:              "STANDARD",
		Name:              "Standard Delivery",
		BaseRate:          10.0,
		PricePerKm:        0.5,
		PricePerKg:        1.0,
		Currency:          "USD",
		VolumetricDivider: 5000,
		SpeedKmph:         60,
	}

	_, err := repo.CreateTariff(ctx, tariff)
	assert.NoError(t, err)

	err = repo.DeleteTariff(ctx, "STANDARD")
	assert.NoError(t, err)

	result, err := repo.GetByCode(ctx, "STANDARD")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tariff not found")
}

func TestMongoTariffRepo_DeleteTariff_NotFound(t *testing.T) {
	ctx, db, cleanup := setupCalculatorTestEnvironment(t)
	defer cleanup()

	repo := repository.NewTariffMongoRepository(db, "tariffs")

	err := repo.DeleteTariff(ctx, "NON_EXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tariff with code NON_EXISTENT not found")
}

func setupCalculatorTestEnvironment(t *testing.T) (context.Context, *mongo.Database, func()) {
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

	db := client.Database("test_calculator_repository")

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
