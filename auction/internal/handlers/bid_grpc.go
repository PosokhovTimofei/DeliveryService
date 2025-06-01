package handlers

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BidGRPCHandler struct {
	auctionpb.UnimplementedAuctionServiceServer
	bidRepo     repository.Bidder
	packageRepo repository.Packager
	logger      *logrus.Logger
}

func NewBidGRPCHandler(bidRepo repository.Bidder, packageRepo repository.Packager, log *logrus.Logger) *BidGRPCHandler {
	return &BidGRPCHandler{
		bidRepo:     bidRepo,
		packageRepo: packageRepo,
		logger:      log,
	}
}

func (s *BidGRPCHandler) PlaceBid(ctx context.Context, req *auctionpb.BidRequest) (*auctionpb.BidResponse, error) {
	if req.PackageId == "" {
		return &auctionpb.BidResponse{Status: "error", Message: "Invalid packageID "}, nil
	}
	if req.UserId == "" {
		return &auctionpb.BidResponse{Status: "error", Message: "Invalid userID "}, nil
	}
	if req.Amount <= 0 {
		return &auctionpb.BidResponse{Status: "error", Message: "Invalid amount "}, nil
	}

	pkg, err := s.packageRepo.FindByID(ctx, req.PackageId)
	if err != nil {
		return &auctionpb.BidResponse{Status: "error", Message: "Package not found"}, nil
	}
	if pkg.Status != "Auctioning" {
		return &auctionpb.BidResponse{Status: "error", Message: "Auction not active"}, nil
	}

	// через сколько аукцион на поссылку закончится
	auctionEnd := pkg.UpdatedAt.Add(5 * time.Minute)
	if time.Now().After(auctionEnd) {
		return &auctionpb.BidResponse{Status: "error", Message: "Auction ended"}, nil
	}

	topBid, err := s.bidRepo.GetTopBidByPackage(ctx, req.PackageId)
	if err == nil && topBid != nil && req.Amount <= topBid.Amount {
		return &auctionpb.BidResponse{Status: "error", Message: "Bid must be greater than current highest"}, nil
	}

	bid := &models.Bid{
		PackageID: req.PackageId,
		UserID:    req.UserId,
		Amount:    req.Amount,
		Timestamp: time.Now(),
	}
	if err := s.bidRepo.PlaceBid(ctx, bid); err != nil {
		return &auctionpb.BidResponse{Status: "error", Message: "Failed to save bid"}, nil
	}

	s.logger.Infof("Placed bid: package=%s user=%s amount=%.2f", req.PackageId, req.UserId, req.Amount)
	return &auctionpb.BidResponse{Status: "Success", Message: "Bid placed"}, nil
}

func (s *BidGRPCHandler) GetBidsByPackage(ctx context.Context, req *auctionpb.BidsRequest) (*auctionpb.BidsResponse, error) {
	if req.PackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "package_id is required")
	}

	bids, err := s.bidRepo.GetBidsByPackage(ctx, req.PackageId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch bids: %v", err)
	}

	var resp auctionpb.BidsResponse
	for _, b := range bids {
		resp.Bids = append(resp.Bids, &auctionpb.Bid{
			BidId:     b.BidID,
			PackageId: b.PackageID,
			UserId:    b.UserID,
			Amount:    b.Amount,
			Timestamp: b.Timestamp.Format(time.RFC3339),
		})
	}
	return &resp, nil
}

func (s *BidGRPCHandler) StreamBids(req *auctionpb.BidsRequest, stream auctionpb.AuctionService_StreamBidsServer) error {
	if req.PackageId == "" {
		return status.Error(codes.InvalidArgument, "package_id is required")
	}

	lastSent := make(map[string]bool)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("Client disconnected from stream")
			return nil

		case <-ticker.C:
			bids, err := s.bidRepo.GetBidsByPackage(stream.Context(), req.PackageId)
			if err != nil {
				s.logger.WithError(err).Error("Failed to fetch bids for stream")
				return status.Errorf(codes.Internal, "fetch failed: %v", err)
			}

			for _, b := range bids {
				if lastSent[b.BidID] {
					continue
				}
				lastSent[b.BidID] = true

				if err := stream.Send(&auctionpb.Bid{
					BidId:     b.BidID,
					PackageId: b.PackageID,
					UserId:    b.UserID,
					Amount:    b.Amount,
					Timestamp: b.Timestamp.Format(time.RFC3339),
				}); err != nil {
					s.logger.WithError(err).Error("Stream send failed")
					return err
				}
			}
		}
	}
}
