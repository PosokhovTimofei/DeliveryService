package service

import (
	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/producer/internal/delivery/kafka"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type PackageService struct {
	producer *kafka.Producer
}

func NewPackageService(producer *kafka.Producer) *PackageService {
	return &PackageService{producer: producer}
}

func (s *PackageService) CreatePackage(pkg pkg.Package) (*pkg.Package, error) {
	pkg.ID = uuid.New().String()
	pkg.Status = "CREATED"

	if err := s.producer.SendPackage(pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}
