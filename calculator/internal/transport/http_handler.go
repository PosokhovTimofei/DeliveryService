package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maksroxx/DeliveryService/calculator/internal/metrics"
	"github.com/maksroxx/DeliveryService/calculator/internal/service"
	"github.com/maksroxx/DeliveryService/calculator/models"
)

type HTTPHandler struct {
	service service.Calculator
}

func NewHTTPHandler(s service.Calculator) *HTTPHandler {
	return &HTTPHandler{service: s}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var pkg models.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		metrics.CalculationFailureTotal.WithLabelValues("POST", "decode").Inc()
		return
	}

	if pkg.Weight <= 0 {
		http.Error(w, "Invalid weight", http.StatusBadRequest)
		metrics.CalculationFailureTotal.WithLabelValues("POST", "validation_weight").Inc()
		return
	}
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		http.Error(w, "Invalid location", http.StatusBadRequest)
		metrics.CalculationFailureTotal.WithLabelValues("POST", "validation_location").Inc()
		return
	}

	result, err := h.service.Calculate(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		metrics.CalculationFailureTotal.WithLabelValues("POST", "calculation").Inc()
		return
	}
	fmt.Println(result)
	metrics.CalculationSuccessTotal.WithLabelValues("POST").Inc()
	metrics.CalculatedCostValue.Observe(result.Cost)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
