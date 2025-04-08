package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type StatusHandler struct {
	proxyURL string
	logger   *logrus.Logger
}

func NewStatusHandler(proxyURL string, logger *logrus.Logger) *StatusHandler {
	return &StatusHandler{
		proxyURL: proxyURL,
		logger:   logger,
	}
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetURL := h.proxyURL + r.URL.Path[len("/api/status"):]
	if err := utils.ProxyRequest(w, r, targetURL); err != nil {
		h.logger.Error("Error processing status request: ", err)
	}
}
