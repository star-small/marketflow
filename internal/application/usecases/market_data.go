package usecases

import (
	"context"
	"log/slog"
	"time"

	"marketflow/internal/application/ports"
	"marketflow/internal/domain/models"
)

// MarketDataUseCase handles market data operations
type MarketDataUseCase struct {
	storage ports.StoragePort
	cache   ports.CachePort
	logger  *slog.Logger
}

// NewMarketDataUseCase creates a new MarketDataUseCase
func NewMarketDataUseCase(storage ports.StoragePort, cache ports.CachePort, logger *slog.Logger) *MarketDataUseCase {
	return &MarketDataUseCase{
		storage: storage,
		cache:   cache,
		logger:  logger,
	}
}

// GetLatestPrice returns the latest price for a symbol
func (uc *MarketDataUseCase) GetLatestPrice(ctx context.Context, symbol, exchange string) (*models.LatestPrice, error) {
	if exchange != "" {
		return uc.cache.GetLatestPrice(ctx, symbol, exchange)
	}

	prices, err := uc.cache.GetLatestPrices(ctx, symbol)
	if err != nil {
		return nil, err
	}

	if len(prices) == 0 {
		return nil, nil
	}

	// Return the most recent price
	latest := prices[0]
	for _, price := range prices[1:] {
		if price.Timestamp.After(latest.Timestamp) {
			latest = price
		}
	}

	return latest, nil
}

// GetHighestPrice returns the highest price within a period
func (uc *MarketDataUseCase) GetHighestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	return uc.storage.GetHighestPrice(ctx, symbol, exchange, period)
}

// GetLowestPrice returns the lowest price within a period
func (uc *MarketDataUseCase) GetLowestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	return uc.storage.GetLowestPrice(ctx, symbol, exchange, period)
}

// GetAveragePrice returns the average price within a period
func (uc *MarketDataUseCase) GetAveragePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	return uc.storage.GetAveragePrice(ctx, symbol, exchange, period)
}
