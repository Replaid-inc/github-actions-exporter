# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy the entire project to maintain module structure
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gh-actions-exporter ./cmd/gh-actions-exporter

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/gh-actions-exporter .

# Make sure the binary is executable
RUN chmod +x ./gh-actions-exporter

# Set Gin to release mode
ENV GIN_MODE=release

CMD ["./gh-actions-exporter"]