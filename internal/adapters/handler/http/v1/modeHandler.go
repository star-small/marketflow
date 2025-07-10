package v1

import (
	"encoding/json"
	"net/http"

	"crypto/internal/core/port"
)

type ModeHandler struct {
	modeService port.ModeService
}

func NewModeHandler(
	modeService port.ModeService,
) *ModeHandler {
	return &ModeHandler{
		modeService: modeService,
	}
}

// Response structures
type ModeResponse struct {
	Mode    string                 `json:"mode"`
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Stats   map[string]interface{} `json:"stats,omitempty"`
}

func (h *ModeHandler) SwitchToTestMode(w http.ResponseWriter, r *http.Request) {
	if h.modeService == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "mode service not available")
		return
	}

	err := h.modeService.SwitchToTestMode(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to switch to test mode: "+err.Error())
		return
	}

	response := ModeResponse{
		Mode:    "test",
		Success: true,
		Message: "Successfully switched to test mode",
		Stats:   h.modeService.GetStats(),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ModeHandler) SwitchToLiveMode(w http.ResponseWriter, r *http.Request) {
	if h.modeService == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "mode service not available")
		return
	}

	err := h.modeService.SwitchToLiveMode(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to switch to live mode: "+err.Error())
		return
	}

	response := ModeResponse{
		Mode:    "live",
		Success: true,
		Message: "Successfully switched to live mode",
		Stats:   h.modeService.GetStats(),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ModeHandler) GetCurrentMode(w http.ResponseWriter, r *http.Request) {
	if h.modeService == nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "mode service not available")
		return
	}

	currentMode := h.modeService.GetCurrentMode()
	isRunning := h.modeService.IsRunning()

	response := ModeResponse{
		Mode:    currentMode,
		Success: true,
		Message: func() string {
			if isRunning {
				return "System is running in " + currentMode + " mode"
			}
			return "System is configured for " + currentMode + " mode but not running"
		}(),
		Stats: h.modeService.GetStats(),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (h *ModeHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, log the error and send a simple error message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error","message":"failed to encode response"}`))
	}
}

func (h *ModeHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorType := "bad_request"
	switch statusCode {
	case http.StatusNotFound:
		errorType = "not_found"
	case http.StatusInternalServerError:
		errorType = "internal_error"
	case http.StatusServiceUnavailable:
		errorType = "service_unavailable"
	}

	response := ErrorResponse{
		Error:   errorType,
		Message: message,
	}

	h.writeJSONResponse(w, statusCode, response)
}
