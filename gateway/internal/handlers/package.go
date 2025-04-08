package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type PackageHandler struct {
	proxyURL string
	logger   *logrus.Logger
}

func NewPackageHandler(proxyUrl string, logger *logrus.Logger) *PackageHandler {
	return &PackageHandler{
		proxyURL: proxyUrl,
		logger:   logger,
	}
}

func (h *PackageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := utils.ProxyRequest(w, r, h.proxyURL); err != nil {
		h.logger.Error("Error processing package request: ", err)
	}
}
