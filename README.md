# Funding Monitor - CEX

A real-time funding rate monitoring application for centralized exchanges (CEX) built in Go. This application aggregates funding rates from multiple exchanges and provides a modern web interface for monitoring and analysis.

## Features

- **Multi-Exchange Support**: Currently supports Binance, Bybit, and OKX
- **Real-time Monitoring**: Automatic refresh every 30 seconds
- **Modern Web Interface**: Built with Tailwind CSS and responsive design
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

## Configuration

The application uses a YAML configuration file (`config.yaml`):

```yaml
port: "8080"

exchanges:
  binance:
    enabled: true
    base_url: "https://api.binance.com"
    api_key: ""      # Optional
    api_secret: ""   # Optional
    
  bybit:
    enabled: true
    base_url: "https://api.bybit.com"
    api_key: ""      # Optional
    api_secret: ""   # Optional
    
  okx:
    enabled: true
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
├── exchanges/           # Exchange implementations
│   ├── binance.go      # Binance exchange
│   ├── bybit.go        # Bybit exchange
│   └── okx.go          # OKX exchange
├── static/              # Web interface
│   └── index.html      # Main web page
├── config.yaml         # Configuration file
├── go.mod              # Go module file
└── README.md           # This file
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