package types

type Package struct {
	ID      string  `json:"id"`
	Weight  float64 `json:"weight"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Address string  `json:"address"`
	Status  string  `json:"status"`
}
