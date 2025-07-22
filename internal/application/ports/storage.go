package ports

import (
	"context"
	"time"

	"marketflow/internal/domain/models"
)

// StoragePort defines the interface for data storage operations
type StoragePort interface {
	// SaveAggregatedData saves aggregated market data
	SaveAggregatedData(ctx context.Context, data []models.AggregatedData) error

	// GetAggregatedData retrieves aggregated data within a time range
	GetAggregatedData(ctx context.Context, symbol, exchange string, from, to time.Time) ([]models.AggregatedData, error)

	// GetHighestPrice returns the highest price within a period
	GetHighestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error)

	// GetLowestPrice returns the lowest price within a period
	GetLowestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error)

	// GetAveragePrice returns the average price within a period
	GetAveragePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error)

	// Close closes the storage connection
	Close() error
}
