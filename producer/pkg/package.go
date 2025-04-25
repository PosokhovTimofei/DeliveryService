package pkg

import "encoding/json"

type Package struct {
	ID             string  `bson:"package_id" json:"package_id"`
	UserID         string  `bson:"user_id" json:"-"`
	Weight         float64 `bson:"weight" json:"weight"`
	From           string  `bson:"from" json:"from"`
	To             string  `bson:"to" json:"to"`
	Address        string  `bson:"address" json:"address"`
	Status         string  `bson:"status" json:"status"`
	Cost           float64 `bson:"cost" json:"cost"`
	EstimatedHours int     `bson:"estimated_hours" json:"estimated_hours"`
	Currency       string  `bson:"currency" json:"currency"`
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
