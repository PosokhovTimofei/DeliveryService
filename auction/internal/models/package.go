package models

import "time"

type Package struct {
	PackageID  string    `bson:"package_id" json:"package_id"`
	UserID     string    `bson:"user_id" json:"user_id"`
	Status     string    `bson:"status" json:"status"`
	Address    string    `bson:"address" json:"address"`
	From       string    `bson:"from" json:"from"`
	To         string    `bson:"to" json:"to"`
	Cost       float64   `bson:"cost" json:"cost"`
	Currency   string    `bson:"currency" json:"currency"`
	TariffCode string    `bson:"tariff_code" json:"tariff_code"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}
