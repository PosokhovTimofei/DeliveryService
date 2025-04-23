package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/maksroxx/DeliveryService/auth/models"
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
		Keys:    bson.D{{Key: "user_id", Value: 1}},
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

func (r *MongoRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.Role = "user"
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user has already exists")
		}
		return err
	}
	return nil
}

func (r *MongoRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	filter := bson.M{"user_id": userID}

	var user models.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	filter := bson.M{"email": email}

	var user models.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepository) UpdateUser(ctx context.Context, userID string, updateFields map[string]any) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": updateFields}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *MongoRepository) DeleteUser(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
