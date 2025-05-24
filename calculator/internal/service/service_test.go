package service_test

import (
	"context"
	"testing"

	"github.com/maksroxx/DeliveryService/calculator/internal/service"
	"github.com/maksroxx/DeliveryService/calculator/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCountryRepo struct {
	mock.Mock
}

func (m *mockCountryRepo) GetCoordinates(ctx context.Context, country string) (*models.CountryCoordinates, error) {
	args := m.Called(ctx, country)
	return args.Get(0).(*models.CountryCoordinates), args.Error(1)
}

type mockTariffRepo struct {
	mock.Mock
}

func (m *mockTariffRepo) GetByCode(ctx context.Context, code string) (*models.Tariff, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*models.Tariff), args.Error(1)
}

func (m *mockTariffRepo) CreateTariff(ctx context.Context, tariff *models.Tariff) (*models.Tariff, error) {
	return nil, nil
}

func (m *mockTariffRepo) DeleteTariff(ctx context.Context, code string) error {
	return nil
}

func (m *mockTariffRepo) GetAll(ctx context.Context) ([]models.Tariff, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Tariff), args.Error(1)
}

func TestDefaultCalculator_Calculate(t *testing.T) {
	countryRepo := new(mockCountryRepo)

	fromCoords := &models.CountryCoordinates{Latitude: 55.75, Longitude: 37.61}
	toCoords := &models.CountryCoordinates{Latitude: 59.93, Longitude: 30.31}

	countryRepo.On("GetCoordinates", mock.Anything, "Russia").Return(fromCoords, nil)
	countryRepo.On("GetCoordinates", mock.Anything, "Russia2").Return(toCoords, nil)

	calculator := service.NewCalculator(countryRepo)

	pkg := models.Package{
		From:   "Russia",
		To:     "Russia2",
		Weight: 2,
		Length: 30,
		Width:  20,
		Height: 10,
	}

	result, err := calculator.Calculate(context.Background(), pkg)

	assert.NoError(t, err)
	assert.Greater(t, result.Cost, 0.0)
	assert.GreaterOrEqual(t, result.EstimatedHours, 6)
	assert.Equal(t, "RUB", result.Currency)
}

func TestExtendedCalculator_CalculateByTariffCode(t *testing.T) {
	countryRepo := new(mockCountryRepo)
	tariffRepo := new(mockTariffRepo)

	fromCoords := &models.CountryCoordinates{Latitude: 48.85, Longitude: 2.35}
	toCoords := &models.CountryCoordinates{Latitude: 51.51, Longitude: -0.13}

	countryRepo.On("GetCoordinates", mock.Anything, "France").Return(fromCoords, nil)
	countryRepo.On("GetCoordinates", mock.Anything, "UK").Return(toCoords, nil)

	tariff := &models.Tariff{
		Code:              "FAST",
		Name:              "Fast",
		BaseRate:          100,
		PricePerKm:        10,
		PricePerKg:        20,
		Currency:          "EUR",
		VolumetricDivider: 4000,
		SpeedKmph:         80,
	}

	tariffRepo.On("GetByCode", mock.Anything, "FAST").Return(tariff, nil)

	extCalc := service.NewExtendedCalculator(countryRepo, tariffRepo)

	pkg := models.Package{
		From:   "France",
		To:     "UK",
		Weight: 1.5,
		Length: 20,
		Width:  15,
		Height: 10,
	}

	result, err := extCalc.CalculateByTariffCode(context.Background(), pkg, "FAST")

	assert.NoError(t, err)
	assert.Greater(t, result.Cost, 0.0)
	assert.GreaterOrEqual(t, result.EstimatedHours, 6)
	assert.Equal(t, "EUR", result.Currency)
}

func TestDefaultCalculator_FallbackOnError(t *testing.T) {
	countryRepo := new(mockCountryRepo)

	countryRepo.On("GetCoordinates", mock.Anything, "Unknown").Return(&models.CountryCoordinates{}, assert.AnError)

	calculator := service.NewCalculator(countryRepo)

	pkg := models.Package{
		From:   "Unknown",
		To:     "Unknown",
		Weight: 3,
		Length: 30,
		Width:  20,
		Height: 15,
	}

	result, err := calculator.Calculate(context.Background(), pkg)

	assert.NoError(t, err)
	assert.Equal(t, 72, result.EstimatedHours)
	assert.Greater(t, result.Cost, 0.0)
}
