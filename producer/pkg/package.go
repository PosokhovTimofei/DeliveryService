package pkg

import "encoding/json"

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

func (p *Package) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Package) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type CalculationResult struct {
	Cost           float64 `json:"cost"`
	EstimatedHours int     `json:"estimated_hours"`
	Currency       string  `json:"currency"`
}
