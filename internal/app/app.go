package app

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"crypto/internal/config"
	"crypto/internal/server"
)

const cfgPath = "./config/config.json"

func Start() error {
	var (
		port     = flag.Int("port", 8080, "Port number")
		helpFlag = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  marketflow [--port <N>]\n")
		fmt.Fprintf(os.Stderr, "  marketflow --help\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --port N     Port number\n")
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return nil
	}

	slog.Info("Loading configuration...")
	config, err := config.GetConfig(cfgPath)
	if err != nil {
		slog.Error("failed to get config", "error", err)
		return fmt.Errorf("failed to get config: %w", err)
	}

	if *port > 0 {
		config.App.Port = *port
	}
	slog.Info("Configuration loaded", "port", config.App.Port)

	slog.Info("Creating application instance...")
	app := server.NewApp(config)

	slog.Info("Initializing application...")
	if err := app.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	// Set up graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Starting server...")
		app.Run()
		serverErrors <- nil
	}()

	// Block until we receive a shutdown signal or server error
	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		slog.Info("Server stopped normally")

	case sig := <-shutdown:
		slog.Info("Received shutdown signal", "signal", sig)

		// Graceful shutdown
		if err := app.Shutdown(); err != nil {
			slog.Error("Failed to shutdown gracefully", "error", err)
			return fmt.Errorf("failed to shutdown gracefully: %w", err)
		}
	}

	slog.Info("Application stopped successfully")
	return nil
}
