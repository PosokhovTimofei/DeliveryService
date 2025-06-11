package repository

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/metrics"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BidRepository struct {
	collection *mongo.Collection
}

func NewBidRepository(db *mongo.Database, collectionName string) *BidRepository {
	return &BidRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *BidRepository) PlaceBid(ctx context.Context, bid *models.Bid) error {
	start := time.Now()
	defer func() {
		metrics.BidOpsDuration.WithLabelValues("PlaceBid").Observe(time.Since(start).Seconds())
	}()

	bid.BidID = bid.PackageID + "-" + bid.UserID + "-" + time.Now().Format("150405")
	bid.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, bid)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.BidOpsCount.WithLabelValues("PlaceBid", status).Inc()

	return err
}

func (r *BidRepository) GetBidsByPackage(ctx context.Context, packageID string) ([]*models.Bid, error) {
	start := time.Now()
	defer func() {
		metrics.BidOpsDuration.WithLabelValues("GetBidsByPackage").Observe(time.Since(start).Seconds())
	}()

	cursor, err := r.collection.Find(ctx, bson.M{"package_id": packageID})
	status := "success"
	if err != nil {
		status = "error"
		metrics.BidOpsCount.WithLabelValues("GetBidsByPackage", status).Inc()
		return nil, err
	}
	defer cursor.Close(ctx)

	var bids []*models.Bid
	for cursor.Next(ctx) {
		var bid models.Bid
		if err := cursor.Decode(&bid); err != nil {
			continue
		}
		bids = append(bids, &bid)
	}
	metrics.BidOpsCount.WithLabelValues("GetBidsByPackage", status).Inc()
	return bids, nil
}

func (r *BidRepository) WatchBidsByPackage(ctx context.Context, packageID string) (*mongo.ChangeStream, error) {
	start := time.Now()
	defer func() {
		metrics.BidOpsDuration.WithLabelValues("WatchBidsByPackage").Observe(time.Since(start).Seconds())
	}()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "fullDocument.package_id", Value: packageID},
			{Key: "operationType", Value: "insert"},
		}}},
	}
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	stream, err := r.collection.Watch(ctx, pipeline, opts)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.BidOpsCount.WithLabelValues("WatchBidsByPackage", status).Inc()
	return stream, err
}

func (r *BidRepository) GetTopBidByPackage(ctx context.Context, packageID string) (*models.Bid, error) {
	start := time.Now()
	defer func() {
		metrics.BidOpsDuration.WithLabelValues("GetTopBidByPackage").Observe(time.Since(start).Seconds())
	}()

	filter := bson.M{"package_id": packageID}
	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})
	var topBid models.Bid
	err := r.collection.FindOne(ctx, filter, opts).Decode(&topBid)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.BidOpsCount.WithLabelValues("GetTopBidByPackage", status).Inc()
	return &topBid, err
}
