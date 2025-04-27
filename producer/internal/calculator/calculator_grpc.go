package calculator

import (
	"context"
	"time"

	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"google.golang.org/grpc"
)

type CalculatorGRPCClient struct {
	conn   *grpc.ClientConn
	client calculatorpb.CalculatorServiceClient
}

func NewCalculatorClient(address string) (*CalculatorGRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	client := calculatorpb.NewCalculatorServiceClient(conn)

	return &CalculatorGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *CalculatorGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *CalculatorGRPCClient) Calculate(weight float64, from, to, address string) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &calculatorpb.CalculateDeliveryCostRequest{
		Weight:  weight,
		From:    from,
		To:      to,
		Address: address,
	}

	return c.client.CalculateDeliveryCost(ctx, req)
}
