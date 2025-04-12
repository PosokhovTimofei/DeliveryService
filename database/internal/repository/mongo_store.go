package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	"go.mongodb.org/mongo-driver/bson"
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
		collection: db.Collection(collectionName),
	}
}

func (r *MongoRepository) GetByID(ctx context.Context, packageID string) (*models.Route, error) {
	filter := bson.M{"package_id": packageID}

	var route models.Route
	err := r.collection.FindOne(ctx, filter).Decode(&route)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("route not found")
		}
		return nil, err
	}

	return &route, nil
}

func (r *MongoRepository) Create(ctx context.Context, route *models.Route) (*models.Route, error) {
	if route.PackageID == "" {
		return nil, errors.New("packageID is required")
	}

	existing, _ := r.GetByID(ctx, route.PackageID)
	if existing != nil {
		return nil, errors.New("route with this packageID already exists")
	}

	now := time.Now()

	doc := bson.M{
		"package_id":      route.PackageID,
		"weight":          route.Weight,
		"from":            route.From,
		"to":              route.To,
		"address":         route.Address,
		"status":          route.Status,
		"cost":            route.Cost,
		"estimated_hours": route.EstimatedHours,
		"currency":        route.Currency,
		"created_at":      now,
		"updated_at":      now,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("route with this packageID already exists")
		}
		return nil, fmt.Errorf("failed to create package: %w", err)
	}

	return route, nil
}

func (r *MongoRepository) GetAllRoutes(ctx context.Context, filter models.RouteFilter) ([]*models.Route, error) {
	bsonFilter := bson.M{}

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

	var routes []*models.Route
	for cur.Next(ctx) {
		var route models.Route
		if err := cur.Decode(&route); err != nil {
			return nil, err
		}
		routes = append(routes, &route)
	}

	return routes, nil
}

func (r *MongoRepository) UpdateRoute(ctx context.Context, packageID string, update models.RouteUpdate) (*models.Route, error) {
	filter := bson.M{"package_id": packageID}

	updateDoc := bson.M{
		"$set": bson.M{
			"status":     update.Status,
			"updated_at": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedRoute models.Route
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

	return &updatedRoute, nil
}

func (r *MongoRepository) DeleteRoute(ctx context.Context, packageID string) error {
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

func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.collection.Database().Client().Ping(ctx, nil)
}
