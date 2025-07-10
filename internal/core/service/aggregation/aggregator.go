package aggregation

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

// PriceAggregator handles minute-by-minute price aggregation
type PriceAggregator struct {
	priceService port.PriceService

	// Aggregation state
	aggregations map[string]*AggregationData
	mutex        sync.RWMutex

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// AggregationData holds the current aggregation state for a symbol-exchange pair
type AggregationData struct {
	Symbol     string
	Exchange   string
	Prices     []float64
	MinPrice   float64
	MaxPrice   float64
	Sum        float64
	Count      int
	StartTime  time.Time
	LastUpdate time.Time
}

func NewPriceAggregator(priceService port.PriceService) *PriceAggregator {
	ctx, cancel := context.WithCancel(context.Background())

	return &PriceAggregator{
		priceService: priceService,
		aggregations: make(map[string]*AggregationData),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the aggregation process
func (pa *PriceAggregator) Start() {
	slog.Info("Starting price aggregator...")

	// Start the minute ticker for aggregation
	pa.wg.Add(1)
	go pa.runAggregationTimer()

	slog.Info("Price aggregator started")
}

// Stop gracefully stops the aggregation process
func (pa *PriceAggregator) Stop() {
	slog.Info("Stopping price aggregator...")

	pa.cancel()
	pa.wg.Wait()

	slog.Info("Price aggregator stopped")
}

// ProcessPrice adds a new price data point to the aggregation
func (pa *PriceAggregator) ProcessPrice(data domain.MarketData) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	key := pa.getAggregationKey(data.Symbol, data.Exchange)
	now := time.Now()

	// Get or create aggregation data
	agg, exists := pa.aggregations[key]
	if !exists {
		agg = &AggregationData{
			Symbol:     data.Symbol,
			Exchange:   data.Exchange,
			MinPrice:   data.Price,
			MaxPrice:   data.Price,
			StartTime:  pa.getMinuteStart(now),
			LastUpdate: now,
		}
		pa.aggregations[key] = agg
	}

	// Check if we need to start a new aggregation period
	currentMinute := pa.getMinuteStart(now)
	if !currentMinute.Equal(agg.StartTime) {
		// Store the previous aggregation if it has data
		if agg.Count > 0 {
			pa.storeAggregation(agg)
		}

		// Start new aggregation period
		agg = &AggregationData{
			Symbol:     data.Symbol,
			Exchange:   data.Exchange,
			MinPrice:   data.Price,
			MaxPrice:   data.Price,
			StartTime:  currentMinute,
			LastUpdate: now,
		}
		pa.aggregations[key] = agg
	}

	// Update aggregation with new price
	agg.Prices = append(agg.Prices, data.Price)
	agg.Sum += data.Price
	agg.Count++
	agg.LastUpdate = now

	// Update min/max
	if data.Price < agg.MinPrice {
		agg.MinPrice = data.Price
	}
	if data.Price > agg.MaxPrice {
		agg.MaxPrice = data.Price
	}
}

// runAggregationTimer runs the periodic aggregation timer
func (pa *PriceAggregator) runAggregationTimer() {
	defer pa.wg.Done()

	// Align to the next minute boundary
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	initialDelay := nextMinute.Sub(now)

	slog.Info("Aggregation timer starting", "initial_delay", initialDelay, "next_minute", nextMinute)

	// Wait for the next minute boundary
	select {
	case <-time.After(initialDelay):
	case <-pa.ctx.Done():
		return
	}

	// Start the minute ticker
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pa.processAggregations()
		case <-pa.ctx.Done():
			// Final aggregation before stopping
			pa.processAggregations()
			return
		}
	}
}

// processAggregations processes all current aggregations and stores completed ones
func (pa *PriceAggregator) processAggregations() {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	now := time.Now()
	currentMinute := pa.getMinuteStart(now)

	slog.Info("Processing aggregations", "current_minute", currentMinute, "total_aggregations", len(pa.aggregations))

	// Process all aggregations
	for key, agg := range pa.aggregations {
		// If aggregation is for a previous minute and has data, store it
		if agg.StartTime.Before(currentMinute) && agg.Count > 0 {
			pa.storeAggregation(agg)
			delete(pa.aggregations, key)
		}
	}
}

// storeAggregation stores the completed aggregation
func (pa *PriceAggregator) storeAggregation(agg *AggregationData) {
	if agg.Count == 0 {
		return
	}

	avgPrice := agg.Sum / float64(agg.Count)

	aggregation := domain.PriceAggregation{
		PairName:     agg.Symbol,
		Exchange:     agg.Exchange,
		Timestamp:    agg.StartTime,
		AveragePrice: avgPrice,
		MinPrice:     agg.MinPrice,
		MaxPrice:     agg.MaxPrice,
		DataPoints:   agg.Count,
	}

	// Store the aggregation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pa.storeAggregationData(ctx, aggregation); err != nil {
		slog.Error("Failed to store aggregation",
			"symbol", agg.Symbol,
			"exchange", agg.Exchange,
			"timestamp", agg.StartTime,
			"error", err,
		)
	} else {
		slog.Info("Stored aggregation",
			"symbol", agg.Symbol,
			"exchange", agg.Exchange,
			"timestamp", agg.StartTime,
			"avg_price", avgPrice,
			"min_price", agg.MinPrice,
			"max_price", agg.MaxPrice,
			"data_points", agg.Count,
		)
	}
}

// storeAggregationData stores aggregation via the price service
func (pa *PriceAggregator) storeAggregationData(ctx context.Context, agg domain.PriceAggregation) error {
	// Use type assertion to access the storage method
	if service, ok := pa.priceService.(interface {
		StorePriceAggregation(ctx context.Context, agg domain.PriceAggregation) error
	}); ok {
		return service.StorePriceAggregation(ctx, agg)
	}

	slog.Warn("Price service does not support aggregation storage")
	return nil
}

// Helper methods

func (pa *PriceAggregator) getAggregationKey(symbol, exchange string) string {
	return symbol + ":" + exchange
}

func (pa *PriceAggregator) getMinuteStart(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

// GetStats returns statistics about the aggregator
func (pa *PriceAggregator) GetStats() map[string]interface{} {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	stats := map[string]interface{}{
		"active_aggregations": len(pa.aggregations),
		"timestamp":           time.Now(),
	}

	// Add details about each aggregation
	aggregations := make(map[string]interface{})
	for key, agg := range pa.aggregations {
		aggregations[key] = map[string]interface{}{
			"symbol":      agg.Symbol,
			"exchange":    agg.Exchange,
			"count":       agg.Count,
			"start_time":  agg.StartTime,
			"last_update": agg.LastUpdate,
			"min_price":   agg.MinPrice,
			"max_price":   agg.MaxPrice,
			"avg_price": func() float64 {
				if agg.Count > 0 {
					return agg.Sum / float64(agg.Count)
				}
				return 0
			}(),
		}
	}
	stats["aggregations"] = aggregations

	return stats
}

// roundPrice rounds price to appropriate decimal places
func (pa *PriceAggregator) roundPrice(price float64) float64 {
	if price > 1000 {
		// For high-value coins like BTC, round to 2 decimal places
		return math.Round(price*100) / 100
	} else if price > 10 {
		// For medium-value coins, round to 3 decimal places
		return math.Round(price*1000) / 1000
	} else {
		// For low-value coins, round to 4 decimal places
		return math.Round(price*10000) / 10000
	}
}
