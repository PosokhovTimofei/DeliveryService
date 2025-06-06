package models

type Notification struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}
