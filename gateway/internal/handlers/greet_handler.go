package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
)

type DefaultHandler struct{}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, r, 200, map[string]string{"message": "Welcome to DeliveryService API Gateway"})
}
