package service

import (
	"context"

	"github.com/maksroxx/DeliveryService/database/internal/models"
)

type PackageService interface {
	GetPackageByID(ctx context.Context, packageID string) (*models.Package, error)
	GetAllPackages(ctx context.Context, filter models.PackageFilter) ([]*models.Package, error)
	CreatePackage(ctx context.Context, pkg *models.Package) (*models.Package, error)
	UpdatePackage(ctx context.Context, packageID string, update models.PackageUpdate) (*models.Package, error)
	DeletePackage(ctx context.Context, packageID string) error
	CancelPackage(ctx context.Context, packageID string) (*models.Package, error)
	GetExpiredPackages(ctx context.Context) ([]*models.Package, error)
	MarkPackageAsExpired(ctx context.Context, packageID string) (*models.Package, error)

	CreatePackageWithCalculation(ctx context.Context, req *models.Package) (*models.Package, error)
	TransferExpiredPackages(ctx context.Context) error
}
