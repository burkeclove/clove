FROM golang:1.24-bookworm as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -v -o server

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates \
    ipmitool && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary
COPY --from=builder /app/server /app/server

# Copy environment-specific .env file based on build arg
ARG BUILD_MODE=local
COPY .env.${BUILD_MODE} /.env

EXPOSE 50000
CMD ["/app/server"]
