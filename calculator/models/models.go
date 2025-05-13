package models

type Package struct {
	Weight  float64 `json:"weight"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Address string  `json:"address"`
	Length  int     `json:"length"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
}

type CalculationResult struct {
	Cost           float64 `json:"cost"`
	EstimatedHours int     `json:"estimated_hours"`
	Currency       string  `json:"currency"`
}
