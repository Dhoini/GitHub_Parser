FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies for protobuf
RUN apk add --no-cache git protobuf

# Copy only go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server

# Use a minimal alpine image for the final container
FROM alpine:3.17

WORKDIR /app

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Create a non-root user to run the application
RUN adduser -D -H -h /app appuser
USER appuser

# Copy the binary from the builder stage
COPY --from=builder --chown=appuser:appuser /app/server .

# Expose gRPC and Prometheus metrics ports
EXPOSE 50051
EXPOSE 9090

# Set environment variables
ENV SERVER_PORT=50051
ENV METRICS_PORT=9090

# Run the application
CMD ["./server"]