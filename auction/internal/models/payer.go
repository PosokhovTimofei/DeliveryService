package models

type PaidPackageEvent struct {
	PackageID string `json:"package_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
}
