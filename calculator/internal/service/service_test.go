package service

// import (
// 	"testing"

// 	"github.com/maksroxx/DeliveryService/calculator/models"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCalculate_SameLocation(t *testing.T) {
// 	calculator := NewCalculator()

// 	pkg := models.Package{
// 		From:   "Moscow",
// 		To:     "Moscow",
// 		Weight: 10,
// 	}

// 	result, err := calculator.Calculate(pkg)
// 	assert.NoError(t, err)
// 	assert.Greater(t, result.Cost, 0.0)
// 	assert.Equal(t, "RUB", result.Currency)
// 	assert.LessOrEqual(t, result.EstimatedHours, 24)
// }

// func TestCalculate_ShortDistance(t *testing.T) {
// 	calculator := NewCalculator()

// 	pkg := models.Package{
// 		From:   "Moscow",
// 		To:     "Tula",
// 		Weight: 5,
// 	}

// 	result, err := calculator.Calculate(pkg)
// 	assert.NoError(t, err)
// 	assert.Greater(t, result.Cost, 0.0)
// 	assert.True(t, result.EstimatedHours >= 1 && result.EstimatedHours <= 48)
// }

// func TestCalculate_LongDistance(t *testing.T) {
// 	calculator := NewCalculator()

// 	pkg := models.Package{
// 		From:   "Moscow",
// 		To:     "Vladivostok",
// 		Weight: 25,
// 	}

// 	result, err := calculator.Calculate(pkg)
// 	assert.NoError(t, err)
// 	assert.Greater(t, result.Cost, 0.0)
// 	assert.GreaterOrEqual(t, result.EstimatedHours, 1)
// }

// func TestCalculate_NightTimeMultiplier(t *testing.T) {
// 	calculator := NewCalculator()

// 	pkg := models.Package{
// 		From:   "Kazan",
// 		To:     "Samara",
// 		Weight: 15,
// 	}

// 	result1, err1 := calculator.Calculate(pkg)
// 	result2, err2 := calculator.Calculate(pkg)

// 	assert.NoError(t, err1)
// 	assert.NoError(t, err2)
// 	assert.Greater(t, result1.Cost, 0.0)
// 	assert.Equal(t, result1.Currency, "RUB")
// 	assert.Equal(t, result1.EstimatedHours, result2.EstimatedHours)
// }
