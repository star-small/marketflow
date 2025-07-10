// Replace the content of internal/core/port/prices.go

package port

import (
	"context"
	"time"

	"crypto/internal/core/domain"
)

type PriceRepository interface {
	// Store aggregated price data
	StorePriceAggregation(ctx context.Context, agg domain.PriceAggregation) error

	// Get aggregated prices in a time range
	GetAggregatedPricesInRange(ctx context.Context, symbol, exchange string, from, to time.Time) ([]domain.PriceAggregation, error)

	// Get highest price in a time range
	GetHighestPriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceAggregation, error)

	// Get lowest price in a time range
	GetLowestPriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceAggregation, error)

	// Get average price in a time range
	GetAveragePriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (float64, error)

	// Cleanup old data
	CleanupOldData(ctx context.Context, olderThan time.Duration) error
}

type PriceService interface {
	// Get the latest price for a symbol across all exchanges
	GetLatestPrice(ctx context.Context, symbol string) (*domain.MarketData, error)

	// Get the latest price for a symbol from a specific exchange
	GetLatestPriceByExchange(ctx context.Context, symbol, exchange string) (*domain.MarketData, error)

	// Get highest price in a time period
	GetHighestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error)
	GetHighestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error)

	// Get lowest price in a time period
	GetLowestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error)
	GetLowestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error)

	// Get average price in a time period
	GetAveragePrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error)
	GetAveragePriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error)

	// Store aggregated price data (used by aggregation service)
	StorePriceAggregation(ctx context.Context, agg domain.PriceAggregation) error
}
