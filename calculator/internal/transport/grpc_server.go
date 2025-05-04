package transport

import (
	"context"
	"net"

	"github.com/maksroxx/DeliveryService/calculator/internal/middleware"
	"github.com/maksroxx/DeliveryService/calculator/internal/service"
	"github.com/maksroxx/DeliveryService/calculator/models"
	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	calculatorpb.UnimplementedCalculatorServiceServer
	service service.Calculator
	logger  *logrus.Logger
}

func NewGRPCServer(calc service.Calculator, logger *logrus.Logger) *GRPCServer {
	return &GRPCServer{
		service: calc,
		logger:  logger,
	}
}

func (s *GRPCServer) CalculateDeliveryCost(ctx context.Context, req *calculatorpb.CalculateDeliveryCostRequest) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	pkg := models.Package{
		Weight:  req.GetWeight(),
		From:    req.GetFrom(),
		To:      req.GetTo(),
		Address: req.GetAddress(),
	}

	if pkg.Weight <= 0 {
		err := status.Error(codes.InvalidArgument, "Invalid weight")
		s.logger.Error("Weight validation error: ", err)
		return nil, err
	}

	if err := ValidateAddress(pkg); err != nil {
		return nil, err
	}

	result, err := s.service.Calculate(pkg)
	if err != nil {
		s.logger.Errorf("gRPC CalculateDeliveryCost error: %v", err)
		return nil, status.Error(codes.Internal, "Calculation failed: "+err.Error())
	}

	return &calculatorpb.CalculateDeliveryCostResponse{
		Cost:           result.Cost,
		EstimatedHours: int32(result.EstimatedHours),
		Currency:       result.Currency,
	}, nil
}

func StartGRPCServer(port string, calc service.Calculator, logger *logrus.Logger) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.AuthInterceptor(),
			middleware.NewLoggingInterceptor(logger),
		),
	)
	calculatorpb.RegisterCalculatorServiceServer(grpcServer, NewGRPCServer(calc, logger))

	logger.Infof("gRPC server listening on :%s", port)
	return grpcServer.Serve(lis)
}
