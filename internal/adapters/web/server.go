package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"marketflow/internal/adapters/web/handlers"
	"marketflow/internal/application/usecases"
)

// Server represents the HTTP server
type Server struct {
	port                  int
	marketDataUseCase     *usecases.MarketDataUseCase
	dataProcessingUseCase *usecases.DataProcessingUseCase
	logger                *slog.Logger
	server                *http.Server
}

// NewServer creates a new HTTP server
func NewServer(port int, marketDataUseCase *usecases.MarketDataUseCase, dataProcessingUseCase *usecases.DataProcessingUseCase, logger *slog.Logger) *Server {
	return &Server{
		port:                  port,
		marketDataUseCase:     marketDataUseCase,
		dataProcessingUseCase: dataProcessingUseCase,
		logger:                logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Initialize handlers
	pricesHandler := handlers.NewPricesHandler(s.marketDataUseCase, s.logger)
	modeHandler := handlers.NewModeHandler(s.dataProcessingUseCase, s.logger)
	healthHandler := handlers.NewHealthHandler(s.logger)
	statusHandler := handlers.NewStatusHandler(s.dataProcessingUseCase, s.logger)

	// Register routes
	mux.HandleFunc("/prices/", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Prices request", "method", r.Method, "path", r.URL.Path)
		pricesHandler.Handle(w, r)
	})

	mux.HandleFunc("/mode/", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Mode request", "method", r.Method, "path", r.URL.Path)
		modeHandler.Handle(w, r)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Health request", "method", r.Method, "path", r.URL.Path)
		healthHandler.Handle(w, r)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Status request", "method", r.Method, "path", r.URL.Path)
		statusHandler.Handle(w, r)
	})

	// Add a catch-all for debugging
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Unmatched request", "method", r.Method, "path", r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/prices/") {
			pricesHandler.Handle(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/mode/") {
			modeHandler.Handle(w, r)
		} else if r.URL.Path == "/health" {
			healthHandler.Handle(w, r)
		} else if r.URL.Path == "/status" {
			statusHandler.Handle(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.logger.Info("Starting HTTP server", "port", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
