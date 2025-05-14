package grpcclient

import (
	"context"
	"time"

	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

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

	return &CalculatorGRPCClient{conn: conn, client: client}, nil
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

	return c.client.CalculateDeliveryCost(ctx, &calculatorpb.CalculateDeliveryCostRequest{
		Weight:  weight,
		From:    from,
		To:      to,
		Address: address,
		Width:   int32(width),
		Length:  int32(length),
		Height:  int32(height),
	})
}

func (c *CalculatorGRPCClient) CalculateByTariffCode(weight float64, userID, from, to, address, tariffCode string, length, width, height int) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	md := metadata.New(map[string]string{"authorization": userID})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.CalculateByTariffCode(ctx, &calculatorpb.CalculateByTariffRequest{
		Weight:     weight,
		From:       from,
		To:         to,
		Address:    address,
		Length:     int32(length),
		Width:      int32(width),
		Height:     int32(height),
		TariffCode: tariffCode,
	})
}

func (c *CalculatorGRPCClient) GetTariffList(userID string) (*calculatorpb.TariffListResponse, error) {
	md := metadata.New(map[string]string{"authorization": userID})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetTariffList(ctx, &calculatorpb.TariffListRequest{})
}
