// Replace internal/core/service/health/service.go

package health

import (
	"context"
	"database/sql"
	"time"

	"crypto/internal/core/port"
)

type HealthService struct {
	db              *sql.DB
	cache           port.Cache
	exchangeService port.ExchangeService
}

func NewHealthService(db *sql.DB, cache port.Cache, exchangeService port.ExchangeService) port.HealthService {
	return &HealthService{
		db:              db,
		cache:           cache,
		exchangeService: exchangeService,
	}
}

func (s *HealthService) CheckSystemHealth(ctx context.Context) (*port.HealthStatus, error) {
	status := &port.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks:    make(map[string]port.CheckResult),
	}

	// Check database
	if s.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := s.db.PingContext(dbCtx); err != nil {
			status.Checks["database"] = port.CheckResult{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			status.Status = "unhealthy"
		} else {
			status.Checks["database"] = port.CheckResult{
				Status: "healthy",
			}
		}
		cancel()
	} else {
		status.Checks["database"] = port.CheckResult{
			Status: "unavailable",
			Error:  "database not configured",
		}
	}

	// Check cache
	if s.cache != nil {
		cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := s.cache.Ping(cacheCtx); err != nil {
			status.Checks["cache"] = port.CheckResult{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			if status.Status == "healthy" {
				status.Status = "degraded"
			}
		} else {
			status.Checks["cache"] = port.CheckResult{
				Status: "healthy",
			}
		}
		cancel()
	} else {
		status.Checks["cache"] = port.CheckResult{
			Status: "unavailable",
			Error:  "cache not configured",
		}
	}

	// Check exchange service
	if s.exchangeService != nil {
		if exchSvc, ok := s.exchangeService.(interface{ IsRunning() bool }); ok {
			if exchSvc.IsRunning() {
				status.Checks["exchange_service"] = port.CheckResult{
					Status: "healthy",
				}
			} else {
				status.Checks["exchange_service"] = port.CheckResult{
					Status: "stopped",
					Error:  "exchange service not running",
				}
				if status.Status == "healthy" {
					status.Status = "degraded"
				}
			}
		}

		// Check individual exchanges
		if exchSvc, ok := s.exchangeService.(interface{ GetActiveAdapters() []port.ExchangeAdapter }); ok {
			adapters := exchSvc.GetActiveAdapters()
			healthyCount := 0
			for _, adapter := range adapters {
				if adapter.IsHealthy() {
					healthyCount++
				}
			}

			exchangeCheck := port.CheckResult{
				Status: "healthy",
				Details: map[string]interface{}{
					"total_exchanges":   len(adapters),
					"healthy_exchanges": healthyCount,
				},
			}

			if healthyCount == 0 && len(adapters) > 0 {
				exchangeCheck.Status = "unhealthy"
				status.Status = "unhealthy"
			} else if healthyCount < len(adapters) {
				exchangeCheck.Status = "degraded"
				if status.Status == "healthy" {
					status.Status = "degraded"
				}
			}

			status.Checks["exchanges"] = exchangeCheck
		}
	}

	return status, nil
}

func (s *HealthService) GetDetailedHealth(ctx context.Context) (*port.DetailedHealthStatus, error) {
	baseHealth, err := s.CheckSystemHealth(ctx)
	if err != nil {
		return nil, err
	}

	detailed := &port.DetailedHealthStatus{
		HealthStatus: *baseHealth,
		SystemInfo:   make(map[string]interface{}),
	}

	// Add system information
	detailed.SystemInfo["version"] = "1.0.0"
	detailed.SystemInfo["uptime"] = time.Since(startTime).String()

	// Add exchange service stats if available
	if s.exchangeService != nil {
		if exchSvc, ok := s.exchangeService.(interface{ GetStats() map[string]interface{} }); ok {
			detailed.SystemInfo["exchange_stats"] = exchSvc.GetStats()
		}
	}

	return detailed, nil
}

var startTime = time.Now()
