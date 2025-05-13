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

type CalculateRequest struct {
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
		utils.RespondError(w, http.StatusUnauthorized, "Missing user ID")
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	grpcResp, err := h.client.Calculate(req.Weight, userID, req.From, req.To, req.Address, req.Length, req.Width, req.Height)
	if err != nil {
		h.logger.Errorf("Failed to call gRPC: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to calculate cost")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]any{
		"cost":            grpcResp.GetCost(),
		"estimated_hours": grpcResp.GetEstimatedHours(),
		"currency":        grpcResp.GetCurrency(),
	})
}
