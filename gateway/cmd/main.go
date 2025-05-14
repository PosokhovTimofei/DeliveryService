package main

import (
	"net/http"
	"time"

	"github.com/maksroxx/DeliveryService/gateway/internal/grpcclient"
	"github.com/maksroxx/DeliveryService/gateway/internal/handlers"
	"github.com/maksroxx/DeliveryService/gateway/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := logrus.New()

	calculatorClient, err := grpcclient.NewCalculatorClient("localhost:50051")
	if err != nil {
		logger.Fatalf("Failed to connect to calculator gRPC: %v", err)
	}
	defer calculatorClient.Close()

	authClient, err := grpcclient.NewAuthGRPCClient("localhost:50052")
	if err != nil {
		logger.Fatalf("Failed to connect to auth gRPC: %v", err)
	}
	defer authClient.Close()

	paymentClient, err := grpcclient.NewPaymentGRPCClient("localhost:50053")
	if err != nil {
		logger.Fatalf("Failed to connect to payment gRPC: %v", err)
	}
	defer paymentClient.Close()

	// publicRoutes := []handlers.RouteConfig{
	// 	{
	// 		Prefix:      "/api/register",
	// 		TargetURL:   "http://localhost:1703",
	// 		PathRewrite: "/register",
	// 	},
	// 	{
	// 		Prefix:      "/api/login",
	// 		TargetURL:   "http://localhost:1703",
	// 		PathRewrite: "/login",
	// 	},
	// }

	protectedRoutes := []handlers.RouteConfig{
		{
			Prefix:      "/api/packages",
			TargetURL:   "http://localhost:8333",
			PathRewrite: "/packages",
		},
		// {
		//     Prefix: "/api/calculate",
		//     TargetURL: "http://localhost:8121",
		//     PathRewrite: "/calculate",
		// },
		{
			Prefix:      "/api/create",
			TargetURL:   "http://localhost:1234",
			PathRewrite: "/producer",
		},
		{
			Prefix:      "/api/profile",
			TargetURL:   "http://localhost:1704",
			PathRewrite: "/profile",
		},
		{
			Prefix:      "/api/my/packages",
			TargetURL:   "http://localhost:8333",
			PathRewrite: "/my/packages",
		},
		// {
		// 	Prefix:      "/api/payment",
		// 	TargetURL:   "http://localhost:5678",
		// 	PathRewrite: "/payment",
		// },
	}

	authHandlers := handlers.NewAuthHandlers(authClient, logger)
	// publicHandler := handlers.NewRouter(publicRoutes, logger)
	// publicChain := middleware.NewLogMiddleware(
	// 	enableCORS(publicHandler),
	// 	logger,
	// )

	protectedHandler := handlers.NewRouter(protectedRoutes, logger)
	protectedWithAuth := middleware.NewAuthMiddleware(
		enableCORS(protectedHandler),
		logger,
		authClient,
	)
	fullProtectedChain := middleware.NewLogMiddleware(
		protectedWithAuth,
		logger,
	)

	calculateHandler := handlers.NewCalculateHandler(calculatorClient, logger)
	calculateWithAuth := middleware.NewAuthMiddleware(
		enableCORS(calculateHandler),
		logger,
		authClient,
	)
	calculateChain := middleware.NewLogMiddleware(
		calculateWithAuth,
		logger,
	)

	paymentHandler := handlers.NewPaymentHandler(paymentClient, logger)
	paymentWithAuth := middleware.NewAuthMiddleware(
		enableCORS(paymentHandler),
		logger,
		authClient,
	)
	paymentChain := middleware.NewLogMiddleware(paymentWithAuth, logger)

	calculateByTariffHandler := handlers.NewCalculateByTariffHandler(calculatorClient, logger)
	calculateByTariffWithAuth := middleware.NewAuthMiddleware(
		enableCORS(calculateByTariffHandler),
		logger,
		authClient,
	)
	calculateByTariffChain := middleware.NewLogMiddleware(
		calculateByTariffWithAuth,
		logger,
	)
	tariffListHandler := handlers.NewTariffListHandler(calculatorClient, logger)
	tariffListWithAuth := middleware.NewAuthMiddleware(
		enableCORS(tariffListHandler),
		logger,
		authClient,
	)
	tariffListChain := middleware.NewLogMiddleware(
		tariffListWithAuth,
		logger,
	)
	http.Handle("/api/payment/confirm", paymentChain)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/api/register", middleware.NewLogMiddleware(
		enableCORS(http.HandlerFunc(authHandlers.Register)),
		logger,
	))
	http.Handle("/api/register-moderator", middleware.NewLogMiddleware(
		enableCORS(http.HandlerFunc(authHandlers.RegisterModerator)),
		logger,
	))

	http.Handle("/api/login", middleware.NewLogMiddleware(
		enableCORS(http.HandlerFunc(authHandlers.Login)),
		logger,
	))
	// http.Handle("/api/register", publicChain)
	// http.Handle("/api/login", publicChain)
	http.Handle("/api/calculate", calculateChain)
	http.Handle("/api/calculate_by_tariff", calculateByTariffChain)
	http.Handle("/api/tariffs", tariffListChain)
	http.Handle("/api/", fullProtectedChain)

	logger.Info("Starting API Gateway on :8228")
	server := &http.Server{
		Addr:              ":8228",
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}
