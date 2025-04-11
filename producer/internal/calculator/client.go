package calculator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Calculate(pack pkg.Package) (*pkg.CalculationResult, error) {
	data, err := json.Marshal(pack)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Post(c.baseURL+"/calculate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %s", resp.Status)
	}

	var result pkg.CalculationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
