# Build stage
FROM golang:1.23-alpine AS builder

# Cài đặt các build dependencies
RUN apk add --no-cache git make build-base

# Tạo non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build ứng dụng với các tối ưu hóa
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/go8

# Final stage
FROM alpine:3.19

# Cài đặt các security packages
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Tạo non-root user
COPY --from=builder /etc/passwd /etc/passwd
USER appuser

# Copy binary từ build stage
COPY --from=builder /app/main /app/main

# Copy các file cấu hình cần thiết
COPY --from=builder /app/config /app/config
COPY --from=builder /app/.env.example /app/.env

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 3080

# Set các biến môi trường
ENV GO_ENV=production \
    TZ=UTC \
    GIN_MODE=release

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3080/health || exit 1

# Security options
STOPSIGNAL SIGTERM

# Chạy ứng dụng
ENTRYPOINT ["/app/main"]
