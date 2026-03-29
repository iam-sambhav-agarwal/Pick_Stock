# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy .env file if exists
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 8001

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8001

# Run the binary
CMD ["./main"]
