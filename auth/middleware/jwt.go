package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/maksroxx/DeliveryService/auth/metrics"
	"github.com/maksroxx/DeliveryService/auth/service"
	"github.com/sirupsen/logrus"
)

const UserIDKey = "user_id"

func JWTAuth(svc *service.AuthService, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

			defer func() {
				duration := time.Since(start).Seconds()
				metrics.HTTPResponseTime.WithLabelValues(
					r.Method,
					r.URL.Path,
					strconv.Itoa(w.(*LoggingResponseWriter).Status),
				).Observe(duration)
			}()

			if token == "" {
				metrics.ValidateFailureTotal.WithLabelValues(r.Method, "missing_token").Inc()
				logger.Warn("Missing authorization token")
				http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			claims, err := svc.ValidateToken(token)
			if err != nil {
				metrics.ValidateFailureTotal.WithLabelValues(r.Method, "invalid_token").Inc()
				logger.WithError(err).Warn("Invalid token")
				http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
				return
			}

			metrics.ValidateSuccessTotal.WithLabelValues(r.Method).Inc()
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
