package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type CalculateHandler struct {
	client *grpcclient.CalculatorGRPCClient
	logger *logrus.Logger
}

func NewCalculateHandler(client *grpcclient.CalculatorGRPCClient, logger *logrus.Logger) *CalculateHandler {
	return &CalculateHandler{
		client: client,
		logger: logger,
	}
}

type calculateRequest struct {
	Weight  float64 `json:"weight"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Address string  `json:"address"`
	Length  int     `json:"length"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
}

func (h *CalculateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Missing user ID")
		return
	}

	var req calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	grpcResp, err := h.client.Calculate(req.Weight, userID, req.From, req.To, req.Address, req.Length, req.Width, req.Height)
	if err != nil {
		h.logger.Errorf("Failed to call gRPC: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Failed to calculate cost")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]any{
		"cost":            grpcResp.GetCost(),
		"estimated_hours": grpcResp.GetEstimatedHours(),
		"currency":        grpcResp.GetCurrency(),
	})
}

type tariff struct {
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	BaseRate          float64 `json:"base_rate"`
	PricePerKm        float64 `json:"price_per_km"`
	PricePerKg        float64 `json:"price_per_kg"`
	Currency          string  `json:"currency"`
	VolumetricDivider float64 `json:"volumetric_divider"`
	SpeedKmph         float64 `json:"speed_kmph"`
}

func (h *CalculateHandler) CreateTariff(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Missing user ID")
		return
	}
	var req tariff
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	gprcResp, err := h.client.CreateTariff(userID, req.Code, req.Name, req.Currency, req.BaseRate, req.PricePerKm, req.PricePerKg, req.VolumetricDivider, req.SpeedKmph)
	if err != nil {
		h.logger.Errorf("Failed to call gRPC: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Failed to create tariff")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, gprcResp)
}

type deleteTariff struct {
	Code string `json:"code"`
}

func (h *CalculateHandler) DeleteTariff(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Missing user ID")
		return
	}
	var req deleteTariff
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	_, err := h.client.DeleteTariff(userID, req.Code)
	if err != nil {
		h.logger.Errorf("Failed to call gRPC: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Failed to delete tariff")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}
