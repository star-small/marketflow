package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:40102", // Redis server port
		Password: "",
	})

	type MarketData struct {
		Symbol    string  `json:"symbol"`
		Price     float64 `json:"price"`
		Timestamp int64   `json:"timestamp"`
	}

	ctx := context.Background()

	// Get the JSON document - you need to specify the key name
	// Replace "marketdata:1" with your actual key
	result, err := client.JSONGet(ctx, "marketdata:1", ".").Result()
	if err != nil {
		fmt.Println("Redis JSON error:", err)
		return
	}

	var marketData MarketData
	err = json.Unmarshal([]byte(result), &marketData)
	if err != nil {
		fmt.Println("JSON unmarshal error:", err)
		return
	}

	fmt.Printf("Symbol: %s, Price: %.2f, Timestamp: %d\n",
		marketData.Symbol, marketData.Price, marketData.Timestamp)
}
