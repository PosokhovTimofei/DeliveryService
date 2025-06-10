package service

import (
	"context"
	"math"
	"time"

	"github.com/maksroxx/DeliveryService/calculator/internal/repository"
	"github.com/maksroxx/DeliveryService/calculator/models"
	"github.com/sirupsen/logrus"
)

type Calculator interface {
	Calculate(ctx context.Context, pkg models.Package) (models.CalculationResult, error)
	CalculateByTariffCode(ctx context.Context, pkg models.Package, code string) (models.CalculationResult, error)
	GetTariffs(ctx context.Context) ([]models.Tariff, error)
	CreateTariff(ctx context.Context, tariff *models.Tariff) (*models.Tariff, error)
	DeleteTariff(ctx context.Context, code string) error
}

type DefaultCalculator struct {
	defaultTariff models.Tariff
	repository    repository.CountryRepository
}

func NewCalculator(rep repository.CountryRepository) *DefaultCalculator {
	return &DefaultCalculator{
		defaultTariff: models.Tariff{
			Code:              "DEFAULT",
			Name:              "Default",
			BaseRate:          300,
			PricePerKm:        5,
			PricePerKg:        50,
			Currency:          "RUB",
			VolumetricDivider: 5000,
			SpeedKmph:         60,
		},
		repository: rep,
	}
}

func (c *DefaultCalculator) Calculate(ctx context.Context, pkg models.Package) (models.CalculationResult, error) {
	from, err := c.repository.GetCoordinates(ctx, pkg.From)
	if err != nil {
		logrus.Printf("Failed to get coordinates for origin country '%s': %v", pkg.From, err)
		return fallbackResult(pkg, c.defaultTariff.Currency), nil
	}

	to, err := c.repository.GetCoordinates(ctx, pkg.To)
	if err != nil {
		logrus.Printf("Failed to get coordinates for destination country '%s': %v", pkg.To, err)
		return fallbackResult(pkg, c.defaultTariff.Currency), nil
	}

	distance := haversine(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
	volumetricWeight := float64(pkg.Length*pkg.Width*pkg.Height) / c.defaultTariff.VolumetricDivider
	effectiveWeight := math.Max(pkg.Weight, volumetricWeight)
	cost := c.defaultTariff.BaseRate +
		distance*c.defaultTariff.PricePerKm +
		effectiveWeight*c.defaultTariff.PricePerKg
	cost *= timeMultiplier() * zoneMultiplier(distance)
	speed := c.defaultTariff.SpeedKmph
	if speed <= 0 {
		speed = 50
	}
	estimatedHours := int(math.Ceil(distance / float64(speed) * timeDelayMultiplier(distance)))
	if estimatedHours < 6 {
		estimatedHours = 6
	}

	return models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: estimatedHours,
		Currency:       c.defaultTariff.Currency,
	}, nil
}

type ExtendedCalculator struct {
	DefaultCalculator
	tariffRepo repository.TariffRepository
}

func NewExtendedCalculator(countryRepo repository.CountryRepository, tariffRepo repository.TariffRepository) *ExtendedCalculator {
	return &ExtendedCalculator{
		DefaultCalculator: *NewCalculator(countryRepo),
		tariffRepo:        tariffRepo,
	}
}

func (c *ExtendedCalculator) CalculateByTariffCode(ctx context.Context, pkg models.Package, code string) (models.CalculationResult, error) {
	tariff, err := c.tariffRepo.GetByCode(ctx, code)
	if err != nil {
		return fallbackResult(pkg, c.defaultTariff.Currency), nil
	}

	from, err := c.repository.GetCoordinates(ctx, pkg.From)
	if err != nil {
		return fallbackResult(pkg, tariff.Currency), nil
	}

	to, err := c.repository.GetCoordinates(ctx, pkg.To)
	if err != nil {
		return fallbackResult(pkg, tariff.Currency), nil
	}

	distance := haversine(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
	logrus.Printf("Distance '%s': %0.2f", pkg.To, distance)
	volumetricWeight := float64(pkg.Length*pkg.Width*pkg.Height) / tariff.VolumetricDivider
	effectiveWeight := math.Max(pkg.Weight, volumetricWeight)
	cost := tariff.BaseRate +
		distance*tariff.PricePerKm +
		effectiveWeight*tariff.PricePerKg
	cost *= timeMultiplier() * zoneMultiplier(distance)
	speed := tariff.SpeedKmph
	if speed <= 0 {
		speed = 50
	}
	estimatedHours := int(math.Ceil(distance / float64(speed) * timeDelayMultiplier(distance)))
	if estimatedHours < 6 {
		estimatedHours = 6
	}

	return models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: estimatedHours,
		Currency:       tariff.Currency,
	}, nil
}

func (c *ExtendedCalculator) GetTariffs(ctx context.Context) ([]models.Tariff, error) {
	return c.tariffRepo.GetAll(ctx)
}

func (c *ExtendedCalculator) CreateTariff(ctx context.Context, tariff *models.Tariff) (*models.Tariff, error) {
	return c.tariffRepo.CreateTariff(ctx, tariff)
}

func (c *ExtendedCalculator) DeleteTariff(ctx context.Context, code string) error {
	return c.tariffRepo.DeleteTariff(ctx, code)
}

func fallbackResult(pkg models.Package, currency string) models.CalculationResult {
	volumetricWeight := float64(pkg.Length*pkg.Width*pkg.Height) / 5000
	effectiveWeight := math.Max(pkg.Weight, volumetricWeight)
	base := 500.0
	weightFee := effectiveWeight * 50
	cost := base + weightFee

	return models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: 72,
		Currency:       currency,
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func timeMultiplier() float64 {
	h := time.Now().Hour()
	m := 1.0
	if h >= 22 || h < 6 {
		m += 0.15
	}
	weekday := time.Now().Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		m += 0.2
	}
	return m
}

func zoneMultiplier(distance float64) float64 {
	switch {
	case distance < 100:
		return 1.0
	case distance < 500:
		return 1.15
	case distance < 1500:
		return 1.3
	default:
		return 1.5
	}
}

func timeDelayMultiplier(distance float64) float64 {
	switch {
	case distance < 100:
		return 1.2
	case distance < 500:
		return 1.5
	case distance < 1500:
		return 1.75
	default:
		return 2.0
	}
}
