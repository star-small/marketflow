package ports

import (
	"context"

	"marketflow/internal/domain/models"
)

// ExchangePort defines the interface for exchange data sources
type ExchangePort interface {
	// Start begins data collection
	Start(ctx context.Context) (<-chan models.PriceUpdate, error)

	// Stop stops data collection
	Stop() error

	// IsConnected returns connection status
	IsConnected() bool

	// GetName returns the exchange name
	GetName() string
}
