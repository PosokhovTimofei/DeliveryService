package models

import "time"

type AuctionResult struct {
	PackageID  string    `json:"package_id"`
	WinnerID   string    `json:"winner_id"`
	FinalPrice float64   `json:"final_price"`
	Currency   string    `json:"currency"`
	FinishedAt time.Time `json:"finished_at"`
}

type Notification struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}
