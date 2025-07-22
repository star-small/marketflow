package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"marketflow/internal/application/usecases"
	"marketflow/internal/domain/models"
)

// ModeHandler handles mode switching requests
type ModeHandler struct {
	dataProcessingUseCase *usecases.DataProcessingUseCase
	logger                *slog.Logger
}

// NewModeHandler creates a new mode handler
func NewModeHandler(dataProcessingUseCase *usecases.DataProcessingUseCase, logger *slog.Logger) *ModeHandler {
	return &ModeHandler{
		dataProcessingUseCase: dataProcessingUseCase,
		logger:                logger,
	}
}

// Handle handles mode switching requests
func (h *ModeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Mode handler called", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodPost {
		h.logger.Warn("Invalid method for mode endpoint", "method", r.Method)
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/mode/")
	h.logger.Info("Processing mode switch", "requested_mode", path)

	var mode models.DataMode
	switch path {
	case "live":
		mode = models.DataModeLive
	case "test":
		mode = models.DataModeTest
	default:
		h.logger.Warn("Invalid mode requested", "mode", path)
		http.Error(w, "Invalid mode. Use 'live' or 'test'.", http.StatusBadRequest)
		return
	}

	h.dataProcessingUseCase.SetMode(mode)
	h.logger.Info("Mode switched successfully", "new_mode", mode)

	response := map[string]interface{}{
		"status": "success",
		"mode":   string(mode),
		"message": "Data mode switched successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
