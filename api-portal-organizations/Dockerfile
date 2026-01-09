# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy api-portal-organizations module
COPY api-portal-organizations/ ./api-portal-organizations/

# Set working directory to the API
WORKDIR /app/api-portal-organizations

# Download dependencies (without go.work)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o organizations-api .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-portal-organizations/organizations-api .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./organizations-api"]
