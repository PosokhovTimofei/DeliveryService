package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/database/internal/models"
	"github.com/maksroxx/DeliveryService/database/internal/repository"
)

type PackageHandler struct {
	rep repository.RouteRepository
}

func NewPackageHandler(rep repository.RouteRepository) *PackageHandler {
	return &PackageHandler{
		rep: rep,
	}
}

func (h *PackageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/packages/{packageID}", h.GetPackage)
	mux.HandleFunc("/packages", h.GetAllPackages)
	mux.HandleFunc("/packages/{packageID}/status", h.GetPackageStatus)
	mux.HandleFunc("POST /packages", h.CreatePackage)
	mux.HandleFunc("PUT /packages/{packageID}", h.UpdatePackage)
	mux.HandleFunc("DELETE /packages/{packageID}", h.DeletePackage)
}

func (h *PackageHandler) GetPackage(w http.ResponseWriter, r *http.Request) {
	packageID := r.PathValue("packageID")
	if packageID == "" {
		respondWithError(w, http.StatusBadRequest, "Package id not found")
	}

	pkg, err := h.rep.GetByID(r.Context(), packageID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Package not found")
		return
	}

	respondWithJSON(w, http.StatusOK, pkg)
}

func (h *PackageHandler) GetAllPackages(w http.ResponseWriter, r *http.Request) {
	filter := models.RouteFilter{
		Status: r.URL.Query().Get("status"),
	}

	if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			filter.CreatedAfter = t
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.ParseInt(limit, 10, 64); err == nil {
			filter.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.ParseInt(offset, 10, 64); err == nil {
			filter.Offset = o
		}
	}

	packages, err := h.rep.GetAllRoutes(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, packages)
}

func (h *PackageHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	var req models.Package
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Weight <= 0 {
		http.Error(w, "Invalid weight", http.StatusBadRequest)
		return
	}
	if req.From == "" || req.To == "" || req.Address == "" {
		http.Error(w, "Invalid location", http.StatusBadRequest)
		return
	}

	pkg, err := h.rep.Create(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.ID = ""

	respondWithJSON(w, http.StatusCreated, pkg)
}

func (h *PackageHandler) GetPackageStatus(w http.ResponseWriter, r *http.Request) {
	packageID := r.PathValue("packageID")
	if packageID == "" {
		respondWithError(w, http.StatusBadRequest, "Package id not found")
	}

	pkg, err := h.rep.GetByID(r.Context(), packageID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Package not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": pkg.Status})
}

func (h *PackageHandler) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	packageID := r.PathValue("packageID")
	if packageID == "" {
		respondWithError(w, http.StatusBadRequest, "Package id not found")
	}

	var update models.RouteUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedPkg, err := h.rep.UpdateRoute(r.Context(), packageID, update)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update package")
		return
	}

	respondWithJSON(w, http.StatusOK, updatedPkg)
}

func (h *PackageHandler) DeletePackage(w http.ResponseWriter, r *http.Request) {
	packageID := r.PathValue("packageID")
	if packageID == "" {
		respondWithError(w, http.StatusBadRequest, "Package id not found")
	}

	if err := h.rep.DeleteRoute(r.Context(), packageID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete package")
		return
	}

	w.WriteHeader(http.StatusOK)
}
