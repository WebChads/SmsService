# Build stage
FROM golang:1.24.3 AS builder

WORKDIR /app

# Build dependencies for Kafka
RUN apt-get update && \
    apt-get install -y librdkafka-dev pkg-config

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -o sms-service ./cmd/app/main.go

# Run stage
FROM debian:stable-slim

WORKDIR /app

# Runtime dependencies for Kafka
RUN apt-get update && \
    apt-get install -y librdkafka1 && \
    rm -rf /var/lib/apt/lists/*

# Binary from the builder stage
COPY --from=builder /app/sms-service .

# Environment variables for configuration
ENV KAFKA_ADDRESS=localhost:9092
ENV BROKERS=localhost:9092

# Run the application
CMD ["./sms-service"]