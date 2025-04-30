package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/kafka"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type PaymentHandler struct {
	repo     db.Paymenter
	producer kafka.Producerer
}

func NewPaymentHandler(repo db.Paymenter, producer kafka.Producerer) *PaymentHandler {
	return &PaymentHandler{repo: repo, producer: producer}
}

func (h *PaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/payment/") {
		h.ConfirmPayment(w, r)
		return
	}
	RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func (h *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		RespondError(w, http.StatusBadRequest, "Missing package ID in URL")
		return
	}
	packageID := parts[2]

	updatedPayment, err := h.repo.UpdatePayment(context.Background(), models.Payment{
		UserID:    userID,
		PackageID: packageID,
		Status:    "PAID",
	})
	if err != nil {
		if err.Error() == "payment already confirmed" {
			RespondError(w, http.StatusConflict, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.producer.PaymentMessage(*updatedPayment, updatedPayment.UserID); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Payment confirmed and event sent",
	})
}

func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}
