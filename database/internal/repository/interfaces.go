package repository

import (
	"context"

	"github.com/maksroxx/DeliveryService/database/internal/models"
)

type RouteRepository interface {
	GetByID(ctx context.Context, id string) (*models.Route, error)
	GetAllRoutes(ctx context.Context, filter models.RouteFilter) ([]*models.Route, error)
	Create(ctx context.Context, route *models.Route) (*models.Route, error)
	UpdateRoute(ctx context.Context, id string, update models.RouteUpdate) (*models.Route, error)
	DeleteRoute(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}
