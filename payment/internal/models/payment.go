package models

type Payment struct {
	UserID    string  `bson:"user_id" json:"user_id"`
	PackageID string  `bson:"package_id" json:"package_id"`
	Cost      float64 `bson:"cost" json:"cost"`
	Currency  string  `bson:"currency" json:"currency"`
	Status    string  `bson:"status" json:"status"`
}
