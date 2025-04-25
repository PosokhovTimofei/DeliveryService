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

func (s *PackageService) CreatePackage(ctx context.Context, pkg pkg.Package, userID string) (*pkg.Package, error) {
	pkg.UserID = userID
	pkg.ID = "PKG-" + uuid.New().String()
	pkg.Status = "CREATED"

	exists, err := s.repo.PackageExists(ctx, pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to check package existence: %w", err)
	}
	if exists {
		return nil, errors.New("package already exists for this user")
	}

	result, err := s.calculatorClient.Calculate(pkg)
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	pkg.Cost = result.Cost
	pkg.EstimatedHours = result.EstimatedHours
	pkg.Currency = result.Currency
	pkg.Status = "PROCESSED"

	if err := s.repo.CreatePackage(ctx, pkg); err != nil {
		return nil, fmt.Errorf("failed to save package: %w", err)
	}

	if err := s.producer.SendPackage(pkg, userID); err != nil {
		return nil, fmt.Errorf("failed to send package: %w", err)
	}

	return &pkg, nil
}
