FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies for protobuf
RUN apk add --no-cache git protobuf

# Copy only go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

# Copy the rest of the code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server