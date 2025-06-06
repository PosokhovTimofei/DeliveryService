package clients

import (
	"context"
	"time"

	databasepb "github.com/maksroxx/DeliveryService/proto/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Packager interface {
	GetUserPackages(userId string) (*databasepb.PackageList, error)
}

type PackageGRPCClient struct {
	conn   *grpc.ClientConn
	client databasepb.PackageServiceClient
}

func NewPackageClient(address string) (*PackageGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
	if err != nil {
		return nil, err
	}
	client := databasepb.NewPackageServiceClient(conn)
	return &PackageGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *PackageGRPCClient) Close() error {
	return c.conn.Close()
}

func (p *PackageGRPCClient) withContext(userID string) (context.Context, context.CancelFunc) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return context.WithTimeout(ctx, 5*time.Second)
}

func (c *PackageGRPCClient) GetUserPackages(userId string) (*databasepb.PackageList, error) {
	ctx, cancel := c.withContext(userId)
	defer cancel()
	return c.client.GetUserPackages(ctx, &databasepb.PackageFilter{
		UserId: userId,
	})
}
