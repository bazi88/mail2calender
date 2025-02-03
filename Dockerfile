# Cache dependencies stage
FROM golang:1.22.7-alpine AS deps
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Build stage
FROM golang:1.22.7-alpine AS builder
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/go8

# Security scan stage  
FROM aquasec/trivy:latest AS security-scan
COPY --from=builder /app/main /app/main
RUN trivy filesystem --no-progress --ignore-unfixed --severity HIGH,CRITICAL /app/main

# Final stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates && \
    adduser -D -g '' appuser

USER appuser
WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/config /app/config

ENV GO_ENV=production \
    TZ=UTC \
    GIN_MODE=release

EXPOSE 3080

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3080/health || exit 1

STOPSIGNAL SIGTERM
ENTRYPOINT ["/app/main"]
