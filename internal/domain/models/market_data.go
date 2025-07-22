package models

import "time"

// PriceUpdate represents a real-time price update from an exchange
type PriceUpdate struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp int64     `json:"timestamp"`
	Exchange  string    `json:"exchange"`
	ReceivedAt time.Time `json:"received_at"`
}

// AggregatedData represents aggregated market data stored in PostgreSQL
type AggregatedData struct {
	ID           int64     `db:"id"`
	PairName     string    `db:"pair_name"`
	Exchange     string    `db:"exchange"`
	Timestamp    time.Time `db:"timestamp"`
	AveragePrice float64   `db:"average_price"`
	MinPrice     float64   `db:"min_price"`
	MaxPrice     float64   `db:"max_price"`
}

// LatestPrice represents cached latest price data in Redis
type LatestPrice struct {
	Symbol    string    `json:"symbol"`
	Exchange  string    `json:"exchange"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// DataMode represents the current data mode (live or test)
type DataMode string

const (
	DataModeLive DataMode = "live"
	DataModeTest DataMode = "test"
)
