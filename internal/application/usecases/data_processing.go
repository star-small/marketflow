package usecases

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"marketflow/internal/application/ports"
	"marketflow/internal/concurrency"
	"marketflow/internal/domain/models"
)

// DataProcessingUseCase handles data processing operations
type DataProcessingUseCase struct {
	storage             ports.StoragePort
	cache               ports.CachePort
	concurrencyManager  *concurrency.Manager
	logger              *slog.Logger
	mode                models.DataMode
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	liveExchange        ports.ExchangePort
	testExchange        ports.ExchangePort
	isRunning           bool
}

// NewDataProcessingUseCase creates a new DataProcessingUseCase
func NewDataProcessingUseCase(storage ports.StoragePort, cache ports.CachePort, concurrencyManager *concurrency.Manager, logger *slog.Logger) *DataProcessingUseCase {
	return &DataProcessingUseCase{
		storage:            storage,
		cache:              cache,
		concurrencyManager: concurrencyManager,
		logger:             logger,
		mode:               models.DataModeLive,
		isRunning:          false,
	}
}

// Start begins data processing
func (uc *DataProcessingUseCase) Start(ctx context.Context, liveExchange, testExchange ports.ExchangePort) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	uc.liveExchange = liveExchange
	uc.testExchange = testExchange
	uc.ctx, uc.cancel = context.WithCancel(ctx)

	// Start aggregation ticker
	go uc.startAggregationTicker(uc.ctx)

	// Start cleanup ticker
	go uc.startCleanupTicker(uc.ctx)

	// Start data processing based on current mode
	go uc.startDataProcessing(uc.ctx)

	uc.isRunning = true
	uc.logger.Info("Data processing use case started")
	return nil
}

// SetMode switches between live and test modes
func (uc *DataProcessingUseCase) SetMode(mode models.DataMode) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.mode != mode {
		oldMode := uc.mode
		uc.mode = mode
		uc.logger.Info("Data mode switching", "from", oldMode, "to", mode)

		// Restart data processing with new mode if system is running
		if uc.isRunning && uc.cancel != nil {
			uc.cancel() // Stop current processing

			// Start new processing with updated mode
			uc.ctx, uc.cancel = context.WithCancel(context.Background())
			go uc.startDataProcessing(uc.ctx)

			uc.logger.Info("Data processing restarted with new mode", "mode", mode)
		}
	}
}

// GetMode returns the current data mode
func (uc *DataProcessingUseCase) GetMode() models.DataMode {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.mode
}

func (uc *DataProcessingUseCase) startDataProcessing(ctx context.Context) {
	uc.logger.Info("Starting data processing pipeline")

	// Choose exchange based on current mode
	uc.mu.RLock()
	var exchange ports.ExchangePort
	if uc.mode == models.DataModeLive {
		exchange = uc.liveExchange
	} else {
		exchange = uc.testExchange
	}
	currentMode := uc.mode
	uc.mu.RUnlock()

	if exchange == nil {
		uc.logger.Error("Exchange not available for mode", "mode", currentMode)
		return
	}

	uc.logger.Info("Starting exchange data stream", "mode", currentMode, "exchange", exchange.GetName())

	// Start exchange data stream
	dataCh, err := exchange.Start(ctx)
	if err != nil {
		uc.logger.Error("Failed to start exchange", "error", err, "mode", currentMode)
		return
	}

	// Create channels for concurrency patterns
	processedCh := make(chan models.PriceUpdate, 1000)

	// Start worker pools (5 workers as specified in requirements)
	numWorkers := 5
	uc.concurrencyManager.StartWorkerPool(ctx, exchange.GetName(), numWorkers, dataCh, processedCh)

	// Start result processor
	go uc.processResults(ctx, processedCh)

	uc.logger.Info("Data processing pipeline started", "mode", currentMode, "exchange", exchange.GetName())
}

func (uc *DataProcessingUseCase) processResults(ctx context.Context, resultCh <-chan models.PriceUpdate) {
	uc.logger.Info("Starting result processor")

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("Result processor stopped")
			return
		case update, ok := <-resultCh:
			if !ok {
				uc.logger.Info("Result channel closed")
				return
			}

			// Process each price update
			if err := uc.processPriceUpdate(ctx, update); err != nil {
				uc.logger.Error("Failed to process price update", "error", err, "symbol", update.Symbol)
			}
		}
	}
}

func (uc *DataProcessingUseCase) processPriceUpdate(ctx context.Context, update models.PriceUpdate) error {
	// Cache the latest price in Redis
	if err := uc.cache.SetLatestPrice(ctx, update); err != nil {
		uc.logger.Error("Failed to cache price update", "error", err)
		// Don't return error - continue processing even if cache fails
	}

	uc.logger.Debug("Processed price update",
		"symbol", update.Symbol,
		"exchange", update.Exchange,
		"price", update.Price)

	return nil
}

func (uc *DataProcessingUseCase) startAggregationTicker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	uc.logger.Info("Starting aggregation ticker")

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("Aggregation ticker stopped")
			return
		case <-ticker.C:
			uc.aggregateData(ctx)
		}
	}
}

func (uc *DataProcessingUseCase) startCleanupTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	uc.logger.Info("Starting cleanup ticker")

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("Cleanup ticker stopped")
			return
		case <-ticker.C:
			if err := uc.cache.CleanupOldData(ctx, 2*time.Minute); err != nil {
				uc.logger.Error("Failed to cleanup old data", "error", err)
			}
		}
	}
}

func (uc *DataProcessingUseCase) aggregateData(ctx context.Context) {
	uc.logger.Info("Starting data aggregation")

	symbols := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}
	exchanges := []string{"exchange1", "exchange2", "exchange3", "test-exchange1", "test-exchange2", "test-exchange3"}

	var aggregatedData []models.AggregatedData

	for _, symbol := range symbols {
		for _, exchange := range exchanges {
			// Get price history for the last minute
			history, err := uc.cache.GetPriceHistory(ctx, symbol, exchange, time.Minute)
			if err != nil || len(history) == 0 {
				continue
			}

			// Calculate aggregations
			var total, min, max float64
			min = history[0].Price
			max = history[0].Price

			for _, price := range history {
				total += price.Price
				if price.Price < min {
					min = price.Price
				}
				if price.Price > max {
					max = price.Price
				}
			}

			avg := total / float64(len(history))

			aggregated := models.AggregatedData{
				PairName:     symbol,
				Exchange:     exchange,
				Timestamp:    time.Now(),
				AveragePrice: avg,
				MinPrice:     min,
				MaxPrice:     max,
			}

			aggregatedData = append(aggregatedData, aggregated)
		}
	}

	// Store aggregated data in PostgreSQL
	if len(aggregatedData) > 0 {
		if err := uc.storage.SaveAggregatedData(ctx, aggregatedData); err != nil {
			uc.logger.Error("Failed to save aggregated data", "error", err)
		} else {
			uc.logger.Info("Saved aggregated data", "count", len(aggregatedData))
		}
	}
}
