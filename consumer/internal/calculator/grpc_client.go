package calculator

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/consumer/types"
)

type GRPCClient struct {
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	return &GRPCClient{}, nil
}

func (c *GRPCClient) Calculate(ctx context.Context, pkg types.Package) error {
	fmt.Printf("gRPC calculation for package: %+v\n", pkg)
	return nil
}

func (c *GRPCClient) Close() error {
	return nil
}
