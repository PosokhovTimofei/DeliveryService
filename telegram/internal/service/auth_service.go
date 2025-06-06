package service

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/telegram/internal/clients"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
)

type AuthService struct {
	Repo   repository.Linker
	Client clients.Auther
}

func NewAuthService(repo repository.Linker, client clients.Auther) *AuthService {
	return &AuthService{Repo: repo, Client: client}
}

func (s *AuthService) GetUserIDByAuthCode(code string) (string, error) {
	resp, err := s.Client.GetUserByTelegramCode(code)
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}

func (s *AuthService) LinkTelegramAccount(ctx context.Context, code string, telegramID int64) error {
	userID, err := s.GetUserIDByAuthCode(code)
	if err != nil {
		return fmt.Errorf("cannot validate code: %w", err)
	}

	err = s.Repo.SaveLink(telegramID, userID)
	if err != nil {
		return fmt.Errorf("cannot save telegram link: %w", err)
	}

	return nil
}
