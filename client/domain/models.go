package domain

type PackageRequest struct {
	Weight  float64 `json:"weight"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Address string  `json:"address"`
}

type PackageStatus struct {
	Status string `json:"status"`
}
