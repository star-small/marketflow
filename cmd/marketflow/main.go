package main

import (
	"context"
	"flag"
	"fmt"
	//"log/slog"
	"os"
	"os/signal"
	"syscall"

	"marketflow/internal/adapters/cache/redis"
	"marketflow/internal/adapters/exchange/live"
	"marketflow/internal/adapters/exchange/test"
	"marketflow/internal/adapters/storage/postgresql"
	"marketflow/internal/adapters/web"
	"marketflow/internal/application/usecases"
	"marketflow/internal/concurrency"
	"marketflow/internal/config"
	"marketflow/internal/logger"
)

func main() {
	var (
		port = flag.Int("port", 8080, "Port number")
		help = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		printUsage()
		return
	}

	// Initialize logger
	log := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize components
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	storage, err := postgresql.New(cfg.Database)
	if err != nil {
		log.Error("Failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer storage.Close()

	// Initialize cache
	cache, err := redis.New(cfg.Cache)
	if err != nil {
		log.Error("Failed to initialize cache", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	// Initialize exchange adapters
	liveExchange := live.New(cfg.Exchanges)
	testExchange := test.New()

	// Initialize concurrency manager
	concurrencyManager := concurrency.NewManager(log)

	// Initialize use cases
	marketDataUseCase := usecases.NewMarketDataUseCase(storage, cache, log)
	dataProcessingUseCase := usecases.NewDataProcessingUseCase(storage, cache, concurrencyManager, log)

	// Initialize web server
	webServer := web.NewServer(*port, marketDataUseCase, dataProcessingUseCase, log)

	// Start data processing
	go func() {
		if err := dataProcessingUseCase.Start(ctx, liveExchange, testExchange); err != nil {
			log.Error("Failed to start data processing", "error", err)
			cancel()
		}
	}()

	// Start web server
	go func() {
		if err := webServer.Start(); err != nil {
			log.Error("Failed to start web server", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Received shutdown signal")
	case <-ctx.Done():
		log.Info("Context cancelled")
	}

	// Graceful shutdown
	log.Info("Shutting down gracefully...")
	webServer.Shutdown(ctx)
	log.Info("Shutdown complete")
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  marketflow [--port <N>]")
	fmt.Println("  marketflow --help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --port N     Port number")
}
