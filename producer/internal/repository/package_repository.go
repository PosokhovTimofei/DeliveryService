package repository

import (
	"context"

	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type Packager interface {
	CreatePackage(ctx context.Context, pkg pkg.Package) error
	PackageExists(ctx context.Context, pkg pkg.Package) (bool, error)
}
