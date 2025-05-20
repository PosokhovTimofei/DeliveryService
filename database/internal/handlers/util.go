package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	pb "github.com/maksroxx/DeliveryService/proto/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func toProto(p *models.Package) *pb.Package {
	return &pb.Package{
		PackageId:      p.PackageID,
		UserId:         p.UserID,
		Weight:         p.Weight,
		Length:         int32(p.Length),
		Width:          int32(p.Width),
		Height:         int32(p.Height),
		From:           p.From,
		To:             p.To,
		Address:        p.Address,
		PaymentStatus:  p.PaymentStatus,
		Status:         p.Status,
		Cost:           p.Cost,
		EstimatedHours: int32(p.EstimatedHours),
		RemainingHours: int32(p.RemainingHours),
		Currency:       p.Currency,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		TariffCode:     p.TariffCode,
	}
}

func toProtoList(list []*models.Package) *pb.PackageList {
	out := &pb.PackageList{}
	for _, p := range list {
		out.Packages = append(out.Packages, toProto(p))
	}
	return out
}
