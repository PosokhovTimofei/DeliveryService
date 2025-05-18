package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
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
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authClient.Register(req.Email, req.Password)
	if err != nil {
		h.logger.Error("gRPC register error: ", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Registration failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]string{
		"user_id": resp.UserId,
		"token":   resp.Token,
		"role":    resp.Role,
	})
}

func (h *AuthHandlers) RegisterModerator(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid register request: ", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authClient.RegisterModerator(req.Email, req.Password)
	if err != nil {
		h.logger.Error("gRPC register error: ", err)
		utils.RespondError(w, r, http.StatusInternalServerError, "Registration failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]string{
		"user_id": resp.UserId,
		"token":   resp.Token,
		"role":    resp.Role,
	})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid login request: ", err)
		utils.RespondError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authClient.Login(req.Email, req.Password)
	if err != nil {
		h.logger.Error("gRPC login error: ", err)
		utils.RespondError(w, r, http.StatusUnauthorized, "Login failed")
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, map[string]string{
		"user_id": resp.UserId,
		"token":   resp.Token,
		"role":    resp.Role,
	})
}
