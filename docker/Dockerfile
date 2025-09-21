# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o konnect-api .

# Final stage
FROM alpine:latest

# Create app directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/konnect-api .

# Copy environment file (optional, can be overridden with volume mount)
COPY --from=builder /app/.env .env

# Command to run the application
CMD ["./konnect-api"]