package middleware

import (
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/calculator/internal/transport"
	"github.com/sirupsen/logrus"
)

type Middleware func(http.Handler) http.Handler

type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares}
}

func (c *Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

func NewLogMiddleware(logger *logrus.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := transport.NewLoggingResponseWriter(w)

			defer func() {
				logger.WithFields(logrus.Fields{
					"method":   r.Method,
					"path":     r.URL.Path,
					"duration": time.Since(start).String(),
					"status":   lrw.StatusCode,
				}).Info("Request processed")
			}()

			next.ServeHTTP(lrw, r)
		})
	}
}
