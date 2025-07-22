package ports

import (
	"context"
	"time"

	"marketflow/internal/domain/models"
)

// CachePort defines the interface for caching operations
type CachePort interface {
	// SetLatestPrice sets the latest price for a symbol from an exchange
	SetLatestPrice(ctx context.Context, update models.PriceUpdate) error

	// GetLatestPrice gets the latest price for a symbol from an exchange
	GetLatestPrice(ctx context.Context, symbol, exchange string) (*models.LatestPrice, error)

	// GetLatestPrices gets latest prices for a symbol from all exchanges
	GetLatestPrices(ctx context.Context, symbol string) ([]*models.LatestPrice, error)

	// GetPriceHistory gets price history for aggregation (last minute)
	GetPriceHistory(ctx context.Context, symbol, exchange string, duration time.Duration) ([]models.PriceUpdate, error)

	// CleanupOldData removes old price data from cache
	CleanupOldData(ctx context.Context, maxAge time.Duration) error

	// Close closes the cache connection
	Close() error
}
