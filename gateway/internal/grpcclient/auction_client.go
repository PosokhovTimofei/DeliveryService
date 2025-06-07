package grpcclient

import (
	"context"
	"time"

	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AuctionGRPCClient struct {
	conn   *grpc.ClientConn
	client auctionpb.AuctionServiceClient
}

func NewAuctionGRPCClient(address string) (*AuctionGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, err
	}
	client := auctionpb.NewAuctionServiceClient(conn)
	return &AuctionGRPCClient{conn: conn, client: client}, nil
}

func (a *AuctionGRPCClient) Close() error {
	return a.conn.Close()
}

func (a *AuctionGRPCClient) withContext(userID string) (context.Context, context.CancelFunc) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return context.WithTimeout(ctx, 5*time.Second)
}

func (a *AuctionGRPCClient) PlaceBid(userID, packageID string, amount float64) (*auctionpb.BidResponse, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.PlaceBid(ctx, &auctionpb.BidRequest{
		PackageId: packageID,
		UserId:    userID,
		Amount:    amount,
	})
}

func (a *AuctionGRPCClient) GetBidsByPackage(userID, packageID string) (*auctionpb.BidsResponse, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.GetBidsByPackage(ctx, &auctionpb.BidsRequest{
		PackageId: packageID,
	})
}

func (a *AuctionGRPCClient) GetAuctioningPackages(userID string) (*auctionpb.Packages, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.GetAuctioningPackages(ctx, &auctionpb.Empty{})
}

func (a *AuctionGRPCClient) GetFailedPackages(userID string) (*auctionpb.Packages, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.GetFailedPackages(ctx, &auctionpb.Empty{})
}

func (a *AuctionGRPCClient) GetUserWonPackages(userID string) (*auctionpb.Packages, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.GetUserWonPackages(ctx, &auctionpb.Empty{})
}

func (a *AuctionGRPCClient) StartAuction(userID string) (*auctionpb.Empty, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.StartAuction(ctx, &auctionpb.Empty{})
}

func (a *AuctionGRPCClient) RepeateAuction(userID string) (*auctionpb.Empty, error) {
	ctx, cancel := a.withContext(userID)
	defer cancel()
	return a.client.RepeateAuction(ctx, &auctionpb.Empty{})
}

func (a *AuctionGRPCClient) StreamBids(userID, packageID string) (auctionpb.AuctionService_StreamBidsClient, context.CancelFunc, error) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithCancel(ctx)

	stream, err := a.client.StreamBids(ctx, &auctionpb.BidsRequest{PackageId: packageID})
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return stream, cancel, nil
}
