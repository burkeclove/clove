# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy api-auth module
COPY api-auth/ ./api-auth/

# Set working directory to the API
WORKDIR /app/api-auth

# Download dependencies (without go.work)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-api .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-auth/auth-api .

# Copy the .env file
COPY api-auth/.env .env

# Copy the keys directory
COPY api-auth/keys ./keys

# Expose ports (8080 for HTTP, 50051 for gRPC)
EXPOSE 8080 50051

# Run the binary
CMD ["./auth-api"]
