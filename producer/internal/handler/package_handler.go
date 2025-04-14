package handler

import (
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	var pkg pkg.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if pkg.Weight <= 0 {
		http.Error(w, "Invalid weight", http.StatusBadRequest)
		return
	}
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		http.Error(w, "Invalid location", http.StatusBadRequest)
		return
	}

	// send to kafka
	createdPkg, err := h.service.CreatePackage(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PackageResponse{
		ID:             createdPkg.ID,
		Status:         createdPkg.Status,
		Cost:           createdPkg.Cost,
		EstimatedHours: createdPkg.EstimatedHours,
		Currency:       createdPkg.Currency,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
