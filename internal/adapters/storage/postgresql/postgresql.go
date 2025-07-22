package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"marketflow/internal/application/ports"
	"marketflow/internal/config"
	"marketflow/internal/domain/models"
)

// Adapter implements the StoragePort interface for PostgreSQL
type Adapter struct {
	db *sql.DB
}

// New creates a new PostgreSQL adapter
func New(cfg config.DatabaseConfig) (ports.StoragePort, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Adapter{
		db: db,
	}, nil
}

// SaveAggregatedData saves aggregated market data
func (a *Adapter) SaveAggregatedData(ctx context.Context, data []models.AggregatedData) error {
	if len(data) == 0 {
		return nil
	}

	query := `INSERT INTO market_data (pair_name, exchange, timestamp, average_price, min_price, max_price)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range data {
		_, err := stmt.ExecContext(ctx, item.PairName, item.Exchange, item.Timestamp,
			item.AveragePrice, item.MinPrice, item.MaxPrice)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAggregatedData retrieves aggregated data within a time range
func (a *Adapter) GetAggregatedData(ctx context.Context, symbol, exchange string, from, to time.Time) ([]models.AggregatedData, error) {
	query := `SELECT id, pair_name, exchange, timestamp, average_price, min_price, max_price
			  FROM market_data
			  WHERE pair_name = $1 AND exchange = $2 AND timestamp BETWEEN $3 AND $4
			  ORDER BY timestamp DESC`

	rows, err := a.db.QueryContext(ctx, query, symbol, exchange, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []models.AggregatedData
	for rows.Next() {
		var item models.AggregatedData
		err := rows.Scan(&item.ID, &item.PairName, &item.Exchange, &item.Timestamp,
			&item.AveragePrice, &item.MinPrice, &item.MaxPrice)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, rows.Err()
}

// GetHighestPrice returns the highest price within a period
func (a *Adapter) GetHighestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	query := `SELECT id, pair_name, exchange, timestamp, average_price, min_price, max_price
			  FROM market_data
			  WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3
			  ORDER BY max_price DESC
			  LIMIT 1`

	from := time.Now().Add(-period)

	var item models.AggregatedData
	err := a.db.QueryRowContext(ctx, query, symbol, exchange, from).Scan(
		&item.ID, &item.PairName, &item.Exchange, &item.Timestamp,
		&item.AveragePrice, &item.MinPrice, &item.MaxPrice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

// GetLowestPrice returns the lowest price within a period
func (a *Adapter) GetLowestPrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	query := `SELECT id, pair_name, exchange, timestamp, average_price, min_price, max_price
			  FROM market_data
			  WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3
			  ORDER BY min_price ASC
			  LIMIT 1`

	from := time.Now().Add(-period)

	var item models.AggregatedData
	err := a.db.QueryRowContext(ctx, query, symbol, exchange, from).Scan(
		&item.ID, &item.PairName, &item.Exchange, &item.Timestamp,
		&item.AveragePrice, &item.MinPrice, &item.MaxPrice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

// GetAveragePrice returns the average price within a period
func (a *Adapter) GetAveragePrice(ctx context.Context, symbol, exchange string, period time.Duration) (*models.AggregatedData, error) {
	query := `SELECT AVG(average_price) as avg_price, MIN(min_price) as min_price, MAX(max_price) as max_price
			  FROM market_data
			  WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3`

	from := time.Now().Add(-period)

	var avgPrice, minPrice, maxPrice sql.NullFloat64
	err := a.db.QueryRowContext(ctx, query, symbol, exchange, from).Scan(&avgPrice, &minPrice, &maxPrice)

	if err != nil {
		return nil, err
	}

	if !avgPrice.Valid {
		return nil, nil
	}

	return &models.AggregatedData{
		PairName:     symbol,
		Exchange:     exchange,
		Timestamp:    time.Now(),
		AveragePrice: avgPrice.Float64,
		MinPrice:     minPrice.Float64,
		MaxPrice:     maxPrice.Float64,
	}, nil
}

// Close closes the storage connection
func (a *Adapter) Close() error {
	return a.db.Close()
}
