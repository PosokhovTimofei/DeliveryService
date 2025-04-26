package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/producer/internal/calculator"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/internal/repository"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type PackageService struct {
	repo             repository.Packager
	producer         *kafka.Producer
	calculatorClient *calculator.Client
}

func NewPackageService(producer *kafka.Producer, client *calculator.Client, rep repository.Packager) *PackageService {
	return &PackageService{producer: producer, calculatorClient: client, repo: rep}
}

func (s *PackageService) CreatePackage(ctx context.Context, pack pkg.Package, userID string) (*pkg.Package, error) {
	pack.UserID = userID
	pack.ID = "PKG-" + uuid.New().String()
	pack.Status = "CREATED"

	exists, err := s.repo.PackageExists(ctx, pack)
	if err != nil {
		return nil, fmt.Errorf("failed to check package existence: %w", err)
	}
	if exists {
		return nil, errors.New("package already exists for this user")
	}

	result, err := s.calculatorClient.Calculate(pack)
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	pack.Cost = result.Cost
	pack.EstimatedHours = result.EstimatedHours
	pack.Currency = result.Currency
	pack.Status = "PROCESSED"

	if err := s.repo.CreatePackage(ctx, pack); err != nil {
		return nil, fmt.Errorf("failed to save package: %w", err)
	}

	if err := s.producer.SendPackage(pack, userID); err != nil {
		return nil, fmt.Errorf("failed to send package: %w", err)
	}

	paymentEvent := pkg.PaymentEvent{
		UserID:    userID,
		PackageID: pack.ID,
		Cost:      pack.Cost,
		Currency:  pack.Currency,
	}

	if err := s.producer.SendPaymentEvent(paymentEvent); err != nil {
		return nil, fmt.Errorf("failed to send payment event: %w", err)
	}

	return &pack, nil
}
