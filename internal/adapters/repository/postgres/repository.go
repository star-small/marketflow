// Create file: internal/adapters/repository/postgres/repository.go

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

type PriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) port.PriceRepository {
	return &PriceRepository{
		db: db,
	}
}

// StorePriceAggregation stores aggregated price data for a minute
func (r *PriceRepository) StorePriceAggregation(ctx context.Context, agg domain.PriceAggregation) error {
	query := `
		INSERT INTO prices (pair_name, exchange, timestamp, average_price, min_price, max_price)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (pair_name, exchange, timestamp) 
		DO UPDATE SET 
			average_price = EXCLUDED.average_price,
			min_price = EXCLUDED.min_price,
			max_price = EXCLUDED.max_price`

	_, err := r.db.ExecContext(ctx, query,
		agg.PairName,
		agg.Exchange,
		agg.Timestamp,
		agg.AveragePrice,
		agg.MinPrice,
		agg.MaxPrice,
	)

	if err != nil {
		return fmt.Errorf("failed to store price aggregation: %w", err)
	}

	slog.Debug("Stored price aggregation",
		"pair", agg.PairName,
		"exchange", agg.Exchange,
		"timestamp", agg.Timestamp,
		"avg_price", agg.AveragePrice,
	)

	return nil
}

// GetAggregatedPricesInRange retrieves aggregated prices within a time range
func (r *PriceRepository) GetAggregatedPricesInRange(ctx context.Context, symbol, exchange string, from, to time.Time) ([]domain.PriceAggregation, error) {
	var query string
	var args []interface{}

	if exchange != "" {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3 AND timestamp <= $4
			ORDER BY timestamp DESC`
		args = []interface{}{symbol, exchange, from, to}
	} else {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND timestamp >= $2 AND timestamp <= $3
			ORDER BY timestamp DESC`
		args = []interface{}{symbol, from, to}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated prices: %w", err)
	}
	defer rows.Close()

	var results []domain.PriceAggregation
	for rows.Next() {
		var agg domain.PriceAggregation
		err := rows.Scan(
			&agg.PairName,
			&agg.Exchange,
			&agg.Timestamp,
			&agg.AveragePrice,
			&agg.MinPrice,
			&agg.MaxPrice,
		)
		if err != nil {
			slog.Error("Failed to scan price aggregation row", "error", err)
			continue
		}
		results = append(results, agg)
	}

	return results, nil
}

// GetHighestPriceInRange gets the highest price in a time range
func (r *PriceRepository) GetHighestPriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceAggregation, error) {
	var query string
	var args []interface{}

	if exchange != "" {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3 AND timestamp <= $4
			ORDER BY max_price DESC
			LIMIT 1`
		args = []interface{}{symbol, exchange, from, to}
	} else {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND timestamp >= $2 AND timestamp <= $3
			ORDER BY max_price DESC
			LIMIT 1`
		args = []interface{}{symbol, from, to}
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var agg domain.PriceAggregation
	err := row.Scan(
		&agg.PairName,
		&agg.Exchange,
		&agg.Timestamp,
		&agg.AveragePrice,
		&agg.MinPrice,
		&agg.MaxPrice,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get highest price: %w", err)
	}

	return &agg, nil
}

// GetLowestPriceInRange gets the lowest price in a time range
func (r *PriceRepository) GetLowestPriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (*domain.PriceAggregation, error) {
	var query string
	var args []interface{}

	if exchange != "" {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3 AND timestamp <= $4
			ORDER BY min_price ASC
			LIMIT 1`
		args = []interface{}{symbol, exchange, from, to}
	} else {
		query = `
			SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
			FROM prices 
			WHERE pair_name = $1 AND timestamp >= $2 AND timestamp <= $3
			ORDER BY min_price ASC
			LIMIT 1`
		args = []interface{}{symbol, from, to}
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var agg domain.PriceAggregation
	err := row.Scan(
		&agg.PairName,
		&agg.Exchange,
		&agg.Timestamp,
		&agg.AveragePrice,
		&agg.MinPrice,
		&agg.MaxPrice,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lowest price: %w", err)
	}

	return &agg, nil
}

// GetAveragePriceInRange calculates the average of average prices in a time range
func (r *PriceRepository) GetAveragePriceInRange(ctx context.Context, symbol, exchange string, from, to time.Time) (float64, error) {
	var query string
	var args []interface{}

	if exchange != "" {
		query = `
			SELECT AVG(average_price)
			FROM prices 
			WHERE pair_name = $1 AND exchange = $2 AND timestamp >= $3 AND timestamp <= $4`
		args = []interface{}{symbol, exchange, from, to}
	} else {
		query = `
			SELECT AVG(average_price)
			FROM prices 
			WHERE pair_name = $1 AND timestamp >= $2 AND timestamp <= $3`
		args = []interface{}{symbol, from, to}
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var avgPrice sql.NullFloat64
	err := row.Scan(&avgPrice)
	if err != nil {
		return 0, fmt.Errorf("failed to get average price: %w", err)
	}

	if !avgPrice.Valid {
		return 0, fmt.Errorf("no price data found")
	}

	return avgPrice.Float64, nil
}

// CleanupOldData removes data older than specified duration
func (r *PriceRepository) CleanupOldData(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)

	query := `DELETE FROM prices WHERE timestamp < $1`
	result, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	slog.Info("Cleaned up old price data", "rows_deleted", rowsAffected, "cutoff_time", cutoffTime)

	return nil
}
