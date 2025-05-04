package grpcclient

import (
	"context"
	"time"

	pb "github.com/maksroxx/DeliveryService/proto/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type PaymentGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.PaymentServiceClient
}

func NewPaymentGRPCClient(address string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewPaymentServiceClient(conn)
	return &PaymentGRPCClient{conn: conn, client: client}, nil
}

func (p *PaymentGRPCClient) Close() error {
	return p.conn.Close()
}

func (p *PaymentGRPCClient) ConfirmPayment(ctx context.Context, userID, packageID string) (string, error) {
	md := metadata.New(map[string]string{"authorization": userID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := p.client.ConfirmPayment(ctx, &pb.ConfirmPaymentRequest{
		UserId:    userID,
		PackageId: packageID,
	})
	if err != nil {
		return "", err
	}

	return resp.GetMessage(), nil
}
