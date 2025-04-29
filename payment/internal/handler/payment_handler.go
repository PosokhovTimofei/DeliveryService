package handler

import (
	"context"
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
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Missing X-User-ID header", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing package ID in URL", http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.producer.PaymentMessage(*updatedPayment, updatedPayment.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment confirmed and event sent"))
}
