package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"marketflow/internal/application/usecases"
)

// StatusHandler handles status requests
type StatusHandler struct {
	dataProcessingUseCase *usecases.DataProcessingUseCase
	logger                *slog.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(dataProcessingUseCase *usecases.DataProcessingUseCase, logger *slog.Logger) *StatusHandler {
	return &StatusHandler{
		dataProcessingUseCase: dataProcessingUseCase,
		logger:                logger,
	}
}

// Handle handles status requests
func (h *StatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentMode := h.dataProcessingUseCase.GetMode()

	response := map[string]interface{}{
		"current_mode": string(currentMode),
		"available_modes": []string{"live", "test"},
		"status": "running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
