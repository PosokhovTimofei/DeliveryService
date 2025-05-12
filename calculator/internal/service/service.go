package service

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/maksroxx/DeliveryService/calculator/internal/repository"
	"github.com/maksroxx/DeliveryService/calculator/models"
)

type Calculator interface {
	Calculate(ctx context.Context, rep repository.CountryRepository, pkg models.Package) (models.CalculationResult, error)
}

type DefaultCalculator struct {
	baseRate   float64
	pricePerKm float64
	pricePerKg float64
	currency   string
}

func NewCalculator() *DefaultCalculator {
	return &DefaultCalculator{
		baseRate:   500,
		pricePerKm: 30,
		pricePerKg: 50,
		currency:   "RUB",
	}
}

func (c *DefaultCalculator) Calculate(ctx context.Context, repo repository.CountryRepository, pkg models.Package) (models.CalculationResult, error) {
	from, err := repo.GetCoordinates(ctx, pkg.From)
	if err != nil {
		log.Printf("Failed to get coordinates for origin country '%s': %v", pkg.From, err)
		return fallbackResult(pkg), nil
	}

	to, err := repo.GetCoordinates(ctx, pkg.To)
	if err != nil {
		log.Printf("Failed to get coordinates for destination country '%s': %v", pkg.To, err)
		return fallbackResult(pkg), nil
	}

	distance := haversine(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
	baseCost := c.baseRate + (distance * c.pricePerKm) + (pkg.Weight * c.pricePerKg)
	cost := baseCost * timeMultiplier() * zoneMultiplier(distance)

	result := models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: int(math.Ceil(distance / 500 * 24)),
		Currency:       c.currency,
	}

	return result, nil
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

func fallbackResult(pkg models.Package) models.CalculationResult {
	base := 500.0
	weightFee := pkg.Weight * 50
	cost := base + weightFee

	return models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: 72,
		Currency:       "RUB",
	}
}
