package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	next   http.Handler
	logger *logrus.Logger
}

func NewAuthMiddleware(next http.Handler, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		next:   next,
		logger: logger,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		m.logger.Warn("Unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// soon will be added check
	m.next.ServeHTTP(w, r)
}
