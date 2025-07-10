# MarketFlow - Cryptocurrency Market Data Processing

MarketFlow is a real-time cryptocurrency market data processing application built with Go, following hexagonal architecture principles. It processes market data from multiple exchanges, provides real-time caching with Redis, stores aggregated data in PostgreSQL, and offers a comprehensive REST API.

## Features

- **Real-time Data Processing**: Processes market data from multiple cryptocurrency exchanges
- **Dual Operation Modes**: Live mode (real exchanges) and Test mode (synthetic data)
- **High-Performance Concurrency**: Implements Fan-in/Fan-out patterns with worker pools
- **Caching Layer**: Redis caching for latest prices with automatic cleanup
- **Data Persistence**: PostgreSQL storage for aggregated minute-by-minute data
- **REST API**: Comprehensive endpoints for price data and statistics
- **Health Monitoring**: System health checks and detailed status reports
- **Graceful Shutdown**: Proper resource cleanup on application termination

## Architecture

The application follows hexagonal (ports and adapters) architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    REST API Layer                           │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              HTTP Handlers                              │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │     Services (Price, Exchange, Health, Mode)           │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │        Business Logic & Domain Models                  │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                       │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │   PostgreSQL │ │    Redis     │ │   Exchange Adapters  │ │
│  │  Repository  │ │    Cache     │ │  (Live/Test modes)   │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Supported Cryptocurrencies

- BTCUSDT (Bitcoin/Tether)
- ETHUSDT (Ethereum/Tether)
- DOGEUSDT (Dogecoin/Tether)
- TONUSDT (Toncoin/Tether)
- SOLUSDT (Solana/Tether)

## Prerequisites

- Go 1.22.6 or later
- Docker and Docker Compose
- PostgreSQL 15
- Redis 7
- Make (optional, for using Makefile commands)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd marketflow

# Setup development environment (installs tools, starts containers, loads exchanges)
make dev-setup

# Or setup manually:
make docker-up
make load-exchanges
make start-exchanges
```

### 2. Build and Run

```bash
# Build the application
make build

# Run with default settings (port 8080)
make run

# Run with custom port
./build/marketflow --port 9090

# Show help
./build/marketflow --help
```

### 3. Test the API

```bash
# Check health
curl http://localhost:8080/health

# Get latest BTC price
curl http://localhost:8080/prices/latest/BTCUSDT

# Switch to test mode
curl -X POST http://localhost:8080/mode/test

# Get current mode
curl http://localhost:8080/mode/current
```

## Configuration

The application uses a JSON configuration file at `config/config.json`:

```json
{
  "app": {
    "port": 8080
  },
  "repository": {
    "db_host": "localhost",
    "db_port": 5432,
    "db_name": "market",
    "db_username": "postgres",
    "db_password": "postgres",
    "db_ssl_mode": "disable",
    "max_conn": 10,
    "max_idle_conn": 10
  },
  "cache": {
    "redis_host": "localhost",
    "redis_port": 6379,
    "redis_password": "",
    "redis_db": 0,
    "pool_size": 10,
    "min_idle_conns": 5
  },
  "exchanges": {
    "live_exchanges": [
      {"name": "exchange1", "host": "localhost", "port": 40101},
      {"name": "exchange2", "host": "localhost", "port": 40102},
      {"name": "exchange3", "host": "localhost", "port": 40103}
    ],
    "test_mode": {
      "update_interval_ms": 1000,
      "symbols": ["BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"]
    }
  }
}
```

### Environment Variables

You can override configuration with environment variables:

```bash
export PORT=9090
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=market
export REDIS_HOST=localhost
export REDIS_PORT=6379
export EXCHANGE1_HOST=localhost
export EXCHANGE1_PORT=40101
# ... etc
```

## API Endpoints

### Market Data API

#### Latest Prices
- `GET /prices/latest/{symbol}` - Get latest price across all exchanges
- `GET /prices/latest/{exchange}/{symbol}` - Get latest price from specific exchange

#### Price Statistics
- `GET /prices/highest/{symbol}[?period=duration]` - Get highest price in period
- `GET /prices/highest/{exchange}/{symbol}[?period=duration]` - Get highest price from exchange
- `GET /prices/lowest/{symbol}[?period=duration]` - Get lowest price in period
- `GET /prices/lowest/{exchange}/{symbol}[?period=duration]` - Get lowest price from exchange
- `GET /prices/average/{symbol}[?period=duration]` - Get average price in period
- `GET /prices/average/{exchange}/{symbol}[?period=duration]` - Get average price from exchange

#### Period Examples
- `?period=1h` - Last hour
- `?period=30m` - Last 30 minutes
- `?period=5s` - Last 5 seconds
- `?period=3600` - Last 3600 seconds (1 hour)

### Data Mode API
- `POST /mode/test` - Switch to test mode (synthetic data)
- `POST /mode/live` - Switch to live mode (real exchanges)
- `GET /mode/current` - Get current mode and status

### System Health API
- `GET /health` - Basic system health check
- `GET /health/detailed` - Detailed health information

## Data Flow

### Live Mode
1. Connects to 3 exchange programs via TCP (ports 40101, 40102, 40103)
2. Receives real-time price updates
3. Processes data through worker pools (15 workers total, 5 per exchange)
4. Stores latest prices in Redis cache
5. Aggregates data every minute and stores in PostgreSQL

### Test Mode
1. Generates synthetic price data using configurable parameters
2. Simulates realistic market movements with trends and volatility
3. Processes data through the same pipeline as live mode

### Concurrency Patterns

```
Exchange 1 ──┐
             ├─► Fan-In ──► Distributor ──► Worker Pool (15 workers) ──► Fan-Out ──► Results
Exchange 2 ──┤                                     │
Exchange 3 ──┘                                     ├─► Redis Cache
                                                    └─► Aggregator ──► PostgreSQL
```

## Database Schema

The application stores aggregated data in PostgreSQL:

```sql
CREATE TABLE prices (
    pair_name TEXT NOT NULL,
    exchange TEXT NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    average_price DOUBLE PRECISION NOT NULL,
    min_price DOUBLE PRECISION NOT NULL,
    max_price DOUBLE PRECISION NOT NULL,
    UNIQUE(pair_name, exchange, timestamp)
);
```

## Development

### Running Tests

```bash
# Run all tests with coverage
make test

# Run specific package tests
go test -v ./internal/core/service/prices/

# Run with race detection
go test -race ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Run all checks
make check
```

### Docker Support

```bash
# Start PostgreSQL and Redis
make docker-up

# Stop containers
make docker-down

# Reset database
make db-reset
```

## Monitoring and Logging

The application uses structured logging with `log/slog`:

- **Info**: Normal operation events
- **Warn**: Recoverable issues (cache misses, connection retries)
- **Error**: Serious issues requiring attention
- **Debug**: Detailed information for troubleshooting

### Health Checks

The `/health` endpoint provides system status:

```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T12:00:00Z",
  "checks": {
    "database": {"status": "healthy"},
    "cache": {"status": "healthy"},
    "exchange_service": {"status": "healthy"},
    "exchanges": {
      "status": "healthy",
      "details": {
        "total_exchanges": 3,
        "healthy_exchanges": 3
      }
    }
  }
}
```

## Deployment

### Production Build

```bash
make build-prod
```

### Environment Configuration

For production, set appropriate environment variables:

```bash
export PORT=8080
export DB_HOST=production-db-host
export DB_PASSWORD=secure-password
export REDIS_HOST=production-redis-host
# ... etc
```

## Troubleshooting

### Common Issues

1. **Connection refused to exchanges**
    - Ensure exchange containers are running: `make start-exchanges`
    - Check ports 40101, 40102, 40103 are available

2. **Database connection errors**
    - Verify PostgreSQL is running: `make docker-up`
    - Check database credentials in config

3. **Redis connection errors**
    - Verify Redis is running
    - Application will continue without Redis (degraded mode)

4. **High memory usage**
    - Check aggregation buffer sizes
    - Monitor Redis memory usage
    - Verify cleanup routines are running


## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes following the existing code style
4. Add tests for new functionality
5. Run `make check` to ensure code quality
6. Submit a pull request
