package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/database/internal/clients"
	"github.com/maksroxx/DeliveryService/database/internal/kafka"
	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"github.com/sirupsen/logrus"
)

type packageService struct {
	repo       repository.RouteRepository
	calculator clients.Calculator
	producer   kafka.PaymentProducer
	logger     *logrus.Logger
}

func NewPackageService(repo repository.RouteRepository, calculator clients.Calculator, producer kafka.PaymentProducer, log *logrus.Logger) *packageService {
	return &packageService{
		repo:       repo,
		calculator: calculator,
		producer:   producer,
		logger:     log,
	}
}

func (s *packageService) GetPackageByID(ctx context.Context, packageID string) (*models.Package, error) {
	return s.repo.GetByID(ctx, packageID)
}

func (s *packageService) GetAllPackages(ctx context.Context, filter models.PackageFilter) ([]*models.Package, error) {
	return s.repo.GetAllPackages(ctx, filter)
}

func (s *packageService) CreatePackage(ctx context.Context, pkg *models.Package) (*models.Package, error) {
	pkg.CreatedAt = time.Now()
	return s.repo.Create(ctx, pkg)
}

func (s *packageService) UpdatePackage(ctx context.Context, packageID string, update models.PackageUpdate) (*models.Package, error) {
	return s.repo.UpdatePackage(ctx, packageID, update)
}

func (s *packageService) DeletePackage(ctx context.Context, packageID string) error {
	return s.repo.DeletePackage(ctx, packageID)
}

func (s *packageService) CancelPackage(ctx context.Context, packageID string) (*models.Package, error) {
	pkg, err := s.repo.GetByID(ctx, packageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status == "Delivered" {
		return nil, errors.New("cannot cancel a delivered package")
	}
	if pkg.Status == "Сanceled" {
		return nil, errors.New("package already canceled")
	}
	update := models.PackageUpdate{
		Status: "Сanceled",
	}
	return s.repo.UpdatePackage(ctx, packageID, update)
}

func (s *packageService) GetExpiredPackages(ctx context.Context) ([]*models.Package, error) {
	return s.repo.GetExpiredPackages(ctx)
}

func (s *packageService) MarkPackageAsExpired(ctx context.Context, packageID string) (*models.Package, error) {
	return s.repo.MarkAsExpiredByID(ctx, packageID)
}

func (s *packageService) CreatePackageWithCalculation(ctx context.Context, pkg *models.Package) (*models.Package, error) {
	if pkg.Weight <= 0 || pkg.From == "" || pkg.To == "" || pkg.Address == "" || pkg.Length <= 0 || pkg.Width <= 0 || pkg.Height <= 0 {
		return nil, errors.New("invalid input")
	}

	var result *calculatorpb.CalculateDeliveryCostResponse
	var err error

	tariff := pkg.TariffCode
	if tariff == "" {
		result, err = s.calculator.Calculate(pkg.Weight, pkg.UserID, pkg.From, pkg.To, pkg.Address, pkg.Length, pkg.Width, pkg.Height)
		tariff = "DEFAULT"
	} else {
		result, err = s.calculator.CalculateByTariff(pkg.Weight, pkg.UserID, pkg.From, pkg.To, pkg.Address, tariff, pkg.Length, pkg.Width, pkg.Height)
	}
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	pkg.PackageID = "PKG-" + uuid.New().String()
	pkg.Status = "Created"
	pkg.PaymentStatus = "PENDING"
	pkg.Cost = result.Cost
	pkg.EstimatedHours = int(result.EstimatedHours)
	pkg.Currency = result.Currency
	pkg.CreatedAt = time.Now()
	pkg.TariffCode = tariff

	created, err := s.repo.Create(ctx, pkg)
	if err != nil {
		return nil, err
	}

	payment := models.Payment{
		UserID:    pkg.UserID,
		PackageID: pkg.PackageID,
		Cost:      pkg.Cost,
		Currency:  pkg.Currency,
	}

	if err := s.producer.SendPaymentEvent(payment); err != nil {
		return nil, fmt.Errorf("failed to send payment event: %w", err)
	}

	return created, nil
}

func (s *packageService) TransferExpiredPackages(ctx context.Context) error {
	expired, err := s.repo.GetExpiredPackages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired packages: %w", err)
	}

	for _, pkg := range expired {
		if err := s.producer.SendExpiredPackageEvent(*pkg); err != nil {
			s.logger.WithError(err).Errorf("failed to send expired event for %s", pkg.PackageID)
			continue
		}

		if err := s.repo.DeletePackage(ctx, pkg.PackageID); err != nil {
			s.logger.WithError(err).Errorf("failed to delete expired package %s", pkg.PackageID)
		}
	}

	return nil
}
