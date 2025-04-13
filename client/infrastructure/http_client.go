package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/maksroxx/DeliveryService/client/domain"
)

type HTTPDeliveryClient struct {
	baseURL string
}

func NewHTTPDeliveryClient(baseURL string) *HTTPDeliveryClient {
	return &HTTPDeliveryClient{baseURL: baseURL}
}

func (c *HTTPDeliveryClient) CreatePackage(ctx context.Context, req domain.PackageRequest) (string, error) {
	url := c.baseURL + "/api/packages"

	body, err := json.Marshal(req)
	if err != nil {
		return "", nil
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("%d", resp.StatusCode)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (c *HTTPDeliveryClient) GetStatus(ctx context.Context, id string) (domain.PackageStatus, error) {
	url := c.baseURL + "/api/packages/" + id
	resp, err := http.Get(url)
	if err != nil {
		return domain.PackageStatus{}, err
	}
	defer resp.Body.Close()

	var status domain.PackageStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return domain.PackageStatus{}, err
	}

	return status, nil
}
