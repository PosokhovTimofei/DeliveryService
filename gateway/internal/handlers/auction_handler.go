package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/metrics"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type AuctionHandler struct {
	client *grpcclient.AuctionGRPCClient
	logger *logrus.Logger
}

func NewAuctionHandler(client *grpcclient.AuctionGRPCClient, log *logrus.Logger) *AuctionHandler {
	return &AuctionHandler{
		client: client,
		logger: log,
	}
}

type placeBidRequest struct {
	PackageID string  `json:"package_id"`
	Amount    float64 `json:"amount"`
}

func (h *AuctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req placeBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PackageID == "" || req.Amount <= 0 {
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid package_id or amount")
		return
	}

	resp, err := h.client.PlaceBid(userID, req.PackageID, req.Amount)
	if err != nil {
		h.logger.WithError(err).Error("PlaceBid failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "Bid failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, resp)
}

type getBidsRequest struct {
	PackageID string `json:"package_id"`
}

func (h *AuctionHandler) GetBidsByPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req getBidsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PackageID == "" {
		utils.RespondError(w, r, http.StatusBadRequest, "Missing package_id")
		return
	}

	resp, err := h.client.GetBidsByPackage(userID, req.PackageID)
	if err != nil {
		h.logger.WithError(err).Error("GetBidsByPackage failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "GetBidsByPackage failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) GetAuctioningPackages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp, err := h.client.GetAuctioningPackages(userID)
	if err != nil {
		h.logger.WithError(err).Error("GetAuctioningPackages failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "GetAuctioningPackages failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) GetFailedPackages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp, err := h.client.GetFailedPackages(userID)
	if err != nil {
		h.logger.WithError(err).Error("GetFailedPackages failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "GetFailedPackages failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) GetUserWonPackages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp, err := h.client.GetUserWonPackages(userID)
	if err != nil {
		h.logger.WithError(err).Error("GetUserWonPackages failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "GetUserWonPackages failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) StartAuction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	_, err := h.client.StartAuction(userID)
	if err != nil {
		h.logger.WithError(err).Error("StartAuction failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "StartAuction failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, map[string]string{"status": "started"})
}

func (h *AuctionHandler) RepeateAuction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	_, err := h.client.RepeateAuction(userID)
	if err != nil {
		h.logger.WithError(err).Error("RepeateAuction failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "RepeateAuction failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, map[string]string{"status": "started"})
}

func (h *AuctionHandler) WebSocketStream(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	packageID := r.URL.Query().Get("package_id")
	userID := r.URL.Query().Get("user_id")

	if packageID == "" || userID == "" {
		http.Error(w, "Missing package_id or user_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	metrics.WSConnectionsTotal.Inc()
	metrics.WSActiveConnections.Inc()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.WSConnectionDuration.Observe(duration)
		metrics.WSDisconnectionsTotal.Inc()
		metrics.WSActiveConnections.Dec()
	}()

	stream, cancel, err := h.client.StreamBids(userID, packageID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("stream error"))
		return
	}
	defer cancel()

	for {
		bid, err := stream.Recv()
		if err != nil {
			h.logger.WithError(err).Error("Stream receive failed")
			break
		}

		metrics.WSMessagesReceivedTotal.Inc()

		err = conn.WriteJSON(bid)
		if err != nil {
			h.logger.WithError(err).Error("WebSocket send failed")
			break
		}

		metrics.WSMessagesSentTotal.Inc()
	}
}
