package service

import (
	"context"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
)

type AuctionService struct {
	bidRepo repository.Bidder
}

func NewAuctionService(bidRepo repository.Bidder) *AuctionService {
	return &AuctionService{bidRepo: bidRepo}
}

func (a *AuctionService) DetermineWinner(ctx context.Context, packageID string) (*models.Bid, error) {
	return a.bidRepo.GetTopBidByPackage(ctx, packageID)
}
