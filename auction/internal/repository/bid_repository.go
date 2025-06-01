package repository

import (
	"context"
	"time"

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
	bid.BidID = bid.PackageID + "-" + bid.UserID + "-" + time.Now().Format("150405")
	bid.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, bid)
	return err
}

func (r *BidRepository) GetBidsByPackage(ctx context.Context, packageID string) ([]*models.Bid, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"package_id": packageID})
	if err != nil {
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
	return bids, nil
}

func (r *BidRepository) GetTopBidByPackage(ctx context.Context, packageID string) (*models.Bid, error) {
	filter := bson.M{"package_id": packageID}
	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})

	var topBid models.Bid
	err := r.collection.FindOne(ctx, filter, opts).Decode(&topBid)
	if err != nil {
		return nil, err
	}
	return &topBid, nil
}
