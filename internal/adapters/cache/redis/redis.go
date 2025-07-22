package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"marketflow/internal/application/ports"
	"marketflow/internal/config"
	"marketflow/internal/domain/models"
)

// Adapter implements the CachePort interface for Redis
type Adapter struct {
	client *redis.Client
}

// New creates a new Redis adapter
func New(cfg config.CacheConfig) (ports.CachePort, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Adapter{
		client: client,
	}, nil
}

// SetLatestPrice sets the latest price for a symbol from an exchange
func (a *Adapter) SetLatestPrice(ctx context.Context, update models.PriceUpdate) error {
	// Store latest price
	key := fmt.Sprintf("latest:%s:%s", update.Exchange, update.Symbol)

	price := models.LatestPrice{
		Symbol:    update.Symbol,
		Exchange:  update.Exchange,
		Price:     update.Price,
		Timestamp: update.ReceivedAt,
	}

	data, err := json.Marshal(price)
	if err != nil {
		return err
	}

	// Set with TTL
	if err := a.client.Set(ctx, key, data, 2*time.Minute).Err(); err != nil {
		return err
	}

	// Also store in history for aggregation
	historyKey := fmt.Sprintf("history:%s:%s", update.Exchange, update.Symbol)
	updateData, err := json.Marshal(update)
	if err != nil {
		return err
	}

	// Use sorted set with timestamp as score for easy range queries
	score := float64(update.ReceivedAt.UnixMilli())
	if err := a.client.ZAdd(ctx, historyKey, redis.Z{
		Score:  score,
		Member: updateData,
	}).Err(); err != nil {
		return err
	}

	// Remove old entries (older than 2 minutes)
	cutoff := time.Now().Add(-2 * time.Minute).UnixMilli()
	a.client.ZRemRangeByScore(ctx, historyKey, "0", fmt.Sprintf("%d", cutoff))

	return nil
}

// GetLatestPrice gets the latest price for a symbol from an exchange
func (a *Adapter) GetLatestPrice(ctx context.Context, symbol, exchange string) (*models.LatestPrice, error) {
	key := fmt.Sprintf("latest:%s:%s", exchange, symbol)

	data, err := a.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var price models.LatestPrice
	if err := json.Unmarshal([]byte(data), &price); err != nil {
		return nil, err
	}

	return &price, nil
}

// GetLatestPrices gets latest prices for a symbol from all exchanges
func (a *Adapter) GetLatestPrices(ctx context.Context, symbol string) ([]*models.LatestPrice, error) {
	pattern := fmt.Sprintf("latest:*:%s", symbol)

	keys, err := a.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []*models.LatestPrice{}, nil
	}

	values, err := a.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var prices []*models.LatestPrice
	for _, value := range values {
		if value == nil {
			continue
		}

		var price models.LatestPrice
		if err := json.Unmarshal([]byte(value.(string)), &price); err != nil {
			continue
		}

		prices = append(prices, &price)
	}

	return prices, nil
}

// GetPriceHistory gets price history for aggregation (last minute)
func (a *Adapter) GetPriceHistory(ctx context.Context, symbol, exchange string, duration time.Duration) ([]models.PriceUpdate, error) {
	key := fmt.Sprintf("history:%s:%s", exchange, symbol)

	now := time.Now()
	start := now.Add(-duration)

	// Use sorted set to get price history within time range
	values, err := a.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start.UnixMilli()),
		Max: fmt.Sprintf("%d", now.UnixMilli()),
	}).Result()
	if err != nil {
		return nil, err
	}

	var updates []models.PriceUpdate
	for _, value := range values {
		var update models.PriceUpdate
		if err := json.Unmarshal([]byte(value), &update); err != nil {
			continue
		}
		updates = append(updates, update)
	}

	return updates, nil
}

// CleanupOldData removes old price data from cache
func (a *Adapter) CleanupOldData(ctx context.Context, maxAge time.Duration) error {
	// Clean up latest prices
	pattern := "latest:*"
	keys, err := a.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	for _, key := range keys {
		ttl, err := a.client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		if ttl < 0 || ttl > maxAge {
			a.client.Del(ctx, key)
		}
	}

	// Clean up history data
	historyPattern := "history:*"
	historyKeys, err := a.client.Keys(ctx, historyPattern).Result()
	if err != nil {
		return err
	}

	for _, key := range historyKeys {
		// Remove old entries from sorted sets
		a.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff.UnixMilli()))
	}

	return nil
}

// Close closes the cache connection
func (a *Adapter) Close() error {
	return a.client.Close()
}
