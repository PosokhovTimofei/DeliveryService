package models

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
