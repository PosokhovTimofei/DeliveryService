package models

import "time"

type Bid struct {
	BidID     string    `bson:"bid_id" json:"bid_id"`
	PackageID string    `bson:"package_id" json:"package_id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Amount    float64   `bson:"amount" json:"amount"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}
