# Build stage for React frontend
FROM node:18-alpine AS frontend-build

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --only=production

COPY frontend/ ./
RUN npm run build

# Build stage for Go backend
FROM golang:1.21-alpine AS backend-build

WORKDIR /app/backend-go
COPY backend-go/go.mod backend-go/go.sum ./
RUN go mod download

COPY backend-go/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy built Go binary
COPY --from=backend-build /app/backend-go/main .

# Copy built React frontend
COPY --from=frontend-build /app/frontend/build ./frontend/build

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S go -u 1001
USER go

EXPOSE 8001

CMD ["./main"]
