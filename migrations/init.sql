CREATE TABLE prices (
                        pair_name TEXT NOT NULL,
                        exchange TEXT NOT NULL,
                        timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
                        average_price DOUBLE PRECISION NOT NULL,
                        min_price DOUBLE PRECISION NOT NULL,
                        max_price DOUBLE PRECISION NOT NULL,

    -- Add unique constraint to prevent duplicate entries
                        UNIQUE(pair_name, exchange, timestamp)
);

-- Create indexes for better query performance
CREATE INDEX idx_prices_pair_timestamp ON prices(pair_name, timestamp DESC);
CREATE INDEX idx_prices_exchange_timestamp ON prices(exchange, timestamp DESC);
CREATE INDEX idx_prices_pair_exchange_timestamp ON prices(pair_name, exchange, timestamp DESC);

-- INSERT INTO prices (pair_name, exchange, timestamp, average_price, min_price, max_price)
-- VALUES
--     ('BTCUSDT', 'exchange1', NOW() - INTERVAL '1 minute', 96000.0, 95900.0, 96100.0),
--     ('ETHUSDT', 'exchange1', NOW() - INTERVAL '1 minute', 3300.0, 3295.0, 3305.0);