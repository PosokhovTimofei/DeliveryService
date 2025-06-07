package models

import "time"

type Package struct {
	PackageID  string    `bson:"package_id" json:"package_id"`
	UserID     string    `bson:"user_id" json:"user_id"`
	Status     string    `bson:"status" json:"status"`
	Address    string    `bson:"address" json:"address"`
	From       string    `bson:"from" json:"from"`
	To         string    `bson:"to" json:"to"`
	Weight     float64   `bson:"weight" json:"weight"`
	Length     int       `bson:"length" json:"length"`
	Width      int       `bson:"width" json:"width"`
	Height     int       `bson:"height" json:"height"`
	Cost       float64   `bson:"cost" json:"cost"`
	Currency   string    `bson:"currency" json:"currency"`
	TariffCode string    `bson:"tariff_code" json:"tariff_code"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

type DeliveryInit struct {
	PackageID string  `json:"package_id"`
	UserID    string  `json:"user_id"`
	From      string  `json:"from"`
	Address   string  `json:"address"`
	Weight    float64 `json:"weight"`
	Length    int     `json:"length"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Cost      float64 `json:"cost"`
	Currency  string  `json:"currency"`
}
