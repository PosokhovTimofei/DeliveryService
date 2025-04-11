package service

import (
	"hash/fnv"
	"math"
	"math/rand"
	"strings"
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
		currency:   "RUB",
	}
}

func calculateDistance(from, to string) float64 {
	from = strings.ToLower(strings.TrimSpace(from))
	to = strings.ToLower(strings.TrimSpace(to))

	if from == to {
		return 0
	}

	hash := fnv.New32a()
	hash.Write([]byte(from + "-" + to))
	seed := int64(hash.Sum32())

	r := rand.New(rand.NewSource(seed))
	switch r.Intn(10) {
	case 0, 1, 2, 3, 4, 5:
		return 50 + r.Float64()*150
	case 6, 7, 8:
		return 500 + r.Float64()*400
	default:
		return 1200 + r.Float64()*5000
	}
}

func (c *DefaultCalculator) Calculate(pkg models.Package) (models.CalculationResult, error) {
	distance := calculateDistance(pkg.From, pkg.To)

	baseCost := c.baseRate + (distance * c.pricePerKm) + (pkg.Weight * c.pricePerKg)
	if isNightTime() {
		baseCost *= 1.15
	}
	days := distance / 500

	estimatedHours := int(math.Ceil(days * 24))

	return models.CalculationResult{
		Cost:           math.Round(baseCost*100) / 100,
		EstimatedHours: estimatedHours,
		Currency:       c.currency,
	}, nil
}

func isNightTime() bool {
	now := time.Now().Hour()
	return now >= 22 || now < 6
}
