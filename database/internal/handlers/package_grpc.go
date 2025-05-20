package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/database/internal/kafka"
	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
	calculatorpb "github.com/maksroxx/DeliveryService/proto/calculator"
	pb "github.com/maksroxx/DeliveryService/proto/database"
)

var (
	ErrInvalidInput          = fmt.Errorf("invalid input")
	ErrCannotCancelDelivered = fmt.Errorf("cannot cancel a delivered package")
	ErrAlreadyCanceled       = fmt.Errorf("package already canceled")
)

type GrpcPackageHandler struct {
	pb.UnimplementedPackageServiceServer
	rep        repository.RouteRepository
	calculator Calculator
	producer   kafka.PaymentProducer
}

func NewGrpcPackageHandler(rep repository.RouteRepository, calculator Calculator, producer kafka.PaymentProducer) *GrpcPackageHandler {
	return &GrpcPackageHandler{
		rep:        rep,
		calculator: calculator,
		producer:   producer,
	}
}

func (h *GrpcPackageHandler) GetPackage(ctx context.Context, req *pb.PackageID) (*pb.Package, error) {
	pkg, err := h.rep.GetByID(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return toProto(pkg), nil
}

func (h *GrpcPackageHandler) GetAllPackages(ctx context.Context, req *pb.PackageFilter) (*pb.PackageList, error) {
	filter := models.PackageFilter{
		Status: req.Status,
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	if req.CreatedAfter != nil {
		filter.CreatedAfter = req.CreatedAfter.AsTime()
	}

	pkgs, err := h.rep.GetAllRoutes(ctx, filter)
	if err != nil {
		return nil, err
	}

	return toProtoList(pkgs), nil
}

func (h *GrpcPackageHandler) GetUserPackages(ctx context.Context, req *pb.PackageFilter) (*pb.PackageList, error) {
	filter := models.PackageFilter{
		UserID: req.UserId,
		Status: req.Status,
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	if req.CreatedAfter != nil {
		filter.CreatedAfter = req.CreatedAfter.AsTime()
	}

	pkgs, err := h.rep.GetAllRoutes(ctx, filter)
	if err != nil {
		return nil, err
	}
	return toProtoList(pkgs), nil
}

func (h *GrpcPackageHandler) CreatePackage(ctx context.Context, req *pb.Package) (*pb.Package, error) {
	if req.Weight <= 0 || req.From == "" || req.To == "" || req.Address == "" || req.Length <= 0 || req.Width <= 0 || req.Height <= 0 {
		return nil, ErrInvalidInput
	}

	pkg := &models.Package{
		PackageID:      req.PackageId,
		UserID:         req.UserId,
		Weight:         req.Weight,
		Length:         int(req.Length),
		Width:          int(req.Width),
		Height:         int(req.Height),
		From:           req.From,
		To:             req.To,
		Address:        req.Address,
		PaymentStatus:  req.PaymentStatus,
		Cost:           req.Cost,
		EstimatedHours: int(req.EstimatedHours),
		Currency:       req.Currency,
		CreatedAt:      time.Now(),
	}

	created, err := h.rep.Create(ctx, pkg)
	if err != nil {
		return nil, err
	}
	return toProto(created), nil
}

func (h *GrpcPackageHandler) CreatePackageWithCalc(ctx context.Context, req *pb.Package) (*pb.Package, error) {

	if req.Weight <= 0 || req.From == "" || req.To == "" || req.Address == "" || req.Length <= 0 || req.Width <= 0 || req.Height <= 0 {
		return nil, ErrInvalidInput
	}

	var (
		result *calculatorpb.CalculateDeliveryCostResponse
		err    error
	)
	tariffCode := req.TariffCode
	if tariffCode == "" {
		result, err = h.calculator.Calculate(req.Weight, req.UserId, req.From, req.To, req.Address, int(req.Length), int(req.Width), int(req.Height))
		tariffCode = "DEFAULT"
	} else {
		result, err = h.calculator.CalculateByTariff(req.Weight, req.UserId, req.From, req.To, req.Address, tariffCode, int(req.Length), int(req.Width), int(req.Height))
	}
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	model := &models.Package{
		PackageID:      "PKG-" + uuid.New().String(),
		UserID:         req.UserId,
		Weight:         req.Weight,
		Length:         int(req.Length),
		Width:          int(req.Width),
		Height:         int(req.Height),
		From:           req.From,
		To:             req.To,
		Address:        req.Address,
		Status:         "Created",
		PaymentStatus:  "PENDING",
		Cost:           result.Cost,
		EstimatedHours: int(result.EstimatedHours),
		Currency:       result.Currency,
		CreatedAt:      time.Now(),
		TariffCode:     tariffCode,
	}

	_, err = h.rep.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	payment := models.Payment{
		UserID:    model.UserID,
		PackageID: model.PackageID,
		Cost:      model.Cost,
		Currency:  model.Currency,
	}

	if err := h.producer.SendPaymentEvent(payment); err != nil {
		return nil, fmt.Errorf("failed to send payment event: %w", err)
	}

	return toProto(model), nil
}

func (h *GrpcPackageHandler) UpdatePackage(ctx context.Context, req *pb.Package) (*pb.Package, error) {
	update := models.PackageUpdate{
		Status:        req.Status,
		PaymentStatus: req.PaymentStatus,
	}
	updated, err := h.rep.UpdateRoute(ctx, req.PackageId, update)
	if err != nil {
		return nil, err
	}
	return toProto(updated), nil
}

func (h *GrpcPackageHandler) DeletePackage(ctx context.Context, req *pb.PackageID) (*pb.Empty, error) {
	err := h.rep.DeleteRoute(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (h *GrpcPackageHandler) CancelPackage(ctx context.Context, req *pb.PackageID) (*pb.Package, error) {
	pkg, err := h.rep.GetByID(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	if pkg.Status == "Delivered" {
		return nil, ErrCannotCancelDelivered
	}
	if pkg.Status == "Сanceled" {
		return nil, ErrAlreadyCanceled
	}

	update := models.PackageUpdate{
		Status: "Сanceled",
	}
	updated, err := h.rep.UpdateRoute(ctx, req.PackageId, update)
	if err != nil {
		return nil, err
	}
	return toProto(updated), nil
}

func (h *GrpcPackageHandler) GetPackageStatus(ctx context.Context, req *pb.PackageID) (*pb.PackageStatus, error) {
	pkg, err := h.rep.GetByID(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return &pb.PackageStatus{Status: pkg.Status}, nil
}
