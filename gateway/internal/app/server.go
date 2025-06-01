package app

import (
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/sirupsen/logrus"
)

func NewServer(logger *logrus.Logger) *http.Server {
	calculatorClient, err := grpcclient.NewCalculatorClient("localhost:50051")
	if err != nil {
		logger.Fatalf("Failed to connect to calculator gRPC: %v", err)
	}
	authClient, err := grpcclient.NewAuthGRPCClient("localhost:50052")
	if err != nil {
		logger.Fatalf("Failed to connect to auth gRPC: %v", err)
	}
	paymentClient, err := grpcclient.NewPaymentGRPCClient("localhost:50053")
	if err != nil {
		logger.Fatalf("Failed to connect to payment gRPC: %v", err)
	}
	packageClient, err := grpcclient.NewPackageGRPCClient("localhost:50054")
	if err != nil {
		logger.Fatalf("Failed to connect to package gRPC: %v", err)
	}

	auctionClient, err := grpcclient.NewAuctionGRPCClient("localhost:50055")
	if err != nil {
		logger.Fatalf("Failed to connect to package gRPC: %v", err)
	}
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, logger, authClient, calculatorClient, paymentClient, packageClient, auctionClient)

	return &http.Server{
		Addr:              ":8228",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
