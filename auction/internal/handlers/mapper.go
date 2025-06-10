package handlers

import (
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/models"
	auctionpb "github.com/maksroxx/DeliveryService/proto/auction"
)

func BidModelToProto(b *models.Bid) *auctionpb.Bid {
	return &auctionpb.Bid{
		BidId:     b.BidID,
		PackageId: b.PackageID,
		UserId:    b.UserID,
		Amount:    b.Amount,
		Timestamp: b.Timestamp.Format(time.RFC3339),
	}
}

func BidProtoToModel(req *auctionpb.BidRequest) *models.Bid {
	return &models.Bid{
		PackageID: req.PackageId,
		UserID:    req.UserId,
		Amount:    req.Amount,
		Timestamp: time.Now(),
	}
}

func PackageModelToProto(p *models.Package) *auctionpb.Package {
	return &auctionpb.Package{
		PackageId:  p.PackageID,
		Status:     p.Status,
		From:       p.From,
		To:         p.To,
		Weight:     p.Weight,
		Length:     int32(p.Length),
		Width:      int32(p.Width),
		Height:     int32(p.Height),
		Cost:       p.Cost,
		Currency:   p.Currency,
		TariffCode: p.TariffCode,
	}
}

func PackagesModelToProto(pkgs []models.Package) *auctionpb.Packages {
	var packagesProto auctionpb.Packages
	for _, p := range pkgs {
		pkg := PackageModelToProto(&p)
		packagesProto.Package = append(packagesProto.Package, pkg)
	}
	return &packagesProto
}
