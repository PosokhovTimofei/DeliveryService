package grpcclient

import (
	"context"
	"time"

	databasepb "github.com/maksroxx/DeliveryService/proto/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type PackageGRPCClient struct {
	conn   *grpc.ClientConn
	client databasepb.PackageServiceClient
}

func NewPackageGRPCClient(address string) (*PackageGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 5 * time.Second}),
	)
	if err != nil {
		return nil, err
	}
	client := databasepb.NewPackageServiceClient(conn)
	return &PackageGRPCClient{conn: conn, client: client}, nil
}

func (p *PackageGRPCClient) Close() error {
	return p.conn.Close()
}

func (p *PackageGRPCClient) withContext(userID string) (context.Context, context.CancelFunc) {
	md := metadata.New(map[string]string{
		"authorization": userID,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return context.WithTimeout(ctx, 5*time.Second)
}

func (p *PackageGRPCClient) GetPackage(userID, packageID string) (*databasepb.Package, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.GetPackage(ctx, &databasepb.PackageID{PackageId: packageID})
}

func (p *PackageGRPCClient) GetAllPackages(userID, status string, limit, offset int64) (*databasepb.PackageList, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.GetAllPackages(ctx, &databasepb.PackageFilter{
		UserId: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
}

func (p *PackageGRPCClient) GetUserPackages(userID, status string, limit, offset int64) (*databasepb.PackageList, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.GetUserPackages(ctx, &databasepb.PackageFilter{
		UserId: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
}

func (p *PackageGRPCClient) CreatePackage(userID string, pkg *databasepb.Package) (*databasepb.Package, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.CreatePackage(ctx, pkg)
}

func (p *PackageGRPCClient) UpdatePackage(userID string, pkg *databasepb.Package) (*databasepb.Package, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.UpdatePackage(ctx, pkg)
}

func (p *PackageGRPCClient) DeletePackage(userID, packageID string) (*databasepb.Empty, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.DeletePackage(ctx, &databasepb.PackageID{PackageId: packageID})
}

func (p *PackageGRPCClient) CancelPackage(userID, packageID string) (*databasepb.Package, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.CancelPackage(ctx, &databasepb.PackageID{PackageId: packageID})
}

func (p *PackageGRPCClient) GetPackageStatus(userID, packageID string) (*databasepb.PackageStatus, error) {
	ctx, cancel := p.withContext(userID)
	defer cancel()
	return p.client.GetPackageStatus(ctx, &databasepb.PackageID{PackageId: packageID})
}
