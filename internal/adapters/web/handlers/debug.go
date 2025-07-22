package handlers

import (
	"encoding/json"
	//"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"marketflow/internal/application/usecases"
)

// DebugHandler handles debug requests
type DebugHandler struct {
	marketDataUseCase *usecases.MarketDataUseCase
	logger            *slog.Logger
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(marketDataUseCase *usecases.MarketDataUseCase, logger *slog.Logger) *DebugHandler {
	return &DebugHandler{
		marketDataUseCase: marketDataUseCase,
		logger:            logger,
	}
}

// Handle handles debug requests
func (h *DebugHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/debug/")

	switch path {
	case "redis":
		h.handleRedisDebug(w, r)
	case "postgres":
		h.handlePostgresDebug(w, r)
	default:
		http.Error(w, "Unknown debug endpoint", http.StatusBadRequest)
	}
}

func (h *DebugHandler) handleRedisDebug(w http.ResponseWriter, r *http.Request) {
	// This is a simple debug endpoint - in a real system you'd want proper debugging
	response := map[string]interface{}{
		"message":   "Redis debug - check server logs for cache operations",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DebugHandler) handlePostgresDebug(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "Check PostgreSQL with: docker exec marketflow-postgres-1 psql -U marketflow -d marketflow -c \"SELECT COUNT(*) FROM market_data;\"",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
