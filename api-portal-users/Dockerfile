# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy api-portal-users module
COPY api-portal-users/ ./api-portal-users/

# Set working directory to the API
WORKDIR /app/api-portal-users

# Download dependencies (without go.work)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o users-api .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-portal-users/users-api .

# Copy the .env file
COPY api-portal-users/.env .env

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./users-api"]
