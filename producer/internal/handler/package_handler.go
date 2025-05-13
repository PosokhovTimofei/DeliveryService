package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/producer/internal/service"
	"github.com/maksroxx/DeliveryService/producer/pkg"
)

type PackageHandler struct {
	service *service.PackageService
}

func NewPackageHandler(svc *service.PackageService) *PackageHandler {
	return &PackageHandler{service: svc}
}

func (h *PackageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

type PackageResponse struct {
	ID             string  `json:"package_id"`
	Status         string  `json:"status"`
	Cost           float64 `json:"cost"`
	EstimatedHours int     `json:"estimated_hours"`
	Currency       string  `json:"currency"`
}

func (h *PackageHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}
	var pkg pkg.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if pkg.Weight <= 0 {
		RespondError(w, http.StatusBadRequest, "Invalid weight")
		return
	}
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		RespondError(w, http.StatusBadRequest, "Invalid location fields")
		return
	}
	if pkg.Length <= 0 || pkg.Height <= 0 || pkg.Width <= 0 {
		RespondError(w, http.StatusBadRequest, "Invalid parameters fields")
		return
	}

	// send to kafka
	createdPkg, err := h.service.CreatePackage(context.Background(), pkg, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create package: "+err.Error())
		return
	}

	response := PackageResponse{
		ID:             createdPkg.ID,
		Status:         createdPkg.Status,
		Cost:           createdPkg.Cost,
		EstimatedHours: createdPkg.EstimatedHours,
		Currency:       createdPkg.Currency,
	}

	RespondJSON(w, http.StatusCreated, response)
}
