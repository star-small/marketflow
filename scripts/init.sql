-- Create the marketflow user if it doesn't exist
DO $$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'marketflow') THEN
      CREATE USER marketflow WITH PASSWORD 'password';
   END IF;
END $$;

-- Grant necessary privileges
ALTER USER marketflow CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE marketflow TO marketflow;

-- Create the market_data table
CREATE TABLE IF NOT EXISTS market_data (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    average_price DECIMAL(20, 8) NOT NULL,
    min_price DECIMAL(20, 8) NOT NULL,
    max_price DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_market_data_pair_exchange ON market_data(pair_name, exchange);
CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_market_data_created_at ON market_data(created_at);

-- Grant table permissions
GRANT ALL PRIVILEGES ON TABLE market_data TO marketflow;
GRANT ALL PRIVILEGES ON SEQUENCE market_data_id_seq TO marketflow;
