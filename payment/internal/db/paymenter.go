package db

import (
	"context"

	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type Paymenter interface {
	CreatePayment(ctx context.Context, payment models.Payment) error
	UpdatePayment(ctx context.Context, update models.Payment) (*models.Payment, error)
}
