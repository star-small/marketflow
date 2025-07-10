package port

import (
	"context"
	"time"
)

type HealthRepository interface{}

type HealthService interface {
	// Check overall system health
	CheckSystemHealth(ctx context.Context) (*HealthStatus, error)

	// Get detailed health information
	GetDetailedHealth(ctx context.Context) (*DetailedHealthStatus, error)
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	Status  string                 `json:"status"`
	Error   string                 `json:"error,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type DetailedHealthStatus struct {
	HealthStatus
	SystemInfo map[string]interface{} `json:"system_info"`
}
