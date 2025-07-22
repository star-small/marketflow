package test

import (
	"context"
	"math/rand"
	"time"

	"marketflow/internal/application/ports"
	"marketflow/internal/domain/models"
)

// Adapter implements the ExchangePort interface for test data
type Adapter struct {
	connected bool
}

// New creates a new test exchange adapter
func New() ports.ExchangePort {
	return &Adapter{
		connected: false,
	}
}

// Start begins data collection
func (a *Adapter) Start(ctx context.Context) (<-chan models.PriceUpdate, error) {
	updateCh := make(chan models.PriceUpdate, 1000)

	symbols := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}

	// Base prices for each symbol
	basePrices := map[string]float64{
		"BTCUSDT":  99000.0,
		"DOGEUSDT": 0.30,
		"TONUSDT":  3.90,
		"SOLUSDT":  200.0,
		"ETHUSDT":  3000.0,
	}

	exchanges := []string{"test-exchange1", "test-exchange2", "test-exchange3"}

	// Start generators for each exchange
	for _, exchange := range exchanges {
		go a.generateData(ctx, symbols, basePrices, exchange, updateCh)
	}

	a.connected = true
	return updateCh, nil
}

// Stop stops data collection
func (a *Adapter) Stop() error {
	a.connected = false
	return nil
}

// IsConnected returns connection status
func (a *Adapter) IsConnected() bool {
	return a.connected
}

// GetName returns the exchange name
func (a *Adapter) GetName() string {
	return "test"
}

func (a *Adapter) generateData(ctx context.Context, symbols []string, basePrices map[string]float64, exchange string, updateCh chan<- models.PriceUpdate) {
	ticker := time.NewTicker(100 * time.Millisecond) // Generate data every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, symbol := range symbols {
				basePrice := basePrices[symbol]

				// Add some random variation (±2%)
				variation := (rand.Float64() - 0.5) * 0.04 // -2% to +2%
				price := basePrice * (1 + variation)

				update := models.PriceUpdate{
					Symbol:     symbol,
					Price:      price,
					Timestamp:  time.Now().UnixMilli(),
					Exchange:   exchange,
					ReceivedAt: time.Now(),
				}

				select {
				case updateCh <- update:
				case <-ctx.Done():
					return
				default:
					// Channel is full, skip this update
				}
			}
		}
	}
}
