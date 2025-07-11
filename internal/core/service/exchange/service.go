package exchange

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"crypto/internal/adapters/exchanges"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

// ExchangeService implements the port.ExchangeService interface
// and manages concurrency patterns for market data processing
type ExchangeService struct {
	// Mode management
	currentMode string
	modeMutex   sync.RWMutex

	// Exchange adapters
	liveAdapters   []port.ExchangeAdapter
	testAdapters   []port.ExchangeAdapter
	activeAdapters []port.ExchangeAdapter

	// Concurrency channels
	aggregatedDataChan chan domain.MarketData   // Fan-in result
	workerPool         []chan domain.MarketData // Fan-out to workers
	resultChan         chan domain.MarketData   // Final processed data

	// Control
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	runMutex  sync.RWMutex
	wg        sync.WaitGroup

	// Configuration
	numWorkers int
}

// NewExchangeService creates a new exchange service with default configuration
// internal/core/service/exchange/service.go

// internal/core/service/exchange/service.go
// Replace the NewExchangeService function with this version that increases buffer sizes

// internal/core/service/exchange/service.go
// Replace the NewExchangeService function with this version that increases buffer sizes

// internal/core/service/exchange/service.go
// Replace the NewExchangeService function with this version that increases buffer sizes

func NewExchangeService() port.ExchangeService {
	ctx, cancel := context.WithCancel(context.Background())

	// Create live adapters for exchanges on ports 40101, 40102, 40103
	liveAdapters := exchanges.CreateLiveExchangeAdapters()

	// Create test adapters (generators)
	testAdapters := exchanges.CreateTestExchangeAdapters()

	numWorkers := 15 // 5 workers per exchange as specified in requirements

	return &ExchangeService{
		currentMode:    "test", // Default to test mode
		liveAdapters:   liveAdapters,
		testAdapters:   testAdapters,
		activeAdapters: testAdapters, // Start with test adapters
		// INCREASED BUFFER SIZES to handle pipeline bottleneck
		aggregatedDataChan: make(chan domain.MarketData, 5000), // Increased from 1000
		workerPool:         make([]chan domain.MarketData, numWorkers),
		resultChan:         make(chan domain.MarketData, 10000), // Increased from 1000
		ctx:                ctx,
		cancel:             cancel,
		numWorkers:         numWorkers,
	}
}

// Also replace the StopDataProcessing method to recreate larger channels
func (e *ExchangeService) StopDataProcessing() error {
	e.runMutex.Lock()
	defer e.runMutex.Unlock()

	if !e.isRunning {
		return nil // Already stopped
	}

	slog.Info("Stopping data processing...")

	// Stop all adapters
	if err := e.stopCurrentAdapters(); err != nil {
		slog.Error("Failed to stop adapters", "error", err)
	}

	// Cancel context to stop all goroutines
	if e.cancel != nil {
		e.cancel()
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		slog.Info("All goroutines stopped")
	case <-time.After(10 * time.Second):
		slog.Warn("Timeout waiting for goroutines to stop")
	}

	// Close worker channels
	for _, workerChan := range e.workerPool {
		if workerChan != nil {
			close(workerChan)
		}
	}

	// Close result channel
	if e.resultChan != nil {
		close(e.resultChan)
	}

	// Recreate channels with LARGER buffers for next start
	e.aggregatedDataChan = make(chan domain.MarketData, 5000) // Increased
	e.resultChan = make(chan domain.MarketData, 10000)        // Increased
	for i := range e.workerPool {
		e.workerPool[i] = make(chan domain.MarketData, 100)
	}

	e.isRunning = false
	slog.Info("Data processing stopped")
	return nil
}

// NewExchangeServiceWithAdapters creates a service with custom adapters
func NewExchangeServiceWithAdapters(liveAdapters, testAdapters []port.ExchangeAdapter) port.ExchangeService {
	ctx, cancel := context.WithCancel(context.Background())

	numWorkers := 15 // 5 workers per exchange

	return &ExchangeService{
		currentMode:        "test",
		liveAdapters:       liveAdapters,
		testAdapters:       testAdapters,
		activeAdapters:     testAdapters,
		aggregatedDataChan: make(chan domain.MarketData, 1000),
		workerPool:         make([]chan domain.MarketData, numWorkers),
		resultChan:         make(chan domain.MarketData, 1000),
		ctx:                ctx,
		cancel:             cancel,
		numWorkers:         numWorkers,
	}
}

// internal/core/service/exchange/service.go
// Replace the SwitchToTestMode method with this complete fix

func (e *ExchangeService) SwitchToTestMode(ctx context.Context) error {
	slog.Info("=== SwitchToTestMode CALLED ===")
	e.modeMutex.Lock()
	defer e.modeMutex.Unlock()

	slog.Info("SwitchToTestMode called", "currentMode", e.currentMode)

	if e.currentMode == "test" {
		slog.Info("Already in test mode, no switch needed")
		return nil
	}

	slog.Info("Switching to test mode...")

	// Check if data processing was running BEFORE stopping
	e.runMutex.RLock()
	wasRunning := e.isRunning
	e.runMutex.RUnlock()

	slog.Info("Checking initial state", "wasRunning", wasRunning)

	// STEP 1: STOP data processing completely if it was running
	if wasRunning {
		slog.Info("Stopping data processing before mode switch...")
		if err := e.StopDataProcessing(); err != nil {
			slog.Error("Failed to stop data processing", "error", err)
			return fmt.Errorf("failed to stop data processing: %w", err)
		}
		slog.Info("Data processing stopped successfully")
	}

	// STEP 2: Create FRESH test adapters (this ensures clean state)
	slog.Info("Creating fresh test adapters...")
	e.testAdapters = exchanges.CreateTestExchangeAdapters()

	// STEP 3: Switch to fresh test adapters
	e.activeAdapters = e.testAdapters
	e.currentMode = "test"
	slog.Info("Switched active adapters to fresh test adapters", "count", len(e.activeAdapters))

	// STEP 4: START data processing with new adapters if it was running before
	if wasRunning {
		slog.Info("Starting data processing with fresh test adapters...")
		bgCtx := context.Background()
		if err := e.StartDataProcessing(bgCtx); err != nil {
			slog.Error("Failed to start data processing with test adapters", "error", err)
			return fmt.Errorf("failed to start data processing: %w", err)
		}
		slog.Info("Data processing started successfully with test adapters")
	}

	slog.Info("Switched to test mode successfully")
	return nil
}

// Also update the SwitchToLiveMode for consistency
func (e *ExchangeService) SwitchToLiveMode(ctx context.Context) error {
	e.modeMutex.Lock()
	defer e.modeMutex.Unlock()

	if e.currentMode == "live" {
		return nil // Already in live mode
	}

	slog.Info("Switching to live mode...")

	// Check if data processing was running BEFORE stopping
	e.runMutex.RLock()
	wasRunning := e.isRunning
	e.runMutex.RUnlock()

	// STEP 1: STOP data processing completely if it was running
	if wasRunning {
		slog.Info("Stopping data processing before mode switch...")
		if err := e.StopDataProcessing(); err != nil {
			slog.Error("Failed to stop data processing", "error", err)
			return fmt.Errorf("failed to stop data processing: %w", err)
		}
		slog.Info("Data processing stopped successfully")
	}

	// STEP 2: Switch to live adapters
	e.activeAdapters = e.liveAdapters
	e.currentMode = "live"

	// STEP 3: START data processing with live adapters if it was running before
	if wasRunning {
		slog.Info("Starting data processing with live adapters...")
		bgCtx := context.Background()
		if err := e.StartDataProcessing(bgCtx); err != nil {
			return fmt.Errorf("failed to restart data processing: %w", err)
		}
		slog.Info("Data processing started successfully with live adapters")
	}

	slog.Info("Switched to live mode successfully")
	return nil
}

func (e *ExchangeService) GetCurrentMode() string {
	e.modeMutex.RLock()
	defer e.modeMutex.RUnlock()
	return e.currentMode
}

// internal/core/service/exchange/service.go
// Replace the StartDataProcessing method with this enhanced version

func (e *ExchangeService) StartDataProcessing(ctx context.Context) error {
	slog.Info("=== StartDataProcessing CALLED ===", "mode", e.currentMode)
	e.runMutex.Lock()
	defer e.runMutex.Unlock()

	if e.isRunning {
		slog.Warn("Data processing already running - this should not happen during mode switch!")
		return fmt.Errorf("data processing already running")
	}

	slog.Info("Starting data processing...", "mode", e.currentMode, "workers", e.numWorkers)

	// Debug: Show active adapters
	slog.Info("Active adapters count", "count", len(e.activeAdapters), "mode", e.currentMode)
	for i, adapter := range e.activeAdapters {
		slog.Info("Available adapter", "index", i, "name", adapter.Name(), "healthy_before_start", adapter.IsHealthy())
	}

	// Create a background context that won't be cancelled by HTTP requests
	e.ctx, e.cancel = context.WithCancel(context.Background())

	// Initialize worker pool channels
	for i := 0; i < e.numWorkers; i++ {
		e.workerPool[i] = make(chan domain.MarketData, 100)
	}

	// Start exchange adapters and collect their channels
	var inputChannels []<-chan domain.MarketData
	for i, adapter := range e.activeAdapters {
		slog.Info("Starting adapter", "index", i, "name", adapter.Name(), "mode", e.currentMode)

		dataChan, err := adapter.Start(e.ctx)
		if err != nil {
			slog.Error("Failed to start adapter", "adapter", adapter.Name(), "error", err)
			continue
		}
		inputChannels = append(inputChannels, dataChan)
		slog.Info("Started exchange adapter", "name", adapter.Name(), "healthy_after_start", adapter.IsHealthy())
	}

	slog.Info("Input channels collected", "count", len(inputChannels))

	if len(inputChannels) == 0 {
		slog.Error("No exchange adapters started successfully")
		return fmt.Errorf("no exchange adapters started successfully")
	}

	// Start concurrency pipeline
	e.wg.Add(1)
	go e.fanIn(inputChannels)

	e.wg.Add(1)
	go e.distributor()

	for i := 0; i < e.numWorkers; i++ {
		e.wg.Add(1)
		go e.worker(i, e.workerPool[i])
	}

	e.isRunning = true
	slog.Info("Data processing started successfully", "exchanges", len(inputChannels), "workers", e.numWorkers)

	// Debug: Verify all adapters are healthy after starting
	healthyCount := 0
	for _, adapter := range e.activeAdapters {
		if adapter.IsHealthy() {
			healthyCount++
		}
	}
	slog.Info("Adapter health summary", "healthy", healthyCount, "total", len(e.activeAdapters))

	return nil
}

func (e *ExchangeService) GetDataStream() <-chan domain.MarketData {
	return e.resultChan
}

// Fan-in: Aggregates data from multiple exchange channels into one
func (e *ExchangeService) fanIn(inputChannels []<-chan domain.MarketData) {
	defer e.wg.Done()
	defer close(e.aggregatedDataChan)

	slog.Info("Starting fan-in aggregator", "inputs", len(inputChannels))

	var fanInWg sync.WaitGroup

	// Start a goroutine for each input channel
	for i, ch := range inputChannels {
		fanInWg.Add(1)
		go func(id int, inputChan <-chan domain.MarketData) {
			defer fanInWg.Done()

			for {
				select {
				case data, ok := <-inputChan:
					if !ok {
						slog.Info("Input channel closed", "channel", id)
						return
					}

					select {
					case e.aggregatedDataChan <- data:
					case <-e.ctx.Done():
						return
					}

				case <-e.ctx.Done():
					return
				}
			}
		}(i, ch)
	}

	fanInWg.Wait()
	slog.Info("Fan-in aggregator completed")
}

// Distributor: Fan-out data to worker pool
func (e *ExchangeService) distributor() {
	defer e.wg.Done()

	slog.Info("Starting distributor", "workers", e.numWorkers)
	workerIndex := 0

	for {
		select {
		case data, ok := <-e.aggregatedDataChan:
			if !ok {
				slog.Info("Aggregated data channel closed")
				return
			}

			// Round-robin distribution to workers
			select {
			case e.workerPool[workerIndex] <- data:
				workerIndex = (workerIndex + 1) % e.numWorkers
			case <-time.After(100 * time.Millisecond):
				slog.Warn("Worker pool full, dropping data", "worker", workerIndex)
				workerIndex = (workerIndex + 1) % e.numWorkers
			case <-e.ctx.Done():
				return
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// internal/core/service/exchange/service.go
// Replace the worker method with this improved version

func (e *ExchangeService) worker(id int, workerChan <-chan domain.MarketData) {
	defer e.wg.Done()

	slog.Debug("Worker started", "id", id)
	defer slog.Debug("Worker stopped", "id", id)

	processedCount := 0
	droppedCount := 0

	for {
		select {
		case data, ok := <-workerChan:
			if !ok {
				slog.Debug("Worker channel closed", "id", id, "processed", processedCount, "dropped", droppedCount)
				return
			}

			// Process the data (validation, enrichment, etc.)
			processedData := e.processMarketData(data)

			// Send to result channel with longer timeout but still track drops
			select {
			case e.resultChan <- processedData:
				processedCount++
				// Log every 100 successful processes (less frequent logging)
				if processedCount%100 == 0 {
					slog.Debug("Worker processed batch", "worker", id, "processed", processedCount, "dropped", droppedCount)
				}
			case <-time.After(500 * time.Millisecond): // Increased timeout from 100ms to 500ms
				droppedCount++
				// Only log every 20 drops to reduce log spam
				if droppedCount%20 == 0 {
					slog.Warn("Result channel still full, dropping processed data", "worker", id, "dropped", droppedCount, "processed", processedCount)
				}
			case <-e.ctx.Done():
				return
			}

		case <-e.ctx.Done():
			slog.Info("Worker context cancelled", "id", id, "processed", processedCount, "dropped", droppedCount)
			return
		}
	}
}

// processMarketData validates and enriches market data
func (e *ExchangeService) processMarketData(data domain.MarketData) domain.MarketData {
	// Validate required fields

	if data.Symbol == "" || data.Price <= 0 {
		slog.Warn("Invalid market data", "symbol", data.Symbol, "price", data.Price, "exchange", data.Exchange)
		return data
	}

	// Validate symbol is supported
	if !exchanges.IsSymbolSupported(data.Symbol) {
		slog.Warn("Unsupported symbol", "symbol", data.Symbol, "exchange", data.Exchange)
		return data
	}

	// Ensure timestamp is set
	if data.Timestamp == 0 {
		data.Timestamp = time.Now().Unix()
	}

	// Validate price range (basic sanity check)
	if data.Price > 1000000 || data.Price < 0.0001 {
		slog.Warn("Price out of expected range", "symbol", data.Symbol, "price", data.Price, "exchange", data.Exchange)
	}

	return data
}

// stopCurrentAdapters stops all currently active adapters
func (e *ExchangeService) stopCurrentAdapters() error {
	var errors []error

	for _, adapter := range e.activeAdapters {
		if err := adapter.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop adapter %s: %w", adapter.Name(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors stopping adapters: %v", errors)
	}

	return nil
}

// GetActiveAdapters returns the currently active adapters (for monitoring)
func (e *ExchangeService) GetActiveAdapters() []port.ExchangeAdapter {
	e.modeMutex.RLock()
	defer e.modeMutex.RUnlock()

	result := make([]port.ExchangeAdapter, len(e.activeAdapters))
	copy(result, e.activeAdapters)
	return result
}

// IsRunning returns whether data processing is currently running
func (e *ExchangeService) IsRunning() bool {
	e.runMutex.RLock()
	defer e.runMutex.RUnlock()
	return e.isRunning
}

// GetStats returns basic statistics about the service
func (e *ExchangeService) GetStats() map[string]interface{} {
	e.modeMutex.RLock()
	e.runMutex.RLock()
	defer e.modeMutex.RUnlock()
	defer e.runMutex.RUnlock()

	healthyAdapters := 0
	for _, adapter := range e.activeAdapters {
		if adapter.IsHealthy() {
			healthyAdapters++
		}
	}

	return map[string]interface{}{
		"current_mode":      e.currentMode,
		"is_running":        e.isRunning,
		"active_adapters":   len(e.activeAdapters),
		"healthy_adapters":  healthyAdapters,
		"num_workers":       e.numWorkers,
		"aggregated_buffer": len(e.aggregatedDataChan),
		"result_buffer":     len(e.resultChan),
	}
}
