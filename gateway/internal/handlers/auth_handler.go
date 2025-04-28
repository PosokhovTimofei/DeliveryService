package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/sirupsen/logrus"
)

type AuthHandlers struct {
	authClient *grpcclient.AuthGRPCClient
	logger     *logrus.Logger
}

func NewAuthHandlers(authClient *grpcclient.AuthGRPCClient, logger *logrus.Logger) *AuthHandlers {
	return &AuthHandlers{
		authClient: authClient,
		logger:     logger,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid register request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Register(req.Email, req.Password)
	if err != nil {
		h.logger.Error("gRPC register error: ", err)
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": resp.UserId,
		"token":   resp.Token,
	})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid login request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Login(req.Email, req.Password)
	if err != nil {
		h.logger.Error("gRPC login error: ", err)
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": resp.UserId,
		"token":   resp.Token,
	})
}
