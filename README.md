# MarketFlow - Real-Time Market Data Processing System

A high-performance real-time market data processing system built with Go, implementing hexagonal architecture and advanced concurrency patterns.

## Features

- **Hexagonal Architecture**: Clean separation of concerns with domain, application, and adapter layers
- **Concurrency Patterns**: Fan-in, fan-out, worker pools for efficient data processing
- **Dual Data Modes**: Live exchange data and test data generation
- **Real-time Processing**: Handles high-volume market data streams
- **Storage & Caching**: PostgreSQL for persistence, Redis for caching
- **REST API**: Comprehensive endpoints for price queries and statistics
- **Graceful Shutdown**: Proper resource cleanup on termination

## Quick Start

1. **Start development environment:**
   ```bash
   make docker-up
   ```

2. **Build and run:**
   ```bash
   make build
   ./marketflow --port 8080
   ```

## Architecture

```
├── Domain Layer (business logic)
├── Application Layer (use cases)
└── Adapters Layer
    ├── Web (HTTP handlers)
    ├── Storage (PostgreSQL)
    ├── Cache (Redis)
    └── Exchange (live/test data)
```

## API Endpoints

- `GET /prices/latest/{symbol}` - Latest price for symbol
- `GET /prices/highest/{symbol}?period=1m` - Highest price in period
- `GET /prices/lowest/{symbol}?period=1m` - Lowest price in period
- `GET /prices/average/{symbol}?period=1m` - Average price in period
- `POST /mode/live` - Switch to live data mode
- `POST /mode/test` - Switch to test data mode
- `GET /health` - System health status

## Configuration

Edit `configs/config.json` to configure database, cache, and exchange connections.

## Development

- `make build` - Build the application
- `make test` - Run tests
- `make fmt` - Format code with gofumpt
- `make docker-up` - Start PostgreSQL and Redis
