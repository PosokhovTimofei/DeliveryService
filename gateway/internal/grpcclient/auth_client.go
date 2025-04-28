package grpcclient

import (
	"context"
	"time"

	authpb "github.com/maksroxx/DeliveryService/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AuthGRPCClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthGRPCClient(address string) (*AuthGRPCClient, error) {
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

func (c *AuthGRPCClient) Register(email, password string) (*authpb.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &authpb.RegisterRequest{
		Email:    email,
		Password: password,
	}

	return c.client.Register(ctx, req)
}

func (c *AuthGRPCClient) Login(email, password string) (*authpb.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &authpb.LoginRequest{
		Email:    email,
		Password: password,
	}

	return c.client.Login(ctx, req)
}

func (c *AuthGRPCClient) Validate(token string) (*authpb.ValidateResponse, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Validate(ctx, &authpb.ValidateRequest{})
}
