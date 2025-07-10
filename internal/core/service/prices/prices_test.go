package prices

import (
	"context"
	"fmt"
	"testing"
	"time"

	"crypto/internal/core/domain"
)

// MockCache implements the Cache interface for testing
type MockCache struct {
	data map[string]domain.MarketData
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]domain.MarketData),
	}
}

func (m *MockCache) SetPrice(ctx context.Context, key string, data domain.MarketData) error {
	m.data[key] = data
	return nil
}

func (m *MockCache) GetLatestPrice(ctx context.Context, symbol string) (*domain.MarketData, error) {
	// Find the latest price for this symbol across all keys
	var latest *domain.MarketData
	var latestTime int64

	for _, data := range m.data {
		if data.Symbol == symbol && data.Timestamp > latestTime {
			latest = &data
			latestTime = data.Timestamp
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no data found")
	}

	return latest, nil
}

func (m *MockCache) GetLatestPriceByExchange(ctx context.Context, symbol, exchange string) (*domain.MarketData, error) {
	key := symbol + ":" + exchange
	if data, exists := m.data[key]; exists {
		return &data, nil
	}
	return nil, fmt.Errorf("no data found")
}

func (m *MockCache) GetPricesInRange(ctx context.Context, symbol string, from, to time.Time) ([]domain.MarketData, error) {
	var result []domain.MarketData
	for _, data := range m.data {
		if data.Symbol == symbol {
			dataTime := time.Unix(data.Timestamp, 0)
			if dataTime.After(from) && dataTime.Before(to) {
				result = append(result, data)
			}
		}
	}
	return result, nil
}

func (m *MockCache) GetPricesInRangeByExchange(ctx context.Context, symbol, exchange string, from, to time.Time) ([]domain.MarketData, error) {
	var result []domain.MarketData
	for _, data := range m.data {
		if data.Symbol == symbol && data.Exchange == exchange {
			dataTime := time.Unix(data.Timestamp, 0)
			if dataTime.After(from) && dataTime.Before(to) {
				result = append(result, data)
			}
		}
	}
	return result, nil
}

func (m *MockCache) CleanupOldData(ctx context.Context, olderThan time.Duration) error {
	return nil
}

func (m *MockCache) Ping(ctx context.Context) error {
	return nil
}

func TestPriceService_GetLatestPrice(t *testing.T) {
	cache := NewMockCache()
	service := NewPriceService(cache, nil)

	ctx := context.Background()

	// Add test data
	testData := domain.MarketData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now().Unix(),
		Exchange:  "test-exchange",
	}

	cache.SetPrice(ctx, "BTCUSDT:test-exchange", testData)

	// Test GetLatestPrice
	result, err := service.GetLatestPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetLatestPrice failed: %v", err)
	}

	if result.Price != testData.Price {
		t.Errorf("Expected price %f, got %f", testData.Price, result.Price)
	}
}
