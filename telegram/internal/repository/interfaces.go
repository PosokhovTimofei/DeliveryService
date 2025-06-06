package repository

type Linker interface {
	SaveLink(telegramID int64, userID string) error
	GetUserIDByTelegramID(telegramID int64) (string, error)
}
