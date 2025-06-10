package handlers

import (
	"context"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/middleware"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/service"
	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BidGRPCHandler struct {
	auctionpb.UnimplementedAuctionServiceServer
	svc    service.AuctionServicer
	logger *logrus.Logger
}

func NewBidGRPCHandler(svc service.AuctionServicer, log *logrus.Logger) *BidGRPCHandler {
	return &BidGRPCHandler{
		svc:    svc,
		logger: log,
	}
}

func (h *BidGRPCHandler) PlaceBid(ctx context.Context, req *auctionpb.BidRequest) (*auctionpb.BidResponse, error) {
	if req.PackageId == "" || req.UserId == "" || req.Amount <= 0 {
		return &auctionpb.BidResponse{Status: "error", Message: "invalid input"}, nil
	}

	bid := &models.Bid{
		PackageID: req.PackageId,
		UserID:    req.UserId,
		Amount:    req.Amount,
		Timestamp: time.Now(),
	}

	if err := h.svc.PlaceBid(ctx, bid); err != nil {
		return &auctionpb.BidResponse{Status: "error", Message: err.Error()}, nil
	}

	h.logger.Infof("Placed bid: %+v", bid)
	return &auctionpb.BidResponse{Status: "Success", Message: "Bid placed"}, nil
}

func (h *BidGRPCHandler) GetBidsByPackage(ctx context.Context, req *auctionpb.BidsRequest) (*auctionpb.BidsResponse, error) {
	if req.PackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "package_id is required")
	}

	bids, err := h.svc.GetBidsByPackage(ctx, req.PackageId)
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

func (h *BidGRPCHandler) StreamBids(req *auctionpb.BidsRequest, stream auctionpb.AuctionService_StreamBidsServer) error {
	if req.PackageId == "" {
		return status.Error(codes.InvalidArgument, "package_id is required")
	}

	ctx := stream.Context()
	changeStream, err := h.svc.StreamBids(ctx, req.PackageId)
	if err != nil {
		h.logger.WithError(err).Error("Failed to open change stream")
		return status.Errorf(codes.Internal, "change stream failed: %v", err)
	}
	defer changeStream.Close(ctx)

	for changeStream.Next(ctx) {
		var event struct {
			FullDocument models.Bid `bson:"fullDocument"`
		}
		if err := changeStream.Decode(&event); err != nil {
			h.logger.WithError(err).Error("decode stream event failed")
			continue
		}

		bid := event.FullDocument
		if err := stream.Send(&auctionpb.Bid{
			BidId:     bid.BidID,
			PackageId: bid.PackageID,
			UserId:    bid.UserID,
			Amount:    bid.Amount,
			Timestamp: bid.Timestamp.Format(time.RFC3339),
		}); err != nil {
			h.logger.WithError(err).Error("stream send failed")
			return err
		}
	}

	if err := changeStream.Err(); err != nil {
		h.logger.WithError(err).Error("change stream error")
		return err
	}

	return nil
}

func (h *BidGRPCHandler) GetAuctioningPackages(ctx context.Context, _ *auctionpb.Empty) (*auctionpb.Packages, error) {
	pkgs, err := h.svc.GetAuctioningPackages(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch auctioning packages: %v", err)
	}

	var res auctionpb.Packages
	for _, p := range pkgs {
		res.Package = append(res.Package, &auctionpb.Package{
			PackageId:  p.PackageID,
			Status:     p.Status,
			From:       p.From,
			To:         p.To,
			Weight:     p.Weight,
			Width:      int32(p.Width),
			Length:     int32(p.Length),
			Height:     int32(p.Height),
			Cost:       p.Cost,
			Currency:   p.Currency,
			TariffCode: p.TariffCode,
		})
	}
	return &res, nil
}

func (h *BidGRPCHandler) GetFailedPackages(ctx context.Context, _ *auctionpb.Empty) (*auctionpb.Packages, error) {
	pkgs, err := h.svc.GetFailedPackages(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch failed packages: %v", err)
	}

	var res auctionpb.Packages
	for _, p := range pkgs {
		res.Package = append(res.Package, &auctionpb.Package{
			PackageId:  p.PackageID,
			Status:     p.Status,
			From:       p.From,
			To:         p.To,
			Weight:     p.Weight,
			Width:      int32(p.Width),
			Length:     int32(p.Length),
			Height:     int32(p.Height),
			Cost:       p.Cost,
			Currency:   p.Currency,
			TariffCode: p.TariffCode,
		})
	}
	return &res, nil
}

func (h *BidGRPCHandler) GetUserWonPackages(ctx context.Context, _ *auctionpb.Empty) (*auctionpb.Packages, error) {
	userIDVal := ctx.Value(middleware.GRPCUserIDKey())
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized: userID not found in context")
	}

	pkgs, err := h.svc.GetUserWonPackages(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user packages: %v", err)
	}

	var res auctionpb.Packages
	for _, p := range pkgs {
		res.Package = append(res.Package, &auctionpb.Package{
			PackageId:  p.PackageID,
			Status:     p.Status,
			From:       p.From,
			To:         p.To,
			Weight:     p.Weight,
			Width:      int32(p.Width),
			Length:     int32(p.Length),
			Height:     int32(p.Height),
			Cost:       p.Cost,
			Currency:   p.Currency,
			TariffCode: p.TariffCode,
		})
	}
	return &res, nil
}

func (h *BidGRPCHandler) StartAuction(ctx context.Context, _ *auctionpb.Empty) (*auctionpb.Empty, error) {
	if err := h.svc.StartWaitingAuctions(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "start auction failed: %v", err)
	}
	return &auctionpb.Empty{}, nil
}

func (h *BidGRPCHandler) RepeateAuction(ctx context.Context, _ *auctionpb.Empty) (*auctionpb.Empty, error) {
	if err := h.svc.RepeatFailedAuctions(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "repeat auction failed: %v", err)
	}
	return &auctionpb.Empty{}, nil
}
