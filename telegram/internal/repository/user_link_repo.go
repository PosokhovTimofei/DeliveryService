package repository

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/telegram/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserLinkRepository struct {
	collection *mongo.Collection
}

func NewUserLinkRepository(db *mongo.Database, collectionName string) *UserLinkRepository {
	return &UserLinkRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *UserLinkRepository) SaveLink(telegramID int64, userID string) error {
	filter := bson.M{"telegram_id": telegramID}
	update := bson.M{
		"$set": models.TelegramUserLink{
			TelegramID: telegramID,
			UserID:     userID,
			LinkedAt:   time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (r *UserLinkRepository) GetUserIDByTelegramID(telegramID int64) (string, error) {
	var link models.TelegramUserLink
	err := r.collection.FindOne(context.Background(), bson.M{"telegram_id": telegramID}).Decode(&link)
	if err != nil {
		return "", err
	}
	return link.UserID, nil
}
