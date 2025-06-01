package handlers

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func RegisterRoutes(
	mux *http.ServeMux,
	logger *logrus.Logger,
	authClient *grpcclient.AuthGRPCClient,
	calculatorClient *grpcclient.CalculatorGRPCClient,
	paymentClient *grpcclient.PaymentGRPCClient,
	packageClient *grpcclient.PackageGRPCClient,
	auctionClient *grpcclient.AuctionGRPCClient,
) {
	// Default
	defaultHandler := NewDefaultHandler()
	mux.Handle("/api", logAndCORS(defaultHandler, logger))

	// Auth
	authHandlers := NewAuthHandlers(authClient, logger)
	mux.Handle("/api/register", logAndCORS(http.HandlerFunc(authHandlers.Register), logger))
	mux.Handle("/api/register-moderator", logAndCORS(http.HandlerFunc(authHandlers.RegisterModerator), logger))
	mux.Handle("/api/login", logAndCORS(http.HandlerFunc(authHandlers.Login), logger))

	// Calculate
	calcHandler := NewCalculateHandler(calculatorClient, logger)
	mux.Handle("/api/calculate", protectAndLog(calcHandler, authClient, logger))
	mux.Handle("/api/calculate-by-tariff", protectAndLog(NewCalculateByTariffHandler(calculatorClient, logger), authClient, logger))
	mux.Handle("/api/tariffs", protectAndLog(NewTariffListHandler(calculatorClient, logger), authClient, logger))
	mux.Handle("/api/tariff", protectAndLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			calcHandler.CreateTariff(w, r)
		case http.MethodDelete:
			calcHandler.DeleteTariff(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}), authClient, logger))

	// Auction
	auctionHandler := NewAuctionHandler(auctionClient, logger)
	mux.Handle("/api/auction/place", protectAndLog(http.HandlerFunc(auctionHandler.PlaceBid), authClient, logger))
	// ws://localhost:8228/api/auction/ws
	mux.Handle("/api/auction/ws", http.HandlerFunc(auctionHandler.WebSocketStream))
	mux.Handle("/api/auction/bids", protectAndLog(http.HandlerFunc(auctionHandler.GetBidsByPackage), authClient, logger))

	// Payment
	mux.Handle("/api/payment/confirm", protectAndLog(NewPaymentHandler(paymentClient, logger), authClient, logger))

	// Packages
	// POST /packages/cancel?id=xxx
	// GET /packages/status?id=xxx
	// DELETE /packages?id=xxx
	// PUT /packages (json body)
	// POST /packages (json body)
	// GET /packages/all?status=delivered&limit=10&offset=0
	// GET /packages?id=xxx
	// GET /packages/my?status=delivered&limit=10&offset=0
	packageHandler := NewPackageHandler(packageClient, logger)
	mux.Handle("/api/packages", protectAndLog(NewPackageHTTPHandler(packageHandler), authClient, logger))
	mux.Handle("/api/packages/", protectAndLog(NewPackageHTTPHandler(packageHandler), authClient, logger))

	// Metrics
	mux.Handle("/metrics", promhttp.Handler())

	protectedRoutes := []RouteConfig{
		{
			Prefix:      "/api/profile",
			TargetURL:   "http://localhost:1704",
			PathRewrite: "/profile",
		},
	}
	proxyRouter := NewRouter(protectedRoutes, logger)
	handlerWithMiddleware := protectAndLog(proxyRouter, authClient, logger)
	mux.Handle("/api/", handlerWithMiddleware)
}
