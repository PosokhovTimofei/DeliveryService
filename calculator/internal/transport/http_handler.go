package transport

import (
	"context"
	"encoding/json"
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
		RespondError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		metrics.CalculationFailureTotal.WithLabelValues("POST", "decode").Inc()
		return
	}

	if pkg.Weight <= 0 {
		RespondError(w, http.StatusBadRequest, "Invalid weight")
		metrics.CalculationFailureTotal.WithLabelValues("POST", "validation_weight").Inc()
		return
	}
	if Validate(pkg) != nil {
		RespondError(w, http.StatusBadRequest, "Invalid location data")
		metrics.CalculationFailureTotal.WithLabelValues("POST", "validation_location").Inc()
		return
	}

	result, err := h.service.Calculate(context.Background(), pkg)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Calculation error: "+err.Error())
		metrics.CalculationFailureTotal.WithLabelValues("POST", "calculation").Inc()
		return
	}

	metrics.CalculationSuccessTotal.WithLabelValues("POST").Inc()
	metrics.CalculatedCostValue.Observe(result.Cost)

	RespondJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) HandleCalculateByTariff(w http.ResponseWriter, r *http.Request) {
	var request struct {
		models.Package
		TariffCode string `json:"tariff_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	extCalc, ok := h.service.(*service.ExtendedCalculator)
	if !ok {
		RespondError(w, http.StatusInternalServerError, "Internal error: calculator not extended")
		return
	}

	result, err := extCalc.CalculateByTariffCode(context.Background(), request.Package, request.TariffCode)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Calculation error: "+err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) HandleTariffList(w http.ResponseWriter, r *http.Request) {
	extCalc, ok := h.service.(*service.ExtendedCalculator)
	if !ok {
		RespondError(w, http.StatusInternalServerError, "Internal error: calculator not extended")
		return
	}

	tariffs, err := extCalc.GetTariffs(context.Background())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch tariffs")
		return
	}

	RespondJSON(w, http.StatusOK, tariffs)
}

func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}
