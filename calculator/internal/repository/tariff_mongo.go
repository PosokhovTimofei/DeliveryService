package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/maksroxx/DeliveryService/calculator/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TariffRepository interface {
	GetAll(ctx context.Context) ([]models.Tariff, error)
	GetByCode(ctx context.Context, code string) (*models.Tariff, error)
	CreateTariff(ctx context.Context, tariff *models.Tariff) (*models.Tariff, error)
	DeleteTariff(ctx context.Context, code string) error
}

type mongoTariffRepo struct {
	collection *mongo.Collection
}

func NewTariffMongoRepository(db *mongo.Database, collectionName string) TariffRepository {
	collection := db.Collection(collectionName)

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "code", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create unique index: %v", err))
	}

	return &mongoTariffRepo{
		collection: collection,
	}
}

func (r *mongoTariffRepo) GetAll(ctx context.Context) ([]models.Tariff, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tariffs []models.Tariff
	if err := cursor.All(ctx, &tariffs); err != nil {
		return nil, err
	}
	return tariffs, nil
}

func (r *mongoTariffRepo) GetByCode(ctx context.Context, code string) (*models.Tariff, error) {
	filter := bson.M{"code": code}
	var tariff models.Tariff
	if err := r.collection.FindOne(ctx, filter).Decode(&tariff); err != nil {
		return nil, errors.New("tariff not found")
	}
	return &tariff, nil
}

func (r *mongoTariffRepo) CreateTariff(ctx context.Context, tariff *models.Tariff) (*models.Tariff, error) {
	_, err := r.collection.InsertOne(ctx, tariff)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("tariff has already exists")
		}
		return nil, err
	}
	return tariff, nil
}

func (r *mongoTariffRepo) DeleteTariff(ctx context.Context, code string) error {
	filter := bson.M{"code": code}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete tariff: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("tariff with code %s not found", code)
	}
	return nil
}
