package domain

type MarketData struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Timestep int64   `json:"timestamp"`
	Exchange string  `json:"exchange"`
}
