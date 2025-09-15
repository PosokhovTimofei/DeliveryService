package service

import (
	"context"
	"errors"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/metrics"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionServicer interface {
	PlaceBid(ctx context.Context, bid *models.Bid) error
	GetBidsByPackage(ctx context.Context, packageID string) ([]*models.Bid, error)
	StreamBids(ctx context.Context, packageID string) (*mongo.ChangeStream, error)
	GetAuctioningPackages(ctx context.Context) ([]*models.Package, error)
	GetFailedPackages(ctx context.Context) ([]*models.Package, error)
	GetUserWonPackages(ctx context.Context, userID string) ([]*models.Package, error)
	StartWaitingAuctions(ctx context.Context) error
	RepeatFailedAuctions(ctx context.Context) error
	DetermineWinner(ctx context.Context, packageID string) (*models.Bid, error)
	SetAuctionDuration(duration time.Duration)
}

type AuctionService struct {
	bidRepo         repository.Bidder
	packageRepo     repository.Packager
	producer        kafka.AucPublisher
	logger          *logrus.Logger
	auctionDuration time.Duration
}

func NewAuctionService(bidRepo repository.Bidder, packageRepo repository.Packager, producer kafka.AucPublisher, logger *logrus.Logger) *AuctionService {
	return &AuctionService{
		bidRepo:         bidRepo,
		packageRepo:     packageRepo,
		producer:        producer,
		logger:          logger,
		auctionDuration: 2 * time.Minute,
	}
}

func (s *AuctionService) PlaceBid(ctx context.Context, bid *models.Bid) error {
	pkg, err := s.packageRepo.FindByID(ctx, bid.PackageID)
	if err != nil {
		metrics.BidErrorsTotal.WithLabelValues("package_not_found").Inc()
		return err
	}
	if pkg.Status != "Auctioning" {
		metrics.BidErrorsTotal.WithLabelValues("auction_not_active").Inc()
		return errors.New("auction not active")
	}
	if time.Now().After(pkg.UpdatedAt.Add(s.auctionDuration)) {
		metrics.BidErrorsTotal.WithLabelValues("auction_ended").Inc()
		return errors.New("auction ended")
	}
	topBid, err := s.bidRepo.GetTopBidByPackage(ctx, bid.PackageID)
	if err == nil && topBid != nil && bid.Amount <= topBid.Amount {
		metrics.BidErrorsTotal.WithLabelValues("bid_too_low").Inc()
		return errors.New("bid must be greater than current highest")
	}
	err = s.bidRepo.PlaceBid(ctx, bid)
	if err == nil {
		metrics.BidsPlacedTotal.Inc()
	}
	return err
}

func (s *AuctionService) GetBidsByPackage(ctx context.Context, packageID string) ([]*models.Bid, error) {
	return s.bidRepo.GetBidsByPackage(ctx, packageID)
}

func (s *AuctionService) StreamBids(ctx context.Context, packageID string) (*mongo.ChangeStream, error) {
	return s.bidRepo.WatchBidsByPackage(ctx, packageID)
}

func (s *AuctionService) GetAuctioningPackages(ctx context.Context) ([]*models.Package, error) {
	return s.packageRepo.FindByAuctioningStatus(ctx)
}

func (s *AuctionService) GetFailedPackages(ctx context.Context) ([]*models.Package, error) {
	return s.packageRepo.FindByFailedStatus(ctx)
}

func (s *AuctionService) GetUserWonPackages(ctx context.Context, userID string) ([]*models.Package, error) {
	return s.packageRepo.FindUserPackages(ctx, userID)
}

func (s *AuctionService) StartWaitingAuctions(ctx context.Context) error {
	pkgs, err := s.packageRepo.FindByWaitingStatus(ctx)
	if err != nil {
		return err
	}
	for _, p := range pkgs {
		StartAuction(p, s, s.producer, s.packageRepo, s.logger, s.auctionDuration)
	}
	return nil
}

func (s *AuctionService) RepeatFailedAuctions(ctx context.Context) error {
	pkgs, err := s.packageRepo.FindByFailedStatus(ctx)
	if err != nil {
		return err
	}
	for _, p := range pkgs {
		StartAuction(p, s, s.producer, s.packageRepo, s.logger, s.auctionDuration)
	}
	return nil
}

func (s *AuctionService) DetermineWinner(ctx context.Context, packageID string) (*models.Bid, error) {
	return s.bidRepo.GetTopBidByPackage(ctx, packageID)
}

func (s *AuctionService) SetAuctionDuration(duration time.Duration) {
	s.auctionDuration = duration
}
