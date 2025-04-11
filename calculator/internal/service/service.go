package service

import (
	"math"
	"math/rand"
	"time"

	"github.com/maksroxx/DeliveryService/calculator/models"
)

type Calculator interface {
	Calculate(pkg models.Package) (models.CalculationResult, error)
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
		currency:   "USD",
	}
}

func (c *DefaultCalculator) Calculate(pkg models.Package) (models.CalculationResult, error) {
	distance := calulateDistance(pkg.From, pkg.To)
	cost := c.baseRate + (distance * c.pricePerKm) + (pkg.Weight * c.pricePerKg)
	return models.CalculationResult{
		Cost:           math.Round(cost*100) / 100,
		EstimatedHours: int(distance / 50 * 1.5),
		Currency:       c.currency,
	}, nil
}

func calulateDistance(from, to string) float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := r.Intn(100) + 1
	return float64(randomNumber)
}
