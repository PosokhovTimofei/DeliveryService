package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
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

func (h *AuctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	packageID := r.URL.Query().Get("package_id")
	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid amount")
		return
	}

	resp, err := h.client.PlaceBid(userID, packageID, amount)
	if err != nil {
		h.logger.WithError(err).Error("PlaceBid failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "Bid failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) GetBidsByPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}
	packageID := r.URL.Query().Get("package_id")
	if packageID == "" {
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid package_id")
	}
	resp, err := h.client.GetBidsByPackage(userID, packageID)
	if err != nil {
		h.logger.WithError(err).Error("GetBidsByPackage failed")
		utils.RespondError(w, r, http.StatusInternalServerError, "GetBidsByPackage failed")
		return
	}
	utils.RespondJSON(w, r, http.StatusOK, resp)
}

func (h *AuctionHandler) WebSocketStream(w http.ResponseWriter, r *http.Request) {
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

	stream, cancel, err := h.client.StreamBids(userID, packageID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("stream error"))
		return
	}
	defer cancel()

	for {
		bid, err := stream.Recv()
		if err != nil {
			break
		}
		conn.WriteJSON(bid)
	}
}
