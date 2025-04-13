package handlers

import (
	"net/http"
	"strings"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type PackageHandler struct {
	baseURL string
	logger  *logrus.Logger
}

func NewPackageHandler(baseURL string, logger *logrus.Logger) *PackageHandler {
	return &PackageHandler{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		logger:  logger,
	}
}

func (h *PackageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Формируем целевой URL
	path := strings.TrimPrefix(r.URL.Path, "/api")
	targetURL := h.baseURL + path

	// Добавляем query-параметры
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	h.logger.Debugf("Proxying to: %s", targetURL)

	if err := utils.ProxyRequest(w, r, targetURL); err != nil {
		h.logger.Errorf("Proxy error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
