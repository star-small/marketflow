package prices

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

type PriceService struct {
	cache      port.Cache
	repository port.PriceRepository
}

// NewPriceService creates a new price service with proper interface dependencies
func NewPriceService(cache port.Cache, repository port.PriceRepository) port.PriceService {
	return &PriceService{
		cache:      cache,
		repository: repository,
	}
}

func (s *PriceService) GetLatestPrice(ctx context.Context, symbol string) (*domain.MarketData, error) {
	// If cache is available, try to get from cache first
	if s.cache != nil {
		data, err := s.cache.GetLatestPrice(ctx, symbol)
		if err == nil && data != nil {
			return data, nil
		}
		// If cache fails or no data, continue to fallback
	}

	// TODO: Fallback to PostgreSQL if cache is unavailable
	// For now, return error if cache is not available
	if s.cache == nil {
		return nil, fmt.Errorf("no cache available and PostgreSQL fallback not implemented")
	}

	return nil, fmt.Errorf("no price data found for symbol %s", symbol)
}

func (s *PriceService) GetLatestPriceByExchange(ctx context.Context, symbol, exchange string) (*domain.MarketData, error) {
	// If cache is available, try to get from cache first
	if s.cache != nil {
		data, err := s.cache.GetLatestPriceByExchange(ctx, symbol, exchange)
		if err == nil && data != nil {
			return data, nil
		}
		// If cache fails or no data, continue to fallback
	}

	// TODO: Fallback to PostgreSQL if cache is unavailable
	// For now, return error if cache is not available
	if s.cache == nil {
		return nil, fmt.Errorf("no cache available and PostgreSQL fallback not implemented")
	}

	return nil, fmt.Errorf("no price data found for symbol %s on exchange %s", symbol, exchange)
}

func (s *PriceService) GetHighestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getHighestPriceInternal(ctx, symbol, "", period)
}

func (s *PriceService) GetHighestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getHighestPriceInternal(ctx, symbol, exchange, period)
}

func (s *PriceService) getHighestPriceInternal(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	// Try cache first for recent data (if period is small)
	if period <= 5*time.Minute && s.cache != nil {
		now := time.Now()
		from := now.Add(-period)

		var prices []domain.MarketData
		var err error

		if exchange != "" {
			prices, err = s.cache.GetPricesInRangeByExchange(ctx, symbol, exchange, from, now)
		} else {
			prices, err = s.cache.GetPricesInRange(ctx, symbol, from, now)
		}

		if err == nil && len(prices) > 0 {
			return s.calculateHighestFromPrices(prices, symbol, exchange, period), nil
		}
	}

	// Fallback to repository for longer periods or if cache fails
	if s.repository != nil {
		now := time.Now()
		from := now.Add(-period)

		agg, err := s.repository.GetHighestPriceInRange(ctx, symbol, exchange, from, now)
		if err != nil {
			return nil, fmt.Errorf("failed to get highest price from repository: %w", err)
		}

		if agg == nil {
			return nil, fmt.Errorf("no price data found for symbol %s", symbol)
		}

		return &domain.PriceStatistics{
			Symbol:     symbol,
			Exchange:   exchange,
			Price:      agg.MaxPrice,
			Timestamp:  agg.Timestamp,
			Period:     period.String(),
			DataPoints: 1, // Repository stores aggregated data
		}, nil
	}

	return nil, fmt.Errorf("no repository available")
}

func (s *PriceService) GetLowestPrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getLowestPriceInternal(ctx, symbol, "", period)
}

func (s *PriceService) GetLowestPriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getLowestPriceInternal(ctx, symbol, exchange, period)
}

func (s *PriceService) getLowestPriceInternal(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	// Try cache first for recent data (if period is small)
	if period <= 5*time.Minute && s.cache != nil {
		now := time.Now()
		from := now.Add(-period)

		var prices []domain.MarketData
		var err error

		if exchange != "" {
			prices, err = s.cache.GetPricesInRangeByExchange(ctx, symbol, exchange, from, now)
		} else {
			prices, err = s.cache.GetPricesInRange(ctx, symbol, from, now)
		}

		if err == nil && len(prices) > 0 {
			return s.calculateLowestFromPrices(prices, symbol, exchange, period), nil
		}
	}

	// Fallback to repository for longer periods or if cache fails
	if s.repository != nil {
		now := time.Now()
		from := now.Add(-period)

		agg, err := s.repository.GetLowestPriceInRange(ctx, symbol, exchange, from, now)
		if err != nil {
			return nil, fmt.Errorf("failed to get lowest price from repository: %w", err)
		}

		if agg == nil {
			return nil, fmt.Errorf("no price data found for symbol %s", symbol)
		}

		return &domain.PriceStatistics{
			Symbol:     symbol,
			Exchange:   exchange,
			Price:      agg.MinPrice,
			Timestamp:  agg.Timestamp,
			Period:     period.String(),
			DataPoints: 1, // Repository stores aggregated data
		}, nil
	}

	return nil, fmt.Errorf("no repository available")
}

func (s *PriceService) GetAveragePrice(ctx context.Context, symbol string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getAveragePriceInternal(ctx, symbol, "", period)
}

func (s *PriceService) GetAveragePriceByExchange(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	return s.getAveragePriceInternal(ctx, symbol, exchange, period)
}

func (s *PriceService) getAveragePriceInternal(ctx context.Context, symbol, exchange string, period time.Duration) (*domain.PriceStatistics, error) {
	// Try cache first for recent data (if period is small)
	if period <= 5*time.Minute && s.cache != nil {
		now := time.Now()
		from := now.Add(-period)

		var prices []domain.MarketData
		var err error

		if exchange != "" {
			prices, err = s.cache.GetPricesInRangeByExchange(ctx, symbol, exchange, from, now)
		} else {
			prices, err = s.cache.GetPricesInRange(ctx, symbol, from, now)
		}

		if err == nil && len(prices) > 0 {
			return s.calculateAverageFromPrices(prices, symbol, exchange, period), nil
		}
	}

	// Fallback to repository for longer periods or if cache fails
	if s.repository != nil {
		now := time.Now()
		from := now.Add(-period)

		avgPrice, err := s.repository.GetAveragePriceInRange(ctx, symbol, exchange, from, now)
		if err != nil {
			return nil, fmt.Errorf("failed to get average price from repository: %w", err)
		}

		return &domain.PriceStatistics{
			Symbol:     symbol,
			Exchange:   exchange,
			Price:      avgPrice,
			Timestamp:  time.Now(),
			Period:     period.String(),
			DataPoints: 1, // Repository stores aggregated data
		}, nil
	}

	return nil, fmt.Errorf("no repository available")
}

// Helper functions for calculating statistics from raw price data

func (s *PriceService) calculateHighestFromPrices(prices []domain.MarketData, symbol, exchange string, period time.Duration) *domain.PriceStatistics {
	if len(prices) == 0 {
		return nil
	}

	highest := prices[0]
	for _, price := range prices[1:] {
		if price.Price > highest.Price {
			highest = price
		}
	}

	return &domain.PriceStatistics{
		Symbol:     symbol,
		Exchange:   exchange,
		Price:      highest.Price,
		Timestamp:  time.Unix(highest.Timestamp, 0),
		Period:     period.String(),
		DataPoints: len(prices),
	}
}

func (s *PriceService) calculateLowestFromPrices(prices []domain.MarketData, symbol, exchange string, period time.Duration) *domain.PriceStatistics {
	if len(prices) == 0 {
		return nil
	}

	lowest := prices[0]
	for _, price := range prices[1:] {
		if price.Price < lowest.Price {
			lowest = price
		}
	}

	return &domain.PriceStatistics{
		Symbol:     symbol,
		Exchange:   exchange,
		Price:      lowest.Price,
		Timestamp:  time.Unix(lowest.Timestamp, 0),
		Period:     period.String(),
		DataPoints: len(prices),
	}
}

func (s *PriceService) calculateAverageFromPrices(prices []domain.MarketData, symbol, exchange string, period time.Duration) *domain.PriceStatistics {
	if len(prices) == 0 {
		return nil
	}

	var sum float64
	var latestTimestamp int64

	for _, price := range prices {
		sum += price.Price
		if price.Timestamp > latestTimestamp {
			latestTimestamp = price.Timestamp
		}
	}

	average := sum / float64(len(prices))

	return &domain.PriceStatistics{
		Symbol:     symbol,
		Exchange:   exchange,
		Price:      average,
		Timestamp:  time.Unix(latestTimestamp, 0),
		Period:     period.String(),
		DataPoints: len(prices),
	}
}

// StorePriceAggregation stores aggregated price data (used by the aggregation system)
func (s *PriceService) StorePriceAggregation(ctx context.Context, agg domain.PriceAggregation) error {
	if s.repository == nil {
		return fmt.Errorf("no repository available")
	}

	if err := s.repository.StorePriceAggregation(ctx, agg); err != nil {
		slog.Error("Failed to store price aggregation", "error", err, "pair", agg.PairName, "exchange", agg.Exchange)
		return fmt.Errorf("failed to store price aggregation: %w", err)
	}

	return nil
}
