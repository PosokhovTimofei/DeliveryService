package main

import (
	"net/http"

	"github.com/maksroxx/DeliveryService/calculator/internal/middleware"
	"github.com/maksroxx/DeliveryService/calculator/internal/service"
	"github.com/maksroxx/DeliveryService/calculator/internal/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := transport.Load()
	log := logrus.New()
	chain := middleware.NewChain(
		middleware.NewMetricsMiddleware(),
		middleware.NewLogMiddleware(log),
	)

	svc := service.NewCalculator()
	http.Handle("/metrics", promhttp.Handler())

	startHTTPServer(cfg.HTTPPort, svc, chain, log)
}

func startHTTPServer(port string, calc service.Calculator, chain *middleware.Chain, log *logrus.Logger) {
	handler := transport.NewHTTPHandler(calc)
	wrappedHandler := chain.Then(handler)

	http.Handle("/calculate", wrappedHandler)
	log.Infof("HTTP server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
