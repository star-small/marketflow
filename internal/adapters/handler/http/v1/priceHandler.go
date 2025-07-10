package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/internal/core/port"
)

type PriceHandler struct {
	priceService port.PriceService
}

func NewPriceHandler(
	priceService port.PriceService,
) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
	}
}

// Response structures
type LatestPriceResponse struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	Exchange  string  `json:"exchange,omitempty"` // omitempty for cross-exchange responses
}

type PriceStatisticsResponse struct {
	Symbol     string    `json:"symbol"`
	Exchange   string    `json:"exchange,omitempty"`
	Price      float64   `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
	Period     string    `json:"period,omitempty"`
	DataPoints int       `json:"data_points,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Supported symbols
var supportedSymbols = map[string]bool{
	"BTCUSDT":  true,
	"DOGEUSDT": true,
	"TONUSDT":  true,
	"SOLUSDT":  true,
	"ETHUSDT":  true,
}

// Default periods for statistics
var defaultPeriod = 1 * time.Hour

func (h *PriceHandler) GetLatestPrice(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	// Normalize symbol to uppercase
	symbol = strings.ToUpper(symbol)

	// Validate symbol
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	// Call service to get latest price
	marketData, err := h.priceService.GetLatestPrice(r.Context(), symbol)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get latest price: "+err.Error())
		return
	}

	if marketData == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol)
		return
	}

	// Prepare response
	response := LatestPriceResponse{
		Symbol:    marketData.Symbol,
		Price:     marketData.Price,
		Timestamp: marketData.Timestamp,
		Exchange:  marketData.Exchange,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetLatestPriceByExchange(w http.ResponseWriter, r *http.Request) {
	// Extract exchange and symbol from URL path
	exchange := r.PathValue("exchange")
	symbol := r.PathValue("symbol")

	if exchange == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing exchange parameter")
		return
	}

	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	// Normalize symbol to uppercase
	symbol = strings.ToUpper(symbol)

	// Validate symbol
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	// Call service to get latest price by exchange
	marketData, err := h.priceService.GetLatestPriceByExchange(r.Context(), symbol, exchange)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get latest price: "+err.Error())
		return
	}

	if marketData == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol+" on exchange: "+exchange)
		return
	}

	// Prepare response
	response := LatestPriceResponse{
		Symbol:    marketData.Symbol,
		Price:     marketData.Price,
		Timestamp: marketData.Timestamp,
		Exchange:  marketData.Exchange,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetHighestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetHighestPrice(r.Context(), symbol, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get highest price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetHighestPriceByExchange(w http.ResponseWriter, r *http.Request) {
	exchange := r.PathValue("exchange")
	symbol := r.PathValue("symbol")

	if exchange == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing exchange parameter")
		return
	}

	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetHighestPriceByExchange(r.Context(), symbol, exchange, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get highest price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol+" on exchange: "+exchange)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetLowestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetLowestPrice(r.Context(), symbol, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get lowest price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetLowestPriceByExchange(w http.ResponseWriter, r *http.Request) {
	exchange := r.PathValue("exchange")
	symbol := r.PathValue("symbol")

	if exchange == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing exchange parameter")
		return
	}

	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetLowestPriceByExchange(r.Context(), symbol, exchange, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get lowest price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol+" on exchange: "+exchange)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetAveragePrice(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetAveragePrice(r.Context(), symbol, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get average price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *PriceHandler) GetAveragePriceByExchange(w http.ResponseWriter, r *http.Request) {
	exchange := r.PathValue("exchange")
	symbol := r.PathValue("symbol")

	if exchange == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing exchange parameter")
		return
	}

	if symbol == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing symbol parameter")
		return
	}

	symbol = strings.ToUpper(symbol)
	if !supportedSymbols[symbol] {
		h.writeErrorResponse(w, http.StatusBadRequest, "unsupported symbol: "+symbol)
		return
	}

	period := h.parsePeriod(r.URL.Query().Get("period"))

	stats, err := h.priceService.GetAveragePriceByExchange(r.Context(), symbol, exchange, period)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get average price: "+err.Error())
		return
	}

	if stats == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "no price data found for symbol: "+symbol+" on exchange: "+exchange)
		return
	}

	response := PriceStatisticsResponse{
		Symbol:     stats.Symbol,
		Exchange:   stats.Exchange,
		Price:      stats.Price,
		Timestamp:  stats.Timestamp,
		Period:     stats.Period,
		DataPoints: stats.DataPoints,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (h *PriceHandler) parsePeriod(periodStr string) time.Duration {
	if periodStr == "" {
		return defaultPeriod
	}

	// Try to parse as duration string (e.g., "1h", "30m", "5s")
	if duration, err := time.ParseDuration(periodStr); err == nil {
		return duration
	}

	// Try to parse as seconds
	if seconds, err := strconv.Atoi(periodStr); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Return default if parsing fails
	return defaultPeriod
}

func (h *PriceHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, log the error and send a simple error message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error","message":"failed to encode response"}`))
	}
}

func (h *PriceHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorType := "bad_request"
	switch statusCode {
	case http.StatusNotFound:
		errorType = "not_found"
	case http.StatusInternalServerError:
		errorType = "internal_error"
	}

	response := ErrorResponse{
		Error:   errorType,
		Message: message,
	}

	h.writeJSONResponse(w, statusCode, response)
}
