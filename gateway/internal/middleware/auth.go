package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/metrics"
	"github.com/maksroxx/DeliveryService/gateway/internal/utils"
	"github.com/sirupsen/logrus"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userIDContextKey).(string)
	return val, ok
}

type AuthMiddleware struct {
	next       http.Handler
	logger     *logrus.Logger
	authClient *grpcclient.AuthGRPCClient
}

func NewAuthMiddleware(next http.Handler, logger *logrus.Logger, authClient *grpcclient.AuthGRPCClient) *AuthMiddleware {
	return &AuthMiddleware{
		next:       next,
		logger:     logger,
		authClient: authClient,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lrw := utils.NewLoggingResponseWriter(w)

	if r.Method == "OPTIONS" {
		m.next.ServeHTTP(lrw, r)
		m.observeMetrics(r, lrw.StatusCode, start)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/register") || strings.HasPrefix(r.URL.Path, "/api/login") || strings.HasPrefix(r.URL.Path, "/api/register-moderator") {
		m.next.ServeHTTP(lrw, r)
		m.observeMetrics(r, lrw.StatusCode, start)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		m.logger.Warn("Missing authorization token")
		utils.RespondError(lrw, r, http.StatusUnauthorized, "Invalid token")
		m.observeMetrics(r, http.StatusUnauthorized, start)
		return
	}

	userID, valid := m.validateToken(token)
	if !valid {
		m.logger.Warn("Invalid token")
		utils.RespondError(lrw, r, http.StatusUnauthorized, "Invalid token")
		m.observeMetrics(r, http.StatusUnauthorized, start)
		return
	}

	ctx := context.WithValue(r.Context(), userIDContextKey, userID)
	r = r.WithContext(ctx)
	m.next.ServeHTTP(lrw, r)
	m.observeMetrics(r, lrw.StatusCode, start)
}

func (m *AuthMiddleware) validateToken(token string) (string, bool) {
	resp, err := m.authClient.Validate(token)
	if err != nil {
		m.logger.Errorf("Failed to validate token: %v", err)
		return "", false
	}

	if resp.Valid != "ok" {
		m.logger.Warnf("Token validation failed: valid=%s", resp.Valid)
		return "", false
	}

	return resp.UserId, true
}

func (m *AuthMiddleware) observeMetrics(r *http.Request, statusCode int, start time.Time) {
	duration := time.Since(start).Seconds()
	metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Inc()
	metrics.HttpResponseTimeSeconds.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Observe(duration)
}
