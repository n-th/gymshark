# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV APP_ENV=docker
ENV GIN_MODE=release

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary and config from builder
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config ./config
COPY --from=builder /app/data ./data

# Create data directory and set permissions
RUN mkdir -p /app/data && chmod 755 /app/data

# Create a non-root user
RUN adduser -D -g '' appuser
RUN chown -R appuser:appuser /app/data

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
