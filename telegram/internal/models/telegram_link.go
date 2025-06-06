package models

import "time"

type TelegramUserLink struct {
	TelegramID int64     `bson:"telegram_id"`
	UserID     string    `bson:"user_id"`
	LinkedAt   time.Time `bson:"linked_at"`
}
