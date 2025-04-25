package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	next        http.Handler
	logger      *logrus.Logger
	validateURL string
	httpClient  *http.Client
}

func NewAuthMiddleware(next http.Handler, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		next:        next,
		logger:      logger,
		validateURL: "http://localhost:1704/validate",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		m.next.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/register") {
		m.next.ServeHTTP(w, r)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		m.logger.Warn("Missing authorization token")
		http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
		return
	}

	userID, valid := m.validateToken(token)
	if valid != "ok" {
		m.logger.Warn("Invalid token")
		http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
		return
	}

	r.Header.Set("X-User-ID", userID)
	m.next.ServeHTTP(w, r)
}

func (m *AuthMiddleware) validateToken(token string) (string, string) {
	req, err := http.NewRequest("GET", m.validateURL, nil)
	if err != nil {
		m.logger.Error("Error creating request:", err)
		return "", ""
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Error("Error making validation request:", err)
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Warnf("Validation service returned status: %d", resp.StatusCode)
		return "", ""
	}

	var response struct {
		Valid  string `json:"status"`
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		m.logger.Error("Error decoding response:", err)
		return "", ""
	}

	return response.UserID, response.Valid
}
