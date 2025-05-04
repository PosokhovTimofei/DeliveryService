package handler

import (
	"context"

	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/kafka"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
	pb "github.com/maksroxx/DeliveryService/proto/payment"
)

type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	repo     db.Paymenter
	producer kafka.Producerer
}

func NewPaymentGRPCServer(repo db.Paymenter, producer kafka.Producerer) *PaymentGRPCServer {
	return &PaymentGRPCServer{repo: repo, producer: producer}
}

func (s *PaymentGRPCServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	payment, err := s.repo.UpdatePayment(ctx, models.Payment{
		UserID:    req.GetUserId(),
		PackageID: req.GetPackageId(),
		Status:    models.PaymentStatusPaid,
	})
	if err != nil {
		return nil, err
	}

	if err := s.producer.PaymentMessage(*payment, req.GetUserId()); err != nil {
		return nil, err
	}

	return &pb.ConfirmPaymentResponse{
		Message: "Payment confirmed and event sent",
	}, nil
}
