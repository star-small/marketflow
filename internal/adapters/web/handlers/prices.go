package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"marketflow/internal/application/usecases"
)

// PricesHandler handles price-related requests
type PricesHandler struct {
	marketDataUseCase *usecases.MarketDataUseCase
	logger            *slog.Logger
}

// NewPricesHandler creates a new prices handler
func NewPricesHandler(marketDataUseCase *usecases.MarketDataUseCase, logger *slog.Logger) *PricesHandler {
	return &PricesHandler{
		marketDataUseCase: marketDataUseCase,
		logger:            logger,
	}
}

// Handle handles price requests
func (h *PricesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/prices/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	operation := parts[0]

	var exchange, symbol string
	var period time.Duration

	// Parse path parameters
	if len(parts) == 2 {
		symbol = parts[1]
	} else if len(parts) == 3 {
		exchange = parts[1]
		symbol = parts[2]
	}

	// Parse period from query parameters
	if periodStr := r.URL.Query().Get("period"); periodStr != "" {
		var err error
		period, err = parsePeriod(periodStr)
		if err != nil {
			http.Error(w, "Invalid period format", http.StatusBadRequest)
			return
		}
	} else {
		period = time.Minute // default period
	}

	ctx := r.Context()
	var response interface{}
	var err error

	switch operation {
	case "latest":
		response, err = h.marketDataUseCase.GetLatestPrice(ctx, symbol, exchange)
	case "highest":
		response, err = h.marketDataUseCase.GetHighestPrice(ctx, symbol, exchange, period)
	case "lowest":
		response, err = h.marketDataUseCase.GetLowestPrice(ctx, symbol, exchange, period)
	case "average":
		response, err = h.marketDataUseCase.GetAveragePrice(ctx, symbol, exchange, period)
	default:
		http.Error(w, "Unknown operation", http.StatusBadRequest)
		return
	}

	if err != nil {
		h.logger.Error("Failed to process request", "error", err, "operation", operation)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if response == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func parsePeriod(periodStr string) (time.Duration, error) {
	// Handle different period formats: 1s, 3s, 5s, 10s, 30s, 1m, 3m, 5m
	if strings.HasSuffix(periodStr, "s") {
		seconds, err := strconv.Atoi(strings.TrimSuffix(periodStr, "s"))
		if err != nil {
			return 0, err
		}
		return time.Duration(seconds) * time.Second, nil
	}

	if strings.HasSuffix(periodStr, "m") {
		minutes, err := strconv.Atoi(strings.TrimSuffix(periodStr, "m"))
		if err != nil {
			return 0, err
		}
		return time.Duration(minutes) * time.Minute, nil
	}

	return time.ParseDuration(periodStr)
}
