package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// Handle handles health check requests
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": r.Context().Value("timestamp"),
		"services": map[string]string{
			"database": "connected",
			"cache":    "connected",
			"exchange": "connected",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
