package calculator

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/consumer/types"
)

type Client interface {
	Calculate(ctx context.Context, pkg types.Package) error
	Close() error
}
type ClientType string

const (
	HTTP ClientType = "http"
	GRPC ClientType = "grpc"
)

func NewClient(client ClientType, address string) (Client, error) {
	switch client {
	case HTTP:
		return NewHTTPClient(address), nil
	case GRPC:
		return NewGRPCClient(address)
	default:
		return nil, fmt.Errorf("unknown client type: %s", client)
	}
}
