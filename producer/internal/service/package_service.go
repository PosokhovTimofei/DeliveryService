package service

import (
	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/producer/internal/calculator"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type PackageService struct {
	producer         *kafka.Producer
	calculatorClient *calculator.Client
}

func NewPackageService(producer *kafka.Producer, client *calculator.Client) *PackageService {
	return &PackageService{producer: producer, calculatorClient: client}
}

func (s *PackageService) CreatePackage(pkg pkg.Package) (*pkg.Package, error) {
	pkg.ID = "PKG-" + uuid.New().String()
	pkg.Status = "CREATED"

	result, err := s.calculatorClient.Calculate(pkg)
	if err != nil {
		return nil, err
	}

	pkg.Cost = result.Cost
	pkg.EstimatedHours = result.EstimatedHours
	pkg.Currency = result.Currency
	pkg.Status = "PROCESSED"

	if err := s.producer.SendPackage(pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}
