# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY cmd ./cmd
COPY internal ./internal

# Build the binary statically (CGO_ENABLED=0 ensures it runs on pure alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Add CA certificates (often required for secure DB/MQ connections in prod)
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/worker .

# Run as non-root user for security
RUN addgroup -S workergroup && adduser -S workeruser -G workergroup
USER workeruser

# Start the worker
CMD ["./worker"]
