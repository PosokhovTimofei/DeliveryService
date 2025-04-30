package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maksroxx/DeliveryService/payment/internal/handler"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type mockRepo struct {
	updateFn func(ctx context.Context, p models.Payment) (*models.Payment, error)
	createFn func(ctx context.Context, p models.Payment) error
}

func (m *mockRepo) UpdatePayment(ctx context.Context, p models.Payment) (*models.Payment, error) {
	return m.updateFn(ctx, p)
}

func (m *mockRepo) CreatePayment(ctx context.Context, p models.Payment) error {
	return m.createFn(ctx, p)
}

type mockProducer struct {
	sendFn func(p models.Payment, userID string) error
}

func (m *mockProducer) PaymentMessage(p models.Payment, userID string) error {
	return m.sendFn(p, userID)
}

func (m mockProducer) Close() error {
	return nil
}

func TestConfirmPayment(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		header         string
		updateFn       func(ctx context.Context, p models.Payment) (*models.Payment, error)
		sendFn         func(p models.Payment, userID string) error
		wantStatusCode int
		wantBody       string
	}{
		{
			name:   "success",
			method: http.MethodPost,
			url:    "/payment/123",
			header: "user1",
			updateFn: func(ctx context.Context, p models.Payment) (*models.Payment, error) {
				return &p, nil
			},
			sendFn: func(p models.Payment, userID string) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "{\"message\":\"Payment confirmed and event sent\"}\n",
		},
		{
			name:           "missing user header",
			method:         http.MethodPost,
			url:            "/payment/123",
			header:         "",
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "{\"error\":\"Missing X-User-ID header\"}\n",
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			url:            "/paymen/123",
			header:         "user1",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "{\"error\":\"Method not allowed\"}\n",
		},
		{
			name:           "missing package ID",
			method:         http.MethodPost,
			url:            "/payment/",
			header:         "user1",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "{\"error\":\"Missing package ID in URL\"}\n",
		},
		{
			name:   "already confirmed",
			method: http.MethodPost,
			url:    "/payment/123",
			header: "user1",
			updateFn: func(ctx context.Context, p models.Payment) (*models.Payment, error) {
				return nil, errors.New("payment already confirmed")
			},
			wantStatusCode: http.StatusConflict,
			wantBody:       "{\"error\":\"payment already confirmed\"}\n",
		},
		{
			name:   "internal server error",
			method: http.MethodPost,
			url:    "/payment/123",
			header: "user1",
			updateFn: func(ctx context.Context, p models.Payment) (*models.Payment, error) {
				return nil, errors.New("db error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "{\"error\":\"db error\"}\n",
		},
		{
			name:   "kafka send error",
			method: http.MethodPost,
			url:    "/payment/123",
			header: "user1",
			updateFn: func(ctx context.Context, p models.Payment) (*models.Payment, error) {
				return &p, nil
			},
			sendFn: func(p models.Payment, userID string) error {
				return errors.New("kafka error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "{\"error\":\"kafka error\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{updateFn: tt.updateFn}
			prod := &mockProducer{sendFn: tt.sendFn}
			h := handler.NewPaymentHandler(repo, prod)

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(nil))
			if tt.header != "" {
				req.Header.Set("X-User-ID", tt.header)
			}
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != strings.TrimSpace(tt.wantBody) {
				t.Errorf("expected body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}
