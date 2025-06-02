package repository

import (
	"context"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PackageRepository struct {
	collection *mongo.Collection
}

func NewPackageRepository(db *mongo.Database, collectionName string) *PackageRepository {
	return &PackageRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *PackageRepository) Create(ctx context.Context, pkg *models.Package) (*models.Package, error) {
	_, err := r.collection.InsertOne(ctx, pkg)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (r *PackageRepository) Update(ctx context.Context, pkg *models.Package) error {
	filter := bson.M{"package_id": pkg.PackageID}
	update := bson.M{"$set": pkg}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *PackageRepository) FindByID(ctx context.Context, packageID string) (*models.Package, error) {
	var pkg models.Package
	err := r.collection.FindOne(ctx, bson.M{"package_id": packageID}).Decode(&pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *PackageRepository) FindByAuctioningStatus(ctx context.Context) ([]*models.Package, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "Auctioning"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var packages []*models.Package
	for cursor.Next(ctx) {
		var pkg models.Package
		if err := cursor.Decode(&pkg); err != nil {
			continue
		}
		packages = append(packages, &pkg)
	}
	return packages, nil
}

func (r *PackageRepository) FindByFailedStatus(ctx context.Context) ([]*models.Package, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "Auction-failed"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var packages []*models.Package
	for cursor.Next(ctx) {
		var pkg models.Package
		if err := cursor.Decode(&pkg); err != nil {
			continue
		}
		packages = append(packages, &pkg)
	}
	return packages, nil
}
