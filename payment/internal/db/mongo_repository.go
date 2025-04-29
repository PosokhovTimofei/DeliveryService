package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/maksroxx/DeliveryService/payment/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PaymentMongoRepository struct {
	collection *mongo.Collection
}

func NewPaymentMongoRepository(db *mongo.Database, collectionName string) *PaymentMongoRepository {
	collection := db.Collection(collectionName)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "package_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create index: %v", err))
	}
	return &PaymentMongoRepository{collection: collection}
}

func (r *PaymentMongoRepository) CreatePayment(ctx context.Context, payment models.Payment) error {
	now := time.Now()
	doc := bson.M{
		"user_id":    payment.UserID,
		"package_id": payment.PackageID,
		"cost":       payment.Cost,
		"currency":   payment.Currency,
		"status":     models.PaymentStatusPending,
		"created_at": now,
		"updated_at": now,
	}
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("payment has already exists")
		}
		return err
	}
	return nil
}

func (r *PaymentMongoRepository) UpdatePayment(ctx context.Context, update models.Payment) (*models.Payment, error) {
	filter := bson.M{
		"user_id":    update.UserID,
		"package_id": update.PackageID,
		"status":     bson.M{"$ne": "PAID"},
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"status":     update.Status,
			"updated_at": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedPayment models.Payment
	err := r.collection.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedPayment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("payment already confirmed")
		}
		return nil, err
	}
	return &updatedPayment, nil
}
