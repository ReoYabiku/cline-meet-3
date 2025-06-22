package handler

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Health returns the health status of the service
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ok",
		Message: "Signaling server is running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Ready returns the readiness status of the service
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you might check database connectivity, etc.
	response := HealthResponse{
		Status:  "ready",
		Message: "Signaling server is ready to accept connections",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
