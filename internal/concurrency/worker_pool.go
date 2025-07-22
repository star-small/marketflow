package concurrency

import (
	"context"
	"log/slog"
	"sync"

	"marketflow/internal/domain/models"
)

// WorkerPool manages a pool of workers for processing price updates
type WorkerPool struct {
	workers int
	logger  *slog.Logger
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		logger:  logger,
		done:    make(chan struct{}),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context, inputCh <-chan models.PriceUpdate, outputCh chan<- models.PriceUpdate) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i, inputCh, outputCh)
	}

	wp.wg.Wait()
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
}

func (wp *WorkerPool) worker(ctx context.Context, id int, inputCh <-chan models.PriceUpdate, outputCh chan<- models.PriceUpdate) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", "worker_id", id)
	defer wp.logger.Debug("Worker stopped", "worker_id", id)

	for {
		select {
		case <-ctx.Done():
			return
		case <-wp.done:
			return
		case update, ok := <-inputCh:
			if !ok {
				return
			}

			// Process the update (validation, transformation, etc.)
			processedUpdate := wp.processUpdate(update)

			select {
			case outputCh <- processedUpdate:
			case <-ctx.Done():
				return
			case <-wp.done:
				return
			}
		}
	}
}

func (wp *WorkerPool) processUpdate(update models.PriceUpdate) models.PriceUpdate {
	// Add any processing logic here (validation, transformation, etc.)
	return update
}
