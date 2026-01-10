# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy api-minio module
COPY api-minio/ ./api-minio/

# Set working directory to the API
WORKDIR /app/api-minio

# Download dependencies (without go.work)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minio-api .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-minio/minio-api .

# Copy the .env file
COPY api-minio/.env .env

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./minio-api"]
