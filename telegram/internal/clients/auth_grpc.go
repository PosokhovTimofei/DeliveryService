package clients

import (
	"context"
	"time"

	authpb "github.com/maksroxx/DeliveryService/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Auther interface {
	GetUserByTelegramCode(code string) (*authpb.TelegramCodeLookupResponse, error)
}

type AuthGRPCClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthClient(address string) (*AuthGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
	if err != nil {
		return nil, err
	}
	client := authpb.NewAuthServiceClient(conn)
	return &AuthGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *AuthGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *AuthGRPCClient) GetUserByTelegramCode(code string) (*authpb.TelegramCodeLookupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.GetUserByTelegramCode(ctx, &authpb.TelegramCodeLookupRequest{Code: code})
}
