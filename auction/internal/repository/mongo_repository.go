package repository

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/metrics"
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
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("Create").Observe(time.Since(start).Seconds())
	}()

	_, err := r.collection.InsertOne(ctx, pkg)
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.PackageOpsCount.WithLabelValues("Create", status).Inc()
	return pkg, err
}

func (r *PackageRepository) Update(ctx context.Context, pkg *models.Package) error {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("Update").Observe(time.Since(start).Seconds())
	}()

	filter := bson.M{"package_id": pkg.PackageID}
	update := bson.M{"$set": pkg}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.PackageOpsCount.WithLabelValues("Update", status).Inc()
	return err
}

func (r *PackageRepository) FindByID(ctx context.Context, packageID string) (*models.Package, error) {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("FindByID").Observe(time.Since(start).Seconds())
	}()

	var pkg models.Package
	err := r.collection.FindOne(ctx, bson.M{"package_id": packageID}).Decode(&pkg)
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.PackageOpsCount.WithLabelValues("FindByID", status).Inc()
	return &pkg, err
}

func (r *PackageRepository) FindUserPackages(ctx context.Context, userId string) ([]*models.Package, error) {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("FindUserPackages").Observe(time.Since(start).Seconds())
	}()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userId})
	status := "success"
	if err != nil {
		status = "error"
		metrics.PackageOpsCount.WithLabelValues("FindUserPackages", status).Inc()
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
	metrics.PackageOpsCount.WithLabelValues("FindUserPackages", status).Inc()
	return packages, nil
}

func (r *PackageRepository) FindByAuctioningStatus(ctx context.Context) ([]*models.Package, error) {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("FindByAuctioningStatus").Observe(time.Since(start).Seconds())
	}()

	cursor, err := r.collection.Find(ctx, bson.M{"status": "Auctioning"})
	status := "success"
	if err != nil {
		status = "error"
		metrics.PackageOpsCount.WithLabelValues("FindByAuctioningStatus", status).Inc()
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
	metrics.PackageOpsCount.WithLabelValues("FindByAuctioningStatus", status).Inc()
	return packages, nil
}

func (r *PackageRepository) FindByFailedStatus(ctx context.Context) ([]*models.Package, error) {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("FindByFailedStatus").Observe(time.Since(start).Seconds())
	}()

	cursor, err := r.collection.Find(ctx, bson.M{"status": "Auction-failed"})
	status := "success"
	if err != nil {
		status = "error"
		metrics.PackageOpsCount.WithLabelValues("FindByFailedStatus", status).Inc()
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
	metrics.PackageOpsCount.WithLabelValues("FindByFailedStatus", status).Inc()
	return packages, nil
}

func (r *PackageRepository) FindByWaitingStatus(ctx context.Context) ([]*models.Package, error) {
	start := time.Now()
	defer func() {
		metrics.PackageOpsDuration.WithLabelValues("FindByWaitingStatus").Observe(time.Since(start).Seconds())
	}()

	cursor, err := r.collection.Find(ctx, bson.M{"status": "Waiting"})
	status := "success"
	if err != nil {
		status = "error"
		metrics.PackageOpsCount.WithLabelValues("FindByWaitingStatus", status).Inc()
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
	metrics.PackageOpsCount.WithLabelValues("FindByWaitingStatus", status).Inc()
	return packages, nil
}
