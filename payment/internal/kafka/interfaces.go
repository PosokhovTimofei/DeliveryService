package kafka

import (
	"context"

	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type Consumerer interface {
	Run(ctx context.Context)
	Close() error
}

type Producerer interface {
	PaymentMessage(payment models.Payment, userID string) error
	PaymentAucitonMessage(payment models.Payment, userID string) error
	Close() error
}
