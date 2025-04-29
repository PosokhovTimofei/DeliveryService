package models

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusPaid      PaymentStatus = "PAID"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
	PaymentStatusExpired   PaymentStatus = "EXPIRED"
)

type Payment struct {
	UserID    string        `bson:"user_id" json:"user_id"`
	PackageID string        `bson:"package_id" json:"package_id"`
	Cost      float64       `bson:"cost" json:"cost"`
	Currency  string        `bson:"currency" json:"currency"`
	Status    PaymentStatus `bson:"status" json:"status"`
}
