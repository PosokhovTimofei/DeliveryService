package clients

import (
	"fmt"
	"time"

	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	databasepb "github.com/maksroxx/DeliveryService/proto/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	Packages databasepb.PackageServiceClient
	Auction  auctionpb.AuctionServiceClient

	pkgConn     *grpc.ClientConn
	auctionConn *grpc.ClientConn
}

func InitGRPCClients() (*GRPCClients, error) {
	pkgConn, err := grpc.NewClient(
		"localhost:50054",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("package grpc: %w", err)
	}
	auctionConn, err := grpc.NewClient(
		"localhost:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("auction grpc: %w", err)
	}

	return &GRPCClients{
		Packages:    databasepb.NewPackageServiceClient(pkgConn),
		Auction:     auctionpb.NewAuctionServiceClient(auctionConn),
		pkgConn:     pkgConn,
		auctionConn: auctionConn,
	}, nil
}

func (c *GRPCClients) Close() {
	if c.pkgConn != nil {
		c.pkgConn.Close()
	}
	if c.auctionConn != nil {
		c.auctionConn.Close()
	}
}
