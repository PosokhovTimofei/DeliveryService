package types

type Package struct {
	ID             string  `json:"package_id"`
	Weight         float64 `json:"weight"`
	From           string  `json:"from"`
	To             string  `json:"to"`
	Address        string  `json:"address"`
	Status         string  `json:"status"`
	Cost           float64 `json:"cost"`
	EstimatedHours int     `json:"estimated_hours"`
	Currency       string  `json:"currency"`
}
