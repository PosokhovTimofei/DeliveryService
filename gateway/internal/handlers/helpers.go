package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/sirupsen/logrus"
)

func logAndCORS(h http.Handler, logger *logrus.Logger) http.Handler {
	return middleware.MetricsMiddleware(
		middleware.NewLogMiddleware(
			middleware.NewCORSMiddleware(h),
			logger,
		),
	)
}

func protectAndLog(h http.Handler, authClient interface{}, logger *logrus.Logger) http.Handler {
	authGRPCClient := authClient.(*grpcclient.AuthGRPCClient)

	return middleware.MetricsMiddleware(
		middleware.NewLogMiddleware(
			middleware.NewCORSMiddleware(
				middleware.NewAuthMiddleware(
					h,
					logger,
					authGRPCClient,
				),
			),
			logger,
		),
	)
}
