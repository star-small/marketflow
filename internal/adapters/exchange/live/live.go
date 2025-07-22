package live

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"marketflow/internal/application/ports"
	"marketflow/internal/config"
	"marketflow/internal/domain/models"
)

// Adapter implements the ExchangePort interface for live exchanges
type Adapter struct {
	exchanges []config.ExchangeConfig
	connected bool
}

// New creates a new live exchange adapter
func New(cfg config.ExchangesConfig) ports.ExchangePort {
	exchanges := []config.ExchangeConfig{
		cfg.Exchange1,
		cfg.Exchange2,
		cfg.Exchange3,
	}

	return &Adapter{
		exchanges: exchanges,
		connected: false,
	}
}

// Start begins data collection
func (a *Adapter) Start(ctx context.Context) (<-chan models.PriceUpdate, error) {
	updateCh := make(chan models.PriceUpdate, 1000)

	for i, exchange := range a.exchanges {
		exchangeName := fmt.Sprintf("exchange%d", i+1)
		go a.connectToExchange(ctx, exchange, exchangeName, updateCh)
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
	return "live"
}

func (a *Adapter) connectToExchange(ctx context.Context, cfg config.ExchangeConfig, exchangeName string, updateCh chan<- models.PriceUpdate) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := a.handleExchangeConnection(ctx, cfg, exchangeName, updateCh); err != nil {
				// Wait before reconnecting
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (a *Adapter) handleExchangeConnection(ctx context.Context, cfg config.ExchangeConfig, exchangeName string, updateCh chan<- models.PriceUpdate) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", exchangeName, err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			var update models.PriceUpdate
			if err := json.Unmarshal(scanner.Bytes(), &update); err != nil {
				continue
			}

			update.Exchange = exchangeName
			update.ReceivedAt = time.Now()

			select {
			case updateCh <- update:
			case <-ctx.Done():
				return nil
			default:
				// Channel is full, skip this update
			}
		}
	}

	return scanner.Err()
}
