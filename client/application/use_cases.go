package application

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/client/domain"
)

type DeliveryClient interface {
	CreatePackage(ctx context.Context, req domain.PackageRequest) (string, error)
	GetStatus(ctx context.Context, id string) (domain.PackageStatus, error)
}

type CreatePackageUseCase struct {
	client DeliveryClient
}

func NewCreatePackageUseCase(client DeliveryClient) *CreatePackageUseCase {
	return &CreatePackageUseCase{client: client}
}

func (uc *CreatePackageUseCase) Execute(req domain.PackageRequest) error {
	id, err := uc.client.CreatePackage(context.Background(), req)
	if err != nil {
		return err
	}
	fmt.Printf("Package created. ID: %s\n", id)
	return nil
}

type GetStatusUseCase struct {
	client DeliveryClient
}

func NewGetStatusUseCase(client DeliveryClient) *GetStatusUseCase {
	return &GetStatusUseCase{client: client}
}

func (uc *GetStatusUseCase) Execute(id string) error {
	status, err := uc.client.GetStatus(context.Background(), id)
	if err != nil {
		return err
	}
	fmt.Printf("Package status %s\n", status.Status)
	return nil
}
