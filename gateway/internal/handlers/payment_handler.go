package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type PaymentHandler struct {
	client *grpcclient.PaymentGRPCClient
	logger *logrus.Logger
}

func NewPaymentHandler(client *grpcclient.PaymentGRPCClient, logger *logrus.Logger) *PaymentHandler {
	return &PaymentHandler{client: client, logger: logger}
}

type ConfirmPaymentRequest struct {
	PackageID string `json:"package_id"`
}

func (h *PaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ConfirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request")
		return
	}

	message, err := h.client.ConfirmPayment(r.Context(), userID, req.PackageID)
	if err != nil {
		h.logger.Errorf("gRPC payment error: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Payment failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]string{"message": message})
}
