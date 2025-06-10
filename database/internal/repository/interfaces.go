package repository

import (
	"context"

	"github.com/maksroxx/DeliveryService/database/internal/models"
)

type RouteRepository interface {
	GetByID(ctx context.Context, id string) (*models.Package, error)
	GetAllPackages(ctx context.Context, filter models.PackageFilter) ([]*models.Package, error)
	GetExpiredPackages(ctx context.Context) ([]*models.Package, error)
	MarkAsExpiredByID(ctx context.Context, packageID string) (*models.Package, error)
	Create(ctx context.Context, route *models.Package) (*models.Package, error)
	UpdatePackage(ctx context.Context, id string, update models.PackageUpdate) (*models.Package, error)
	DeletePackage(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}
