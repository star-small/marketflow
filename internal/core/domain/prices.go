package domain

import (
	"time"
)

// Existing types (keep these as-is)
type Prices struct {
	PairName     string    `json:"pair_name"`     // The trading pair name (e.g., BTC/USD)
	Exchange     string    `json:"exchange"`      // Exchange from which the data was received
	Timestamp    time.Time `json:"timestamp"`     // Time when the data is stored
	AveragePrice float64   `json:"average_price"` // Average price over the last minute
	MinPrice     float64   `json:"min_price"`     // Minimum price over the last minute
	MaxPrice     float64   `json:"max_price"`     // Maximum price over the last minute
}

type GetPrice struct {
	Price int
}

// ADD THESE NEW TYPES:

// PriceAggregation represents aggregated price data for a specific time period
type PriceAggregation struct {
	PairName     string    `json:"pair_name"`     // The trading pair name (e.g., BTCUSDT)
	Exchange     string    `json:"exchange"`      // Exchange from which the data was received
	Timestamp    time.Time `json:"timestamp"`     // Time when the data is stored (rounded to minute)
	AveragePrice float64   `json:"average_price"` // Average price over the time period
	MinPrice     float64   `json:"min_price"`     // Minimum price over the time period
	MaxPrice     float64   `json:"max_price"`     // Maximum price over the time period
	DataPoints   int       `json:"data_points"`   // Number of price updates used for aggregation
}

// PriceStatistics represents statistical data about prices
type PriceStatistics struct {
	Symbol     string    `json:"symbol"`
	Exchange   string    `json:"exchange,omitempty"`
	Price      float64   `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
	Period     string    `json:"period,omitempty"`
	DataPoints int       `json:"data_points,omitempty"`
}
