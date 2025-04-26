package pkg

type PaymentEvent struct {
	UserID    string  `json:"user_id"`
	PackageID string  `json:"package_id"`
	Cost      float64 `json:"cost"`
	Currency  string  `json:"currency"`
}
