package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/maksroxx/DeliveryService/database/internal/metrics"
	"github.com/maksroxx/DeliveryService/database/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database, collectionName string) *MongoRepository {
	collection := db.Collection(collectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "package_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create unique index: %v", err))
	}

	return &MongoRepository{
		collection: collection,
	}
}

func (r *MongoRepository) GetByID(ctx context.Context, packageID string) (*models.Package, error) {
	filter := bson.M{"package_id": packageID}

	var route models.Package
	err := r.collection.FindOne(ctx, filter).Decode(&route)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("route not found")
		}
		return nil, err
	}
	route.RemainingHours = r.calculateRemainingHours(ctx, &route)
	return &route, nil
}

func (r *MongoRepository) Create(ctx context.Context, route *models.Package) (*models.Package, error) {
	if route.UserID == "" {
		return nil, errors.New("user ID is required")
	}
	if route.PackageID == "" {
		return nil, errors.New("packageID is required")
	}
	alreadyCreated, err := r.alreadyCreatedToday(ctx, route)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if alreadyCreated {
		metrics.FailedPackageCreations.Inc()
		return nil, errors.New("limit: only 3 identical packages allowed per day")
	}

	now := time.Now()
	doc := bson.M{
		"user_id":         route.UserID,
		"package_id":      route.PackageID,
		"weight":          route.Weight,
		"length":          route.Length,
		"width":           route.Width,
		"height":          route.Height,
		"from":            route.From,
		"to":              route.To,
		"address":         route.Address,
		"payment_status":  "PENDING",
		"status":          "Created",
		"cost":            route.Cost,
		"estimated_hours": route.EstimatedHours,
		"currency":        route.Currency,
		"created_at":      route.CreatedAt,
		"updated_at":      now,
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		metrics.FailedPackageCreations.Inc()
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("package has already exists")
		}
		return nil, fmt.Errorf("failed to create package: %w", err)
	}

	metrics.CreatedPackages.Inc()

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		route.ID = oid.Hex()
	} else {
		return nil, errors.New("failed to get generated ID")
	}

	return route, nil
}

func (r *MongoRepository) GetAllPackages(ctx context.Context, filter models.PackageFilter) ([]*models.Package, error) {
	bsonFilter := bson.M{}

	if filter.UserID != "" {
		bsonFilter["user_id"] = filter.UserID
	}

	if filter.Status != "" {
		bsonFilter["status"] = filter.Status
	}
	if !filter.CreatedAfter.IsZero() {
		bsonFilter["created_at"] = bson.M{"$gte": filter.CreatedAfter}
	}

	opts := options.Find()
	if filter.Limit > 0 {
		opts.SetLimit(filter.Limit)
		opts.SetSkip(filter.Offset)
	}

	cur, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var routes []*models.Package
	for cur.Next(ctx) {
		var route models.Package
		if err := cur.Decode(&route); err != nil {
			return nil, err
		}
		route.RemainingHours = r.calculateRemainingHours(ctx, &route)
		routes = append(routes, &route)
	}
	return routes, nil
}

func (r *MongoRepository) UpdatePackage(ctx context.Context, packageID string, update models.PackageUpdate) (*models.Package, error) {
	filter := bson.M{"package_id": packageID}

	setFields := bson.M{}
	if update.Status != "" {
		setFields["status"] = update.Status
	}
	if update.PaymentStatus != "" {
		setFields["payment_status"] = update.PaymentStatus
	}
	setFields["updated_at"] = time.Now()

	updateDoc := bson.M{
		"$set": setFields,
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedRoute models.Package
	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		updateDoc,
		opts,
	).Decode(&updatedRoute)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("route with packageID %s not found", packageID)
		}
		return nil, err
	}

	metrics.UpdatedPackages.Inc()
	return &updatedRoute, nil
}

func (r *MongoRepository) DeletePackage(ctx context.Context, packageID string) error {
	filter := bson.M{"package_id": packageID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("route with packageID %s not found", packageID)
	}

	return nil
}

func (r *MongoRepository) GetExpiredPackages(ctx context.Context) ([]*models.Package, error) {
	cutoff := time.Now().AddDate(0, 0, -60)

	filter := bson.M{
		"status":     "In pick-up point",
		"updated_at": bson.M{"$lte": cutoff},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var packages []*models.Package
	for cursor.Next(ctx) {
		var pkg models.Package
		if err := cursor.Decode(&pkg); err != nil {
			return nil, err
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}

func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.collection.Database().Client().Ping(ctx, nil)
}

func (r *MongoRepository) calculateRemainingHours(ctx context.Context, route *models.Package) int {
	if route.Status == "Сanceled" || route.Status == "Delivered" {
		return 0
	}

	elapsed := int(time.Since(route.CreatedAt).Hours())
	remaining := route.EstimatedHours - elapsed
	if remaining < 0 {
		remaining = 0
	}

	if remaining == 0 && route.Status != "Delivered" && route.Status != "Сanceled" {
		filter := bson.M{"package_id": route.PackageID}
		update := bson.M{
			"$set": bson.M{
				"status":     "Delivered",
				"updated_at": time.Now(),
			},
		}
		_, err := r.collection.UpdateOne(ctx, filter, update)
		if err == nil {
			route.Status = "Delivered"
			metrics.DeliveredPackagesTotal.Inc()

			duration := time.Since(route.CreatedAt).Seconds()
			metrics.PackageDeliveryDuration.Observe(duration)
		}
	}

	return remaining
}

func (r *MongoRepository) alreadyCreatedToday(ctx context.Context, route *models.Package) (bool, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"user_id": route.UserID,
		"from":    route.From,
		"to":      route.To,
		"address": route.Address,
		"weight":  route.Weight,
		"length":  route.Length,
		"width":   route.Width,
		"height":  route.Height,
		"created_at": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count >= 3, nil
}
