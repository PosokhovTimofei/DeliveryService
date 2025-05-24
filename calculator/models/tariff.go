package models

import "fmt"

type Tariff struct {
	Code              string  `bson:"code" json:"code"`
	Name              string  `bson:"name" json:"name"`
	BaseRate          float64 `bson:"base_rate" json:"base_rate"`
	PricePerKm        float64 `bson:"price_per_km" json:"price_per_km"`
	PricePerKg        float64 `bson:"price_per_kg" json:"price_per_kg"`
	Currency          string  `bson:"currency" json:"currency"`
	VolumetricDivider float64 `bson:"volumetric_divider" json:"volumetric_divider"`
	SpeedKmph         float64 `bson:"speed_kmph" json:"speed_kmph"`
}

func (t *Tariff) Validate() error {
	if t.Code == "" {
		return fmt.Errorf("code is required")
	}
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.BaseRate <= 0 {
		return fmt.Errorf("base_rate must be positive")
	}
	if t.PricePerKm <= 0 {
		return fmt.Errorf("price_per_km must be positive")
	}
	if t.PricePerKg <= 0 {
		return fmt.Errorf("price_per_kg must be positive")
	}
	if t.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if t.VolumetricDivider <= 0 {
		return fmt.Errorf("volumetric_divider must be positive")
	}
	if t.SpeedKmph <= 0 {
		return fmt.Errorf("speed_kmph must be positive")
	}
	return nil
}
