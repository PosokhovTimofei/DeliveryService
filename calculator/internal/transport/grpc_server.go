package transport

import (
	"context"
	"net"

	"github.com/maksroxx/DeliveryService/calculator/internal/middleware"
	"github.com/maksroxx/DeliveryService/calculator/internal/repository"
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
	service service.ExtendedCalculator
	logger  *logrus.Logger
}

func NewGRPCServer(calc service.ExtendedCalculator, logger *logrus.Logger) *GRPCServer {
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
		Length:  int(req.GetLength()),
		Height:  int(req.GetHeight()),
		Width:   int(req.GetWidth()),
	}

	if pkg.Weight <= 0 {
		err := status.Error(codes.InvalidArgument, "Invalid weight")
		s.logger.Error("Weight validation error: ", err)
		return nil, err
	}

	if err := Validate(pkg); err != nil {
		return nil, err
	}

	result, err := s.service.Calculate(context.Background(), pkg)
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

func (s *GRPCServer) CalculateByTariffCode(ctx context.Context, req *calculatorpb.CalculateByTariffRequest) (*calculatorpb.CalculateDeliveryCostResponse, error) {
	pkg := models.Package{
		Weight: req.Weight,
		From:   req.From,
		To:     req.To,
		Length: int(req.Length),
		Width:  int(req.Width),
		Height: int(req.Height),
	}
	res, err := s.service.CalculateByTariffCode(ctx, pkg, req.TariffCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "calculation error: %v", err)
	}
	return &calculatorpb.CalculateDeliveryCostResponse{
		Cost:           res.Cost,
		EstimatedHours: int32(res.EstimatedHours),
		Currency:       res.Currency,
	}, nil
}

func (s *GRPCServer) GetTariffList(ctx context.Context, _ *calculatorpb.TariffListRequest) (*calculatorpb.TariffListResponse, error) {
	tariffs, err := s.service.GetTariffs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tariffs: %v", err)
	}
	var result []*calculatorpb.Tariff
	for _, t := range tariffs {
		result = append(result, &calculatorpb.Tariff{
			Code:              t.Code,
			Name:              t.Name,
			BaseRate:          t.BaseRate,
			PricePerKm:        t.PricePerKm,
			PricePerKg:        t.PricePerKg,
			Currency:          t.Currency,
			VolumetricDivider: t.VolumetricDivider,
			SpeedKmph:         int32(t.SpeedKmph),
		})
	}
	return &calculatorpb.TariffListResponse{Tariffs: result}, nil
}

func StartGRPCServer(port string, rep repository.CountryRepository, calc service.ExtendedCalculator, logger *logrus.Logger) error {
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
