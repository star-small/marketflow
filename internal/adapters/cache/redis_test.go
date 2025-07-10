package cache

import (
	"context"
	"testing"
	"time"

	"crypto/internal/core/domain"

	"github.com/redis/go-redis/v9"
)

func TestRedisAdapter_SetAndGetPrice(t *testing.T) {
	// This is a unit test example - in real scenarios you'd use a test Redis instance
	// or mock the Redis client

	// Skip if Redis not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use different DB for tests
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	// Clean up after test
	defer client.FlushDB(ctx)
	defer client.Close()

	adapter := NewRedisAdapter(client)

	// Test data
	testData := domain.MarketData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now().Unix(),
		Exchange:  "test-exchange",
	}

	// Test SetPrice
	err := adapter.SetPrice(ctx, "test-key", testData)
	if err != nil {
		t.Fatalf("SetPrice failed: %v", err)
	}

	// Test GetLatestPriceByExchange
	result, err := adapter.GetLatestPriceByExchange(ctx, "BTCUSDT", "test-exchange")
	if err != nil {
		t.Fatalf("GetLatestPriceByExchange failed: %v", err)
	}

	if result.Price != testData.Price {
		t.Errorf("Expected price %f, got %f", testData.Price, result.Price)
	}
}
