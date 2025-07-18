# Multi-stage Dockerfile for MarketFlow

# Build stage
FROM golang:1.22.6-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o marketflow ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh marketflow

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/marketflow .

# Copy configuration files
COPY --from=builder /app/config ./config

# Change ownership
RUN chown -R marketflow:marketflow /app

# Switch to non-root user
USER marketflow

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./marketflow"]