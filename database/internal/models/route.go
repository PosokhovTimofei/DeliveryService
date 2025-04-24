package models

import (
	"time"
)

type Package struct {
	ID             string    `bson:"_id,omitempty" json:"id,omitempty"`
	PackageID      string    `bson:"package_id" json:"package_id"`
	UserID         string    `bson:"user_id" json:"-"`
	Weight         float64   `bson:"weight" json:"weight"`
	From           string    `bson:"from" json:"from"`
	To             string    `bson:"to" json:"to"`
	Address        string    `bson:"address" json:"address"`
	Status         string    `bson:"status" json:"status"`
	Cost           float64   `bson:"cost" json:"cost"`
	EstimatedHours int       `bson:"estimated_hours" json:"estimated_hours"`
	Currency       string    `bson:"currency" json:"currency"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
}

type RouteFilter struct {
	UserID       string    `form:"user_id"`
	Status       string    `form:"status"`
	CreatedAfter time.Time `form:"created_after"`
	Limit        int64     `form:"limit,default=20"`
	Offset       int64     `form:"offset,default=0"`
}

type RouteUpdate struct {
	Status string `json:"status" validate:"oneof=created processing delivered canceled"`
}
