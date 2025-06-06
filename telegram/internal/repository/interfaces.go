package repository

import "context"

type Linker interface {
	SaveLink(ctx context.Context, telegramID int64, userID string) error
	GetUserIDByTelegramID(ctx context.Context, telegramID int64) (string, error)
	GetTelegramIDByUserID(ctx context.Context, userId string) (int64, error)
}
