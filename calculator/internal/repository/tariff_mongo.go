package repository

import (
	"context"
	"errors"

	"github.com/maksroxx/DeliveryService/calculator/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TariffRepository interface {
	GetAll(ctx context.Context) ([]models.Tariff, error)
	GetByCode(ctx context.Context, code string) (*models.Tariff, error)
}

type mongoTariffRepo struct {
	collection *mongo.Collection
}

func NewTariffMongoRepository(db *mongo.Database, collectionName string) TariffRepository {
	return &mongoTariffRepo{
		collection: db.Collection(collectionName),
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
