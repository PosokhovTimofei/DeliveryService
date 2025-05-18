package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type TariffListHandler struct {
	client *grpcclient.CalculatorGRPCClient
	logger *logrus.Logger
}

func NewTariffListHandler(client *grpcclient.CalculatorGRPCClient, logger *logrus.Logger) *TariffListHandler {
	return &TariffListHandler{client: client, logger: logger}
}

func (h *TariffListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.RespondError(w, r, http.StatusUnauthorized, "Missing user ID")
		return
	}

	resp, err := h.client.GetTariffList(userID)
	if err != nil {
		h.logger.Errorf("Failed to fetch tariffs: %v", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Failed to fetch tariffs")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, resp.Tariffs)
}
