package clients

import (
	"context"
	"time"

	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Calculator interface {
	Calculate(weight float64, userID, from, to, address string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error)
	CalculateByTariff(weight float64, userID, from, to, address, tariffCode string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error)
}

type CalculatorGRPCClient struct {
	conn   *grpc.ClientConn
	client calculatorpb.CalculatorServiceClient
}

func NewCalculatorClient(address string) (*CalculatorGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
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

func (c *CalculatorGRPCClient) Calculate(weight float64, userID, from, to, address string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &calculatorpb.CalculateDeliveryCostRequest{
		Weight:  weight,
		From:    from,
		To:      to,
		Address: address,
		Width:   int32(width),
		Length:  int32(length),
		Height:  int32(height),
	}

	return c.client.CalculateDeliveryCost(ctx, req)
}

func (c *CalculatorGRPCClient) CalculateByTariff(weight float64, userID, from, to, address, tariff_code string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &calculatorpb.CalculateByTariffRequest{
		Weight:     weight,
		From:       from,
		To:         to,
		Address:    address,
		Width:      int32(width),
		Length:     int32(length),
		Height:     int32(height),
		TariffCode: tariff_code,
	}
	return c.client.CalculateByTariffCode(ctx, req)
}
