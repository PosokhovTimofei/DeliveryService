package handlers

import (
	"context"
	"fmt"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/service"
	pb "github.com/maksroxx/DeliveryService/proto/database"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidInput          = fmt.Errorf("invalid input")
	ErrCannotCancelDelivered = fmt.Errorf("cannot cancel a delivered package")
	ErrAlreadyCanceled       = fmt.Errorf("package already canceled")
)

type GrpcPackageHandler struct {
	pb.UnimplementedPackageServiceServer
	service service.PackageService
	logger  *logrus.Logger
}

func NewGrpcPackageHandler(service service.PackageService, log *logrus.Logger) *GrpcPackageHandler {
	return &GrpcPackageHandler{
		logger: log,
	}
}

func (h *GrpcPackageHandler) GetPackage(ctx context.Context, req *pb.PackageID) (*pb.Package, error) {
	pkg, err := h.service.GetPackageByID(ctx, req.PackageId)
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
	pkgs, err := h.service.GetAllPackages(ctx, filter)
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
	pkgs, err := h.service.GetAllPackages(ctx, filter)
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
	}
	created, err := h.service.CreatePackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	return toProto(created), nil
}

func (h *GrpcPackageHandler) UpdatePackage(ctx context.Context, req *pb.Package) (*pb.Package, error) {
	update := models.PackageUpdate{
		Status:        req.Status,
		PaymentStatus: req.PaymentStatus,
	}
	updated, err := h.service.UpdatePackage(ctx, req.PackageId, update)
	if err != nil {
		return nil, err
	}
	return toProto(updated), nil
}

func (h *GrpcPackageHandler) DeletePackage(ctx context.Context, req *pb.PackageID) (*pb.Empty, error) {
	err := h.service.DeletePackage(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (h *GrpcPackageHandler) CancelPackage(ctx context.Context, req *pb.PackageID) (*pb.Package, error) {
	pkg, err := h.service.CancelPackage(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return toProto(pkg), nil
}

func (h *GrpcPackageHandler) GetPackageStatus(ctx context.Context, req *pb.PackageID) (*pb.PackageStatus, error) {
	pkg, err := h.service.GetPackageByID(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return &pb.PackageStatus{Status: pkg.Status}, nil
}

func (h *GrpcPackageHandler) GetExpiredPackages(ctx context.Context, req *pb.Empty) (*pb.PackageList, error) {
	pkgs, err := h.service.GetExpiredPackages(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoList(pkgs), nil
}

func (h *GrpcPackageHandler) MarkAsExpiredByID(ctx context.Context, req *pb.PackageID) (*pb.Package, error) {
	pkg, err := h.service.MarkPackageAsExpired(ctx, req.PackageId)
	if err != nil {
		return nil, err
	}
	return toProto(pkg), nil
}

func (h *GrpcPackageHandler) CreatePackageWithCalc(ctx context.Context, req *pb.Package) (*pb.Package, error) {
	if req.Weight <= 0 || req.From == "" || req.To == "" || req.Address == "" || req.Length <= 0 || req.Width <= 0 || req.Height <= 0 || req.TariffCode == "" {
		return nil, ErrInvalidInput
	}
	model := &models.Package{
		UserID:     req.UserId,
		Weight:     req.Weight,
		Length:     int(req.Length),
		Width:      int(req.Width),
		Height:     int(req.Height),
		From:       req.From,
		To:         req.To,
		Address:    req.Address,
		TariffCode: req.TariffCode,
	}
	created, err := h.service.CreatePackageWithCalculation(ctx, model)
	if err != nil {
		return nil, err
	}
	return toProto(created), nil
}

func (h *GrpcPackageHandler) TransferExpiredPackages(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	if err := h.service.TransferExpiredPackages(ctx); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
