package v1

import (
	"encoding/json"
	"net/http"

	"crypto/internal/core/port"
)

type HealthHandler struct {
	healthService port.HealthService
}

func NewHealthHandler(
	healthService port.HealthService,
) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// SimpleHealthResponse for basic health check
type SimpleHealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func (h *HealthHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	if h.healthService == nil {
		// Fallback basic health check if service is not available
		response := SimpleHealthResponse{
			Status:    "degraded",
			Timestamp: "2025-01-08T12:00:00Z",
			Message:   "Health service not available",
		}
		h.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	healthStatus, err := h.healthService.CheckSystemHealth(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to check system health: "+err.Error())
		return
	}

	// Determine HTTP status code based on health status
	statusCode := http.StatusOK
	switch healthStatus.Status {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusPartialContent // 206
	}

	h.writeJSONResponse(w, statusCode, healthStatus)
}

func (h *HealthHandler) GetDetailedHealth(w http.ResponseWriter, r *http.Request) {
	if h.healthService == nil {
		// Fallback basic health check if service is not available
		response := SimpleHealthResponse{
			Status:    "degraded",
			Timestamp: "2025-01-08T12:00:00Z",
			Message:   "Health service not available",
		}
		h.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	detailedHealth, err := h.healthService.GetDetailedHealth(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get detailed health: "+err.Error())
		return
	}

	// Determine HTTP status code based on health status
	statusCode := http.StatusOK
	switch detailedHealth.Status {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusPartialContent // 206
	}

	h.writeJSONResponse(w, statusCode, detailedHealth)
}

// Helper methods

func (h *HealthHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, log the error and send a simple error message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error","message":"failed to encode response"}`))
	}
}

func (h *HealthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
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
