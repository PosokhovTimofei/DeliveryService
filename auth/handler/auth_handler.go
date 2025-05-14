package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/maksroxx/DeliveryService/auth/metrics"
	"github.com/maksroxx/DeliveryService/auth/middleware"
	"github.com/maksroxx/DeliveryService/auth/models"
	"github.com/maksroxx/DeliveryService/auth/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		duration := time.Since(start).Seconds()
		metrics.HTTPResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(w.(*middleware.LoggingResponseWriter).Status),
		).Observe(duration)
	}(time.Now())

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if !validateEmail(req.Email) {
		RespondError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if len(req.Password) < 3 {
		RespondError(w, http.StatusBadRequest, "password must be at least 3 characters")
		return
	}

	user, token, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		RespondError(w, getStatusCode(err), err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) RegisterModerator(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		duration := time.Since(start).Seconds()
		metrics.HTTPResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(w.(*middleware.LoggingResponseWriter).Status),
		).Observe(duration)
	}(time.Now())

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if !validateEmail(req.Email) {
		RespondError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if len(req.Password) < 3 {
		RespondError(w, http.StatusBadRequest, "password must be at least 3 characters")
		return
	}

	user, token, err := h.service.RegisterModerator(r.Context(), req.Email, req.Password)
	if err != nil {
		RespondError(w, getStatusCode(err), err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		duration := time.Since(start).Seconds()
		metrics.HTTPResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(w.(*middleware.LoggingResponseWriter).Status),
		).Observe(duration)
	}(time.Now())

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if !validateEmail(req.Email) {
		RespondError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		RespondError(w, getStatusCode(err), err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"token": token, "role": user.Role})
}

func validateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}

func getStatusCode(err error) int {
	switch err {
	case models.ErrEmailAlreadyExists:
		return http.StatusConflict
	case models.ErrInvalidCredentials:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
