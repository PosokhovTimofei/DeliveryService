package service

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/kafka"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type PaymentService interface {
	ConfirmPayment(ctx context.Context, userID, packageID string) (*models.Payment, error)
	ConfirmAuctionPayment(ctx context.Context, userID, packageID string) (*models.Payment, error)
}

type paymentService struct {
	repo     db.Paymenter
	producer kafka.Producerer
}

func NewPaymentService(repo db.Paymenter, producer kafka.Producerer) PaymentService {
	return &paymentService{repo: repo, producer: producer}
}

func (s *paymentService) ConfirmPayment(ctx context.Context, userID, packageID string) (*models.Payment, error) {
	payment, err := s.repo.UpdatePayment(ctx, models.Payment{
		UserID:    userID,
		PackageID: packageID,
		Status:    models.PaymentStatusPaid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	if err := s.producer.PaymentMessage(*payment, userID); err != nil {
		return nil, fmt.Errorf("failed to send payment message: %w", err)
	}

	return payment, nil
}

func (s *paymentService) ConfirmAuctionPayment(ctx context.Context, userID, packageID string) (*models.Payment, error) {
	payment, err := s.repo.UpdatePayment(ctx, models.Payment{
		UserID:    userID,
		PackageID: packageID,
		Status:    models.PaymentStatusPaid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update auction payment: %w", err)
	}

	if err := s.producer.PaymentAucitonMessage(*payment, userID); err != nil {
		return nil, fmt.Errorf("failed to send auction payment message: %w", err)
	}

	return payment, nil
}
