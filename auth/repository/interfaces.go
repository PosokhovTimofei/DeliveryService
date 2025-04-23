package repository

import (
	"context"

	"github.com/maksroxx/DeliveryService/auth/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, updateFields map[string]any) error
	DeleteUser(ctx context.Context, userID string) error
}
