package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type CalculateByTariffHandler struct {
	client *grpcclient.CalculatorGRPCClient
	logger *logrus.Logger
}

func NewCalculateByTariffHandler(client *grpcclient.CalculatorGRPCClient, logger *logrus.Logger) *CalculateByTariffHandler {
	return &CalculateByTariffHandler{client: client, logger: logger}
}

type CalculateByTariffRequest struct {
	Weight     float64 `json:"weight"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	Address    string  `json:"address"`
	Length     int     `json:"length"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	TariffCode string  `json:"tariff_code"`
}

func (h *CalculateByTariffHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Missing user ID")
		return
	}

	var req CalculateByTariffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.client.CalculateByTariffCode(req.Weight, userID, req.From, req.To, req.Address, req.TariffCode, req.Length, req.Width, req.Height)
	if err != nil {
		h.logger.Errorf("Failed to calculate by tariff code: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Calculation failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]any{
		"cost":            resp.GetCost(),
		"estimated_hours": resp.GetEstimatedHours(),
		"currency":        resp.GetCurrency(),
	})
}
