// Replace internal/core/service/mode/service.go

package mode

import (
	"context"
	"fmt"
	"log/slog"

	"crypto/internal/core/port"
)

type ModeService struct {
	exchangeService port.ExchangeService
}

func NewModeService(exchangeService port.ExchangeService) port.ModeService {
	return &ModeService{
		exchangeService: exchangeService,
	}
}

func (s *ModeService) SwitchToTestMode(ctx context.Context) error {
	slog.Info("Switching to test mode...")

	if err := s.exchangeService.SwitchToTestMode(ctx); err != nil {
		slog.Error("Failed to switch to test mode", "error", err)
		return fmt.Errorf("failed to switch to test mode: %w", err)
	}

	slog.Info("Successfully switched to test mode")
	return nil
}

func (s *ModeService) SwitchToLiveMode(ctx context.Context) error {
	slog.Info("Switching to live mode...")

	if err := s.exchangeService.SwitchToLiveMode(ctx); err != nil {
		slog.Error("Failed to switch to live mode", "error", err)
		return fmt.Errorf("failed to switch to live mode: %w", err)
	}

	slog.Info("Successfully switched to live mode")
	return nil
}

func (s *ModeService) GetCurrentMode() string {
	return s.exchangeService.GetCurrentMode()
}

func (s *ModeService) IsRunning() bool {
	if exchSvc, ok := s.exchangeService.(interface{ IsRunning() bool }); ok {
		return exchSvc.IsRunning()
	}
	return false
}

func (s *ModeService) GetStats() map[string]interface{} {
	if exchSvc, ok := s.exchangeService.(interface{ GetStats() map[string]interface{} }); ok {
		return exchSvc.GetStats()
	}
	return map[string]interface{}{
		"current_mode": s.GetCurrentMode(),
		"is_running":   s.IsRunning(),
	}
}
