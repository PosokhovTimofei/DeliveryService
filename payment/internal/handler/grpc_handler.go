package handler

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/payment/internal/service"
	pb "github.com/maksroxx/DeliveryService/proto/payment"
)

type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	svc service.PaymentService
}

func NewPaymentGRPCServer(svc service.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{svc: svc}
}

func (s *PaymentGRPCServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	payment, err := s.svc.ConfirmPayment(ctx, req.GetUserId(), req.GetPackageId())
	if err != nil {
		return nil, err
	}
	return &pb.ConfirmPaymentResponse{Message: fmt.Sprintf("Payment confirmed: %s", payment.PackageID)}, nil
}

func (s *PaymentGRPCServer) ConfirmAuctionPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	payment, err := s.svc.ConfirmAuctionPayment(ctx, req.GetUserId(), req.GetPackageId())
	if err != nil {
		return nil, err
	}
	return &pb.ConfirmPaymentResponse{Message: fmt.Sprintf("Auction payment confirmed: %s", payment.PackageID)}, nil
}
