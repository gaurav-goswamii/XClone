package controllers

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response payload
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// HealthCheckHandler handles API health checks
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Service: "xclone-backend",
	})
}
