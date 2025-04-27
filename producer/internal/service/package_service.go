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
	"github.com/sirupsen/logrus"
)

type PackageService struct {
	repo             repository.Packager
	producer         *kafka.Producer
	calculatorClient *calculator.CalculatorGRPCClient
	logger           *logrus.Logger
}

func NewPackageService(producer *kafka.Producer, client *calculator.CalculatorGRPCClient, repo repository.Packager, logger *logrus.Logger) *PackageService {
	return &PackageService{
		producer:         producer,
		calculatorClient: client,
		repo:             repo,
		logger:           logger,
	}
}

func (s *PackageService) CreatePackage(ctx context.Context, pack pkg.Package, userID string) (*pkg.Package, error) {
	pack.UserID = userID
	pack.ID = "PKG-" + uuid.New().String()
	pack.Status = "CREATED"

	result, err := s.calculatorClient.Calculate(pack.Weight, userID, pack.From, pack.To, pack.Address)
	if err != nil {
		s.logger.WithError(err).Error("Cost calculation failed")
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	pack.Cost = result.Cost
	pack.EstimatedHours = int(result.EstimatedHours)
	pack.Currency = result.Currency
	pack.Status = "PROCESSED"

	exists, err := s.repo.PackageExists(ctx, pack)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check package existence")
		return nil, fmt.Errorf("failed to check package existence: %w", err)
	}
	if exists {
		s.logger.Warnf("Package already exists for user: %s", userID)
		return nil, errors.New("package already exists for this user")
	}

	if err := s.repo.CreatePackage(ctx, pack); err != nil {
		s.logger.WithError(err).Error("Failed to save package into repository")
		return nil, fmt.Errorf("failed to save package: %w", err)
	}

	if err := s.producer.BeginTransaction(); err != nil {
		s.logger.WithError(err).Error("Failed to begin Kafka transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = s.producer.AbortTransaction()
			s.logger.Warn("Transaction aborted due to an error")
		}
	}()

	if err = s.producer.SendPackage(pack, userID); err != nil {
		s.logger.WithError(err).Error("Failed to send package to Kafka")
		return nil, fmt.Errorf("failed to send package: %w", err)
	}

	paymentEvent := pkg.PaymentEvent{
		UserID:    userID,
		PackageID: pack.ID,
		Cost:      pack.Cost,
		Currency:  pack.Currency,
	}

	if err = s.producer.SendPaymentEvent(paymentEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send payment event to Kafka")
		return nil, fmt.Errorf("failed to send payment event: %w", err)
	}

	if err = s.producer.CommitTransaction(); err != nil {
		s.logger.WithError(err).Error("Failed to commit Kafka transaction")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pack, nil
}
