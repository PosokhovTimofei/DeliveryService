package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/maksroxx/DeliveryService/producer/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PackageRepository struct {
	collection *mongo.Collection
}

func NewPackageRepository(db *mongo.Database, collectionName string) *PackageRepository {
	collection := db.Collection(collectionName)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "weight", Value: 1},
			{Key: "from", Value: 1},
			{Key: "to", Value: 1},
			{Key: "address", Value: 1},
			{Key: "cost", Value: 1},
			{Key: "estimated_hours", Value: 1},
			{Key: "currency", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create unique index: %v", err))
	}
	return &PackageRepository{collection: collection}
}

func (r *PackageRepository) CreatePackage(ctx context.Context, pkg pkg.Package) error {
	pkgWithoutID := pkg
	pkgWithoutID.ID = ""

	exists, err := r.PackageExists(ctx, pkgWithoutID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("package already exists for this user")
	}

	_, err = r.collection.InsertOne(ctx, pkg)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("package already exists for this user")
		}
		return err
	}
	return nil
}

func (r *PackageRepository) PackageExists(ctx context.Context, pkg pkg.Package) (bool, error) {
	filter := bson.M{
		"user_id":         pkg.UserID,
		"weight":          pkg.Weight,
		"from":            pkg.From,
		"to":              pkg.To,
		"address":         pkg.Address,
		"cost":            pkg.Cost,
		"estimated_hours": pkg.EstimatedHours,
		"currency":        pkg.Currency,
		"length":          pkg.Length,
		"width":           pkg.Width,
		"height":          pkg.Height,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
