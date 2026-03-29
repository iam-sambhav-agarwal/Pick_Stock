# Stock Analysis API (Go)

This is the Go implementation of the Stock Analysis API, converted from the original Python FastAPI version.

## Features

- Real-time stock data fetching from Yahoo Finance
- Comprehensive stock analysis with buy/sell/hold recommendations
- Support and resistance level calculations
- Historical analysis storage in MongoDB
- RESTful API with JSON responses
- Concurrent data fetching for better performance
- Static file serving for React frontend

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [MongoDB Driver](https://go.mongodb.org/mongo-driver) - Official MongoDB driver
- [CORS](https://github.com/gin-contrib/cors) - Cross-Origin Resource Sharing
- [godotenv](https://github.com/joho/godotenv) - Environment variable loading

## Getting Started

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Set up environment variables:**
   Create a `.env` file in the backend-go directory:
   ```env
   MONGO_URL=mongodb://localhost:27017
   DB_NAME=stock_analysis
   PORT=8001
   CORS_ORIGINS=*
   ```

3. **Run the server:**
   ```bash
   go run main.go
   ```

## API Endpoints

### Base URL: `http://localhost:8001`

#### API Routes (prefixed with `/api`)

- `GET /api/` - API information
- `GET /api/stocks/popular` - Get list of popular stocks
- `POST /api/analyze` - Analyze a stock
- `GET /api/history/{symbol}` - Get analysis history for a stock

#### Health Check
- `GET /health` - Server health status

### Request/Response Examples

#### Analyze Stock
```bash
curl -X POST http://localhost:8001/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"symbol": "TCS", "exchange": "NSE"}'
```

#### Get Popular Stocks
```bash
curl http://localhost:8001/api/stocks/popular
```

#### Get Analysis History
```bash
curl "http://localhost:8001/api/history/TCS?limit=5"
```

## Project Structure

```
backend-go/
├── main.go                 # Main server setup
├── go.mod                  # Go module definition
├── models/
│   └── models.go          # Data structures and types
├── services/
│   ├── yahoo_finance_service.go  # Stock data fetching
│   └── analysis_service.go       # Stock analysis logic
└── handlers/
    └── stock_handler.go    # HTTP request handlers
```

## Key Differences from Python Version

1. **Concurrency**: Uses goroutines for concurrent data fetching instead of asyncio
2. **Type Safety**: Strong typing with Go structs instead of Pydantic models
3. **Error Handling**: Explicit error handling instead of exceptions
4. **Performance**: Generally faster execution and lower memory usage
5. **Deployment**: Single binary deployment instead of Python dependencies

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGO_URL` | `mongodb://localhost:27017` | MongoDB connection string |
| `DB_NAME` | `stock_analysis` | MongoDB database name |
| `PORT` | `8001` | Server port |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |

## Building for Production

```bash
# Build for current platform
go build -o stock-analysis-api

# Build for Linux (for deployment)
GOOS=linux GOARCH=amd64 go build -o stock-analysis-api-linux

# Run the binary
./stock-analysis-api
```

## Docker Support

You can containerize the Go application:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8001
CMD ["./main"]
```

## Development

The Go version maintains the same API contract as the Python version, making it a drop-in replacement. All endpoints return identical JSON responses for seamless frontend compatibility.
