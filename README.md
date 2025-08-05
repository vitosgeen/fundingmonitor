# Funding Monitor - CEX

A real-time funding rate monitoring application for centralized exchanges (CEX) built in Go. This application aggregates funding rates from multiple exchanges and provides a modern web interface for monitoring and analysis.

## Features

- **Multi-Exchange Support**: Currently supports Binance, Bybit, OKX, MEXC, BitGet, Gate.io, Deribit, XT, KuCoin, and more
- **Real-time Monitoring**: Automatic refresh every 30 seconds
- **Historical Data Logging**: Automatic logging of funding rates to individual files per trading pair
- **Modern Web Interface**: Built with Tailwind CSS and responsive design
- **Funding Spread Analysis**: New dedicated page for analyzing funding rate spreads across exchanges
- **Advanced Filtering**: Filter by exchange, symbol, and funding rate direction
- **Sorting Options**: Sort by funding rate, symbol, exchange, or next funding time
- **RESTful API**: JSON API endpoints for programmatic access
- **Health Monitoring**: Built-in health checks for all exchanges

## Supported Exchanges

| Exchange | Status | API Endpoint |
|----------|--------|--------------|
| Binance  | ✅     | `/fapi/v1/premiumIndex` |
| Bybit    | ✅     | `/v5/market/funding/history` |
| OKX      | ✅     | `/api/v5/public/funding-rate` |

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd fundingmonitor
```

2. Install dependencies:
```bash
go mod tidy
```

3. Configure the application:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your API keys (optional)
```

4. Run the application:
```bash
go run .
```

5. Open your browser and navigate to `http://localhost:8080`

## Web Interface

The application provides two main web pages:

### Overview Page (`/`)
- Real-time funding rate monitoring
- Filtering and sorting capabilities
- Historical data analysis
- Alert settings

### Funding Spread Analysis (`/funding.html`)
- Cross-exchange funding rate comparison
- Maximum spread calculation per coin
- Visual indicators for positive/negative spreads
- Exchange-specific funding rate display
- Advanced filtering and sorting options

## Configuration

The application uses a YAML configuration file (`config.yaml`):

```yaml
port: "8080"
logging_interval: 1  # minutes
log_directory: "funding_logs"

exchanges:
  binance:
    enabled: true
    base_url: "https://fapi.binance.com"
    api_key: ""      # Optional
    api_secret: ""   # Optional
    
  bybit:
    enabled: true
    base_url: "https://api.bybit.com"
    api_key: ""      # Optional
    api_secret: ""   # Optional
    
  okx:
    enabled: false   # Disabled due to API complexity
    base_url: "https://www.okx.com"
    api_key: ""      # Optional
    api_secret: ""   # Optional
```

### API Keys (Optional)

While the application works without API keys for public endpoints, you can add your API keys for:
- Higher rate limits
- Access to additional data
- Better reliability

## API Endpoints

### Get All Funding Rates
```
GET /api/funding
```

Response:
```json
{
  "timestamp": 1640995200,
  "rates": [
    {
      "symbol": "BTCUSDT",
      "exchange": "binance",
      "funding_rate": 0.0001,
      "next_funding_time": "2024-01-01T08:00:00Z",
      "timestamp": "2024-01-01T07:30:00Z",
      "mark_price": 45000.50,
      "index_price": 44998.25,
      "last_funding_rate": 0.0002
    }
  ]
}
```

### Get Exchange-Specific Funding Rates
```
GET /api/funding/{exchange}
```

Response:
```json
{
  "exchange": "binance",
  "timestamp": 1640995200,
  "rates": [...],
  "status": "healthy"
}
```

### Health Check
```
GET /api/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "exchanges": 3
}
```

### Logging Endpoints

#### Get All Log Files
```
GET /api/logs
```
Returns a list of all available log files with metadata.

#### Get Symbol Logs
```
GET /api/logs/{symbol}
GET /api/logs/{symbol}?date=28-07-2025
```
Returns historical funding rate data for a specific symbol. Optionally specify a date parameter in DD-MM-YYYY format.

Example response:
```json
{
  "timestamp": "2025-07-28T21:06:53.528617398+03:00",
  "symbol": "BTCUSDT",
  "rates": [
    {
      "symbol": "BTCUSDT",
      "exchange": "binance",
      "funding_rate": 0.0001,
      "next_funding_time": "2025-07-29T03:00:00+03:00",
      "timestamp": "2025-07-28T21:06:53.506090985+03:00",
      "mark_price": 117493.8,
      "index_price": 117513.18136364,
      "last_funding_rate": 0.0001
    }
  ]
}
```

## Logging System

The application automatically logs funding rates to individual files for each trading pair. This provides historical data tracking and analysis capabilities.

### Log File Structure
- **Location**: `funding_logs/` directory
- **Organization**: `funding_logs/{SYMBOL}/{DATE}.log` (e.g., `funding_logs/BTCUSDT/28-07-2025.log`)
- **Format**: JSON with timestamp, symbol, and rates from all exchanges
- **Frequency**: Every minute (configurable)

### Configuration
```yaml
logging_interval: 1  # minutes
log_directory: "funding_logs"
```

### Log File Example
```json
{
  "timestamp": "2025-07-28T21:06:53.528617398+03:00",
  "symbol": "BTCUSDT",
  "rates": [
    {
      "symbol": "BTCUSDT",
      "exchange": "binance",
      "funding_rate": 0.0001,
      "next_funding_time": "2025-07-29T03:00:00+03:00",
      "timestamp": "2025-07-28T21:06:53.506090985+03:00",
      "mark_price": 117493.8,
      "index_price": 117513.18136364,
      "last_funding_rate": 0.0001
    },
    {
      "symbol": "BTCUSDT",
      "exchange": "bybit",
      "funding_rate": 0.0001,
      "next_funding_time": "2025-07-29T03:00:00+03:00",
      "timestamp": "2025-07-28T21:06:53.175859664+03:00",
      "mark_price": 117500.1,
      "index_price": 117513.36
    }
  ]
}
```

### Log Management

The application includes a cleanup script to manage log files and disk space:

```bash
# Show log statistics
./cleanup_logs.sh -s

# List all log files with sizes
./cleanup_logs.sh -l

# Clean up logs older than 7 days (default)
./cleanup_logs.sh

# Clean up logs older than 30 days
./cleanup_logs.sh -d 30
```

## Web Interface Features

### Dashboard
- Real-time statistics showing total pairs, positive/negative funding rates
- Connection status indicator
- Last update timestamp

### Filtering Options
- **Exchange**: Filter by specific exchange (Binance, Bybit, OKX)
- **Symbol**: Search for specific trading pairs (e.g., BTCUSDT)
- **Funding Rate**: Filter by positive, negative, or neutral funding rates
- **Sort By**: Sort by funding rate magnitude, symbol, exchange, or next funding time

### Data Display
- Color-coded funding rates (green for positive, red for negative)
- Responsive table with scrollable content
- Hover effects for better user experience

## Development

### Project Structure
```
fundingmonitor/
├── main.go              # Application entry point
├── types.go             # Core types and interfaces
├── binance.go           # Binance exchange implementation
├── bybit.go             # Bybit exchange implementation
├── okx.go               # OKX exchange implementation
├── static/              # Web interface
│   └── index.html      # Main web page
├── funding_logs/        # Historical data logs (auto-generated)
│   ├── BTCUSDT/
│   │   └── 28-07-2025.log
│   ├── ETHUSDT/
│   │   └── 28-07-2025.log
│   └── ... (one directory per trading pair)
├── config.yaml          # Configuration file
├── go.mod               # Go module file
├── .gitignore           # Git ignore rules
├── Dockerfile           # Docker configuration
├── docker-compose.yml   # Docker Compose setup
├── Makefile             # Build and deployment scripts
├── run.sh               # Quick start script
├── cleanup_logs.sh      # Log cleanup and management script
└── README.md            # This file
```

### Adding New Exchanges

To add support for a new exchange:

1. Create a new file in the `exchanges/` directory (e.g., `exchanges/kraken.go`)
2. Implement the `Exchange` interface:
   ```go
   type Exchange interface {
       GetFundingRates() ([]FundingRate, error)
       GetName() string
       IsHealthy() bool
   }
   ```
3. Add the exchange to the configuration file
4. Update the exchange initialization in `main.go`

### Building

Build the application:
```bash
go build -o fundingmonitor .
```

Run the binary:
```bash
./fundingmonitor
```

## Docker Support

Create a Dockerfile:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o fundingmonitor .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/fundingmonitor .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./fundingmonitor"]
```

Build and run with Docker:
```bash
docker build -t fundingmonitor .
docker run -p 8080:8080 fundingmonitor
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This application is for educational and monitoring purposes only. Always verify data from official exchange sources before making trading decisions. The authors are not responsible for any financial losses incurred through the use of this software. 