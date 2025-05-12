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

type CountryRepository interface {
	GetCoordinates(ctx context.Context, country string) (*models.CountryCoordinates, error)
}

type mongoCityRepo struct {
	collection *mongo.Collection
}

func NewCityMongoRepository(db *mongo.Database, collectionName string) CountryRepository {
	collection := db.Collection(collectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create unique index: %v", err))
	}
	return &mongoCityRepo{collection: db.Collection(collectionName)}
}

func (r *mongoCityRepo) GetCoordinates(ctx context.Context, country string) (*models.CountryCoordinates, error) {
	filter := bson.M{"name": bson.M{"$regex": "^" + country + "$", "$options": "i"}}
	var doc models.CountryDoc
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, errors.New("country not found in DB")
	}
	return &models.CountryCoordinates{
		Name:      doc.Name,
		Latitude:  doc.Location.Coordinates[1],
		Longitude: doc.Location.Coordinates[0],
		Zone:      doc.Zone,
	}, nil
}
