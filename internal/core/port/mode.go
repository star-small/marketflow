package port

import "context"

type ModeRepository interface{}

type ModeService interface {
	// Switch to test mode (use synthetic data)
	SwitchToTestMode(ctx context.Context) error

	// Switch to live mode (connect to real exchanges)
	SwitchToLiveMode(ctx context.Context) error

	// Get current mode
	GetCurrentMode() string

	// Check if service is running
	IsRunning() bool

	// Get service statistics
	GetStats() map[string]interface{}
}
