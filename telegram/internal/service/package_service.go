package service

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/telegram/internal/clients"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
)

type PackageService struct {
	Repo   repository.Linker
	Client clients.Packager
}

func NewPackageService(repo repository.Linker, client clients.Packager) *PackageService {
	return &PackageService{Repo: repo, Client: client}
}

func (s *PackageService) GetUserPackages(ctx context.Context, telegramID int64) (string, error) {
	userID, err := s.Repo.GetUserIDByTelegramID(telegramID)
	if err != nil {
		return "", fmt.Errorf("user not linked")
	}

	res, err := s.Client.GetUserPackages(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch packages: %w", err)
	}

	return formatPackageList(res.Packages), nil
}
