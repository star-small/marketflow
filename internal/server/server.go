package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"crypto/internal/adapters/cache"
	v1 "crypto/internal/adapters/handler/http/v1"
	"crypto/internal/adapters/repository/postgres"
	"crypto/internal/config"
	"crypto/internal/core/domain"
	"crypto/internal/core/port"
	"crypto/internal/core/service/aggregation"
	"crypto/internal/core/service/exchange"
	"crypto/internal/core/service/health"
	"crypto/internal/core/service/mode"
	"crypto/internal/core/service/prices"

	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

type App struct {
	cfg         *config.Config
	router      *http.ServeMux
	db          *sql.DB
	redisClient *redis.Client

	// Services
	exchangeService port.ExchangeService
	priceService    port.PriceService
	healthService   port.HealthService
	modeService     port.ModeService
	priceAggregator *aggregation.PriceAggregator

	// Repositories
	priceRepository port.PriceRepository

	// Cache
	cacheAdapter port.Cache

	// For graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

func NewApp(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (app *App) Initialize() error {
	slog.Info("Initializing application...")
	app.router = http.NewServeMux()

	// Database connection
	dbConn, err := postgres.NewDbConnInstance(&app.cfg.Repository)
	if err != nil {
		slog.Error("Connection to database failed", "error", err)
		return err
	}
	app.db = dbConn
	slog.Info("Database connected successfully")

	// Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", app.cfg.Cache.RedisHost, app.cfg.Cache.RedisPort),
		Password:     app.cfg.Cache.RedisPassword,
		DB:           app.cfg.Cache.RedisDB,
		PoolSize:     app.cfg.Cache.PoolSize,
		MinIdleConns: app.cfg.Cache.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cacheAdapter port.Cache
	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis connection failed, continuing without cache", "error", err)
		app.redisClient = nil
		cacheAdapter = nil
	} else {
		app.redisClient = redisClient
		cacheAdapter = cache.NewRedisAdapter(redisClient)
		slog.Info("Redis connected successfully")
	}
	app.cacheAdapter = cacheAdapter

	// Initialize repositories
	app.priceRepository = postgres.NewPriceRepository(app.db)

	// Initialize services following hexagonal architecture

	// 1. Create Exchange Service (handles data collection)
	app.exchangeService = exchange.NewExchangeService()

	// 2. Create Price Service (business logic layer)
	app.priceService = prices.NewPriceService(cacheAdapter, app.priceRepository)

	// 3. Create Health Service
	app.healthService = health.NewHealthService(app.db, cacheAdapter, app.exchangeService)

	// 4. Create Mode Service
	app.modeService = mode.NewModeService(app.exchangeService)

	// 5. CRITICAL FIX: Create Price Aggregator
	slog.Info("Creating price aggregator...")
	app.priceAggregator = aggregation.NewPriceAggregator(app.priceService)
	slog.Info("Price aggregator created successfully")

	// 6. Create Handlers (adapters layer)
	priceHandler := v1.NewPriceHandler(app.priceService)
	healthHandler := v1.NewHealthHandler(app.healthService)
	modeHandler := v1.NewModeHandler(app.modeService)

	// 7. Set up routes
	v1.SetMarketRoutes(app.router, priceHandler, healthHandler, modeHandler)

	// 8. Start background data processing
	go app.startMarketDataProcessor()

	slog.Info("Application initialized successfully")
	return nil
}
func (app *App) initializeDatabase() error {
	dbConn, err := postgres.NewDbConnInstance(&app.cfg.Repository)
	if err != nil {
		slog.Error("Connection to database failed", "error", err)
		return err
	}
	app.db = dbConn
	slog.Info("Database connected successfully")
	return nil
}

// internal/server/server.go
// Replace the initializeCache method with this version

func (app *App) initializeCache() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", app.cfg.Cache.RedisHost, app.cfg.Cache.RedisPort),
		Password:     app.cfg.Cache.RedisPassword,
		DB:           app.cfg.Cache.RedisDB,
		PoolSize:     app.cfg.Cache.PoolSize,
		MinIdleConns: app.cfg.Cache.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test Redis connection with retry
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var lastErr error
	for attempts := 0; attempts < 3; attempts++ {
		if err := redisClient.Ping(ctx).Err(); err != nil {
			lastErr = err
			slog.Warn("Redis connection attempt failed", "attempt", attempts+1, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Success!
		app.redisClient = redisClient
		app.cacheAdapter = cache.NewRedisAdapter(redisClient)
		slog.Info("Redis connected successfully")
		return nil
	}

	// All attempts failed
	slog.Error("Failed to connect to Redis after 3 attempts", "lastError", lastErr)

	// OPTION 1: Fail hard (recommended for production)
	return fmt.Errorf("failed to connect to Redis: %w", lastErr)

	// OPTION 2: Continue without cache (for development)
	// app.redisClient = nil
	// app.cacheAdapter = nil
	// slog.Warn("Running without Redis cache - data will not be stored")
	// return nil
}

func (app *App) initializeServices() error {
	// 1. Create Exchange Service (handles data collection)
	app.exchangeService = exchange.NewExchangeService()

	// 2. Create Price Service (business logic layer)
	app.priceService = prices.NewPriceService(app.cacheAdapter, app.priceRepository)

	// 3. Create Health Service
	app.healthService = health.NewHealthService(app.db, app.cacheAdapter, app.exchangeService)

	// 4. Create Mode Service
	app.modeService = mode.NewModeService(app.exchangeService)

	// 5. Create Price Aggregator
	app.priceAggregator = aggregation.NewPriceAggregator(app.priceService)
	slog.Info("Services initialized successfully")
	slog.Info("Services initialized successfully")
	return nil
}

func (app *App) initializeHandlers() error {
	// Create Handlers (adapters layer)
	priceHandler := v1.NewPriceHandler(app.priceService)
	healthHandler := v1.NewHealthHandler(app.healthService)
	modeHandler := v1.NewModeHandler(app.modeService)

	// Set up routes
	v1.SetMarketRoutes(app.router, priceHandler, healthHandler, modeHandler)

	slog.Info("HTTP handlers initialized successfully")
	return nil
}

func (app *App) startBackgroundProcesses() error {
	// Start price aggregator
	slog.Info("Starting price aggregator...")
	if app.priceAggregator != nil {
		app.priceAggregator.Start()
		slog.Info("Price aggregator started successfully")
	} else {
		slog.Error("Price aggregator is nil, cannot start")
	}

	// Start cleanup routines
	if app.redisClient != nil {
		go app.startRedisCleanupRoutine()
	}

	go app.startPostgreSQLCleanupRoutine()

	slog.Info("Background processes started successfully")
	return nil
}

func (app *App) Run() {
	// Start background processes BEFORE starting the HTTP server
	if err := app.startBackgroundProcesses(); err != nil {
		slog.Error("Failed to start background processes", "error", err)
		return
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.cfg.App.Port),
		Handler: app.router,
	}

	slog.Info("Starting server", "port", app.cfg.App.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		panic(err)
	}
}

// Background task for processing market data
func (app *App) startMarketDataProcessor() {
	slog.Info("Starting market data processor...")

	// Start exchange service in live mode by default
	if err := app.exchangeService.SwitchToLiveMode(app.ctx); err != nil {
		slog.Error("Failed to switch to live mode, trying test mode", "error", err)

		// Fallback to test mode if live mode fails
		if err := app.exchangeService.SwitchToTestMode(app.ctx); err != nil {
			slog.Error("Failed to switch to test mode", "error", err)
			return
		}
	}

	// Start data processing
	if err := app.exchangeService.StartDataProcessing(app.ctx); err != nil {
		slog.Error("Failed to start data processing", "error", err)
		return
	}

	// Get data stream from exchange service
	dataStream := app.exchangeService.GetDataStream()

	// Process incoming market data
	go app.processMarketData(dataStream)

	slog.Info("Market data processor started successfully")
}

// internal/server/server.go
// Replace the processMarketData function with this debug version

// internal/server/server.go
// Replace the processMarketData function with this debug version

// internal/server/server.go
// Add this quick fix to the processMarketData function

// internal/server/server.go
// Replace the processMarketData function with this version that connects the aggregator

func (app *App) processMarketData(dataStream <-chan domain.MarketData) {
	slog.Info("🔍 CONSUMER STARTED: Starting market data processing goroutine...")

	processedCount := 0
	errorCount := 0
	aggregatorCount := 0

	for {
		select {
		case data, ok := <-dataStream:
			if !ok {
				slog.Info("🔍 CONSUMER: Market data stream closed", "processed", processedCount, "errors", errorCount, "aggregated", aggregatorCount)
				return
			}

			processedCount++

			// STORE IN REDIS CACHE
			if app.cacheAdapter != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				key := fmt.Sprintf("latest:%s:%s", data.Symbol, data.Exchange)

				if err := app.cacheAdapter.SetPrice(ctx, key, data); err != nil {
					errorCount++
					if errorCount%50 == 0 {
						slog.Error("🔍 CONSUMER: Redis error", "error", err, "errorCount", errorCount)
					}
				}
				cancel()
			}

			// CRITICAL FIX: SEND TO PRICE AGGREGATOR
			if app.priceAggregator != nil {
				// Send data to aggregator for PostgreSQL storage
				app.priceAggregator.ProcessPrice(data)
				aggregatorCount++

				// Log aggregator activity
				if aggregatorCount%100 == 0 {
					slog.Info("🔍 CONSUMER: Sent to aggregator", "aggregated", aggregatorCount, "symbol", data.Symbol, "exchange", data.Exchange)
				}
			} else {
				if processedCount%100 == 0 {
					slog.Warn("🔍 CONSUMER: No price aggregator available")
				}
			}

			// Log progress
			if processedCount%200 == 0 {
				slog.Info("🔍 CONSUMER: Processing progress", "processed", processedCount, "aggregated", aggregatorCount, "errors", errorCount)
			}

		case <-app.ctx.Done():
			slog.Info("🔍 CONSUMER: Context cancelled", "processed", processedCount, "errors", errorCount, "aggregated", aggregatorCount)
			return
		}
	}
}

// startRedisCleanupRoutine cleans up old data from Redis
func (app *App) startRedisCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second) // Clean up every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Clean up data older than 2 minutes
			if err := app.cacheAdapter.CleanupOldData(ctx, 2*time.Minute); err != nil {
				slog.Error("Failed to cleanup old Redis data", "error", err)
			}

			cancel()

		case <-app.ctx.Done():
			slog.Info("Redis cleanup routine stopped")
			return
		}
	}
}

// startPostgreSQLCleanupRoutine cleans up old data from PostgreSQL
func (app *App) startPostgreSQLCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			// Clean up data older than 7 days
			if err := app.priceRepository.CleanupOldData(ctx, 7*24*time.Hour); err != nil {
				slog.Error("Failed to cleanup old PostgreSQL data", "error", err)
			}

			cancel()

		case <-app.ctx.Done():
			slog.Info("PostgreSQL cleanup routine stopped")
			return
		}
	}
}

// Shutdown gracefully shuts down the application
func (app *App) Shutdown() error {
	slog.Info("Shutting down application...")

	// Cancel context to stop all goroutines
	app.cancel()

	// Stop price aggregator
	if app.priceAggregator != nil {
		app.priceAggregator.Stop()
		slog.Info("Price aggregator stopped")
	}

	// Stop exchange service
	if app.exchangeService != nil {
		if err := app.exchangeService.StopDataProcessing(); err != nil {
			slog.Error("Failed to stop exchange service", "error", err)
		}
	}

	// Close database connection
	if app.db != nil {
		if err := app.db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}

	// Close Redis connection
	if app.redisClient != nil {
		if err := app.redisClient.Close(); err != nil {
			slog.Error("Failed to close Redis", "error", err)
		}
	}

	slog.Info("Application shutdown complete")
	return nil
}

// GetStats returns application statistics
func (app *App) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"timestamp": time.Now(),
	}

	// Exchange service stats
	if app.exchangeService != nil {
		if exchSvc, ok := app.exchangeService.(interface{ GetStats() map[string]interface{} }); ok {
			stats["exchange_service"] = exchSvc.GetStats()
		}
	}

	// Aggregator stats
	if app.priceAggregator != nil {
		stats["price_aggregator"] = app.priceAggregator.GetStats()
	}

	// Database status
	if app.db != nil {
		if err := app.db.Ping(); err != nil {
			stats["database"] = "unhealthy"
		} else {
			stats["database"] = "healthy"
		}
	}

	// Cache status
	if app.cacheAdapter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		if err := app.cacheAdapter.Ping(ctx); err != nil {
			stats["cache"] = "unhealthy"
		} else {
			stats["cache"] = "healthy"
		}
		cancel()
	} else {
		stats["cache"] = "unavailable"
	}

	return stats
}
