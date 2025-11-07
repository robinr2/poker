package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthCheckHandler returns an HTTP handler function for the health check endpoint.
func HealthCheckHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{
			"status": "ok",
		}

		json.NewEncoder(w).Encode(response)
	}
}
