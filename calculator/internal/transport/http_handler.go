package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		return
	}

	if pkg.Weight <= 0 {
		http.Error(w, "Invalid weight", http.StatusBadRequest)
		return
	}
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		http.Error(w, "Invalid location", http.StatusBadRequest)
		return
	}

	result, err := h.service.Calculate(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
