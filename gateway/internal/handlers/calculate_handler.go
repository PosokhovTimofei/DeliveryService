package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
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
}

func (h *CalculateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "missing user id in context", http.StatusUnauthorized)
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.client.Calculate(req.Weight, userID, req.From, req.To, req.Address)
	if err != nil {
		h.logger.Errorf("Failed to call gRPC: %v", err)
		http.Error(w, "Failed to calculate cost", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"cost":            grpcResp.GetCost(),
		"estimated_hours": grpcResp.GetEstimatedHours(),
		"currency":        grpcResp.GetCurrency(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.logger.Errorf("Failed to marshal response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
