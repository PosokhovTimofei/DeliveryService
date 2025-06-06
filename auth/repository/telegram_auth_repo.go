package repository

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/auth/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TelegramAuthRepo struct {
	collection *mongo.Collection
}

func NewTelegramAuthRepo(db *mongo.Database, collectionName string) *TelegramAuthRepo {
	return &TelegramAuthRepo{
		collection: db.Collection("telegram_auth_codes"),
	}
}

func (r *TelegramAuthRepo) Save(code string, userID string, ttl time.Duration) error {
	doc := models.TelegramAuthCode{
		Code:     code,
		UserID:   userID,
		ExpireAt: time.Now().Add(ttl),
	}
	_, err := r.collection.InsertOne(context.TODO(), doc)
	return err
}

func (r *TelegramAuthRepo) FindUserIDByCode(code string) (string, error) {
	var result models.TelegramAuthCode
	err := r.collection.FindOne(context.TODO(), bson.M{
		"code":      code,
		"expire_at": bson.M{"$gt": time.Now()},
	}).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.UserID, nil
}
