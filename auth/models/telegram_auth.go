package models

import "time"

type TelegramAuthCode struct {
	Code     string    `bson:"code"`
	UserID   string    `bson:"user_id"`
	ExpireAt time.Time `bson:"expire_at"`
}
