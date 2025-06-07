package models

import (
	"time"
)

type Package struct {
	ID             string    `bson:"_id,omitempty" json:"-"`
	PackageID      string    `bson:"package_id" json:"package_id"`
	UserID         string    `bson:"user_id" json:"-"`
	Weight         float64   `bson:"weight" json:"weight"`
	Length         int       `bson:"length" json:"length"`
	Width          int       `bson:"width" json:"width"`
	Height         int       `bson:"height" json:"height"`
	From           string    `bson:"from" json:"from"`
	To             string    `bson:"to" json:"to"`
	Address        string    `bson:"address" json:"address"`
	PaymentStatus  string    `bson:"payment_status" json:"payment_status"`
	Status         string    `bson:"status" json:"status"`
	Cost           float64   `bson:"cost" json:"cost"`
	EstimatedHours int       `bson:"estimated_hours" json:"estimated_hours"`
	RemainingHours int       `bson:"-" json:"remaining_hours"`
	Currency       string    `bson:"currency" json:"currency"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bosn:"updated_at" json:"updated_at"`
	TariffCode     string    `bson:"tariff_code" json:"tariff_code"`
}

type Payment struct {
	UserID    string  `bson:"user_id" json:"user_id"`
	PackageID string  `bson:"package_id" json:"package_id"`
	Cost      float64 `bson:"cost" json:"cost"`
	Currency  string  `bson:"currency" json:"currency"`
	Status    string  `bson:"status" json:"status"`
}

type PackageFilter struct {
	UserID       string    `form:"user_id"`
	Status       string    `form:"status"`
	CreatedAfter time.Time `form:"created_after"`
	Limit        int64     `form:"limit,default=20"`
	Offset       int64     `form:"offset,default=0"`
}

type PackageUpdate struct {
	Status        string `json:"status" validate:"oneof=created processing delivered canceled"`
	PaymentStatus string `json:"payment_status"`
}

// статусы
// Created
// Delivered
// Сanceled
// In pick-up point

type ExpiredPackageEvent struct {
	PackageID  string    `json:"package_id"`
	UserID     string    `json:"user_id"`
	Status     string    `json:"status"`
	Address    string    `json:"address"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Weight     float64   `json:"weight"`
	Length     int       `json:"length"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Cost       float64   `json:"cost"`
	Currency   string    `json:"currency"`
	TariffCode string    `json:"tariff_code"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
