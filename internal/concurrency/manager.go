package concurrency

import (
	"context"
	"log/slog"
	"sync"

	"marketflow/internal/domain/models"
)

// Manager handles concurrency patterns for data processing
type Manager struct {
	logger      *slog.Logger
	workerPools map[string]*WorkerPool
	mu          sync.RWMutex
}

// NewManager creates a new concurrency manager
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		logger:      logger,
		workerPools: make(map[string]*WorkerPool),
	}
}

// StartWorkerPool starts a worker pool for an exchange
func (m *Manager) StartWorkerPool(ctx context.Context, exchange string, workers int, inputCh <-chan models.PriceUpdate, outputCh chan<- models.PriceUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workerPools[exchange]; exists {
		return
	}

	pool := NewWorkerPool(workers, m.logger)
	m.workerPools[exchange] = pool

	go pool.Start(ctx, inputCh, outputCh)
}

// StopWorkerPool stops a worker pool for an exchange
func (m *Manager) StopWorkerPool(exchange string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pool, exists := m.workerPools[exchange]; exists {
		pool.Stop()
		delete(m.workerPools, exchange)
	}
}

// FanIn aggregates multiple input channels into a single output channel
func (m *Manager) FanIn(ctx context.Context, inputs []<-chan models.PriceUpdate) <-chan models.PriceUpdate {
	output := make(chan models.PriceUpdate)

	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan models.PriceUpdate) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case update, ok := <-ch:
					if !ok {
						return
					}
					select {
					case output <- update:
					case <-ctx.Done():
						return
					}
				}
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// FanOut distributes input from a single channel to multiple output channels
func (m *Manager) FanOut(ctx context.Context, input <-chan models.PriceUpdate, outputs []chan<- models.PriceUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-input:
				if !ok {
					return
				}
				for _, output := range outputs {
					select {
					case output <- update:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
}
