package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	databasepb "github.com/maksroxx/DeliveryService/proto/database"
	"github.com/sirupsen/logrus"
)

type PackageHandler struct {
	client *grpcclient.PackageGRPCClient
	logger *logrus.Logger
}

func NewPackageHandler(client *grpcclient.PackageGRPCClient, logger *logrus.Logger) *PackageHandler {
	return &PackageHandler{
		client: client,
		logger: logger,
	}
}

func (h *PackageHandler) GetPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	packageID := r.URL.Query().Get("id")
	if packageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Missing package ID")
		return
	}

	pkg, err := h.client.GetPackage(userID, packageID)
	if err != nil {
		h.logger.Errorf("Failed to get package: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch package")
		return
	}

	utils.RespondJSON(w, http.StatusOK, pkg)
}

func (h *PackageHandler) GetAllPackages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	list, err := h.client.GetAllPackages(userID, status, limit, offset)
	if err != nil {
		h.logger.Errorf("Failed to get all packages: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch packages")
		return
	}

	utils.RespondJSON(w, http.StatusOK, list)
}

func (h *PackageHandler) GetAllUserPackages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}
	status := r.URL.Query().Get("status")
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	list, err := h.client.GetUserPackages(userID, status, limit, offset)
	if err != nil {
		h.logger.Errorf("Failed to get all packages: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch packages")
		return
	}

	utils.RespondJSON(w, http.StatusOK, list)
}

func (h *PackageHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	var pkg databasepb.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		h.logger.Errorf("Failed to decode package: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid package data")
		return
	}

	pkg.UserId = userID

	created, err := h.client.CreatePackage(userID, &pkg)
	if err != nil {
		h.logger.Errorf("Failed to create package: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create package")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, created)
}

func (h *PackageHandler) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	var pkg databasepb.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		h.logger.Errorf("Failed to decode update data: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid update data")
		return
	}

	updated, err := h.client.UpdatePackage(userID, &pkg)
	if err != nil {
		h.logger.Errorf("Failed to update package: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update package")
		return
	}

	utils.RespondJSON(w, http.StatusOK, updated)
}

func (h *PackageHandler) DeletePackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	packageID := r.URL.Query().Get("id")
	if packageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Missing package ID")
		return
	}

	_, err := h.client.DeletePackage(userID, packageID)
	if err != nil {
		h.logger.Errorf("Failed to delete package: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete package")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Package deleted"})
}

func (h *PackageHandler) CancelPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	packageID := r.URL.Query().Get("id")
	if packageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Missing package ID")
		return
	}

	cancelled, err := h.client.CancelPackage(userID, packageID)
	if err != nil {
		h.logger.Errorf("Failed to cancel package: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to cancel package")
		return
	}

	utils.RespondJSON(w, http.StatusOK, cancelled)
}

func (h *PackageHandler) GetPackageStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	packageID := r.URL.Query().Get("id")
	if packageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Missing package ID")
		return
	}

	status, err := h.client.GetPackageStatus(userID, packageID)
	if err != nil {
		h.logger.Errorf("Failed to get package status: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to get status")
		return
	}

	utils.RespondJSON(w, http.StatusOK, status)
}

func NewPackageHTTPHandler(handler *PackageHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/packages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Has("id") {
				handler.GetPackage(w, r)
			} else {
				handler.GetAllPackages(w, r)
			}
		case http.MethodPost:
			handler.CreatePackage(w, r)
		case http.MethodPut:
			handler.UpdatePackage(w, r)
		case http.MethodDelete:
			handler.DeletePackage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/packages/my", handler.GetAllUserPackages)
	mux.HandleFunc("/api/packages/cancel", handler.CancelPackage)
	mux.HandleFunc("/api/packages/status", handler.GetPackageStatus)

	return mux
}
