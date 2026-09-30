# ─── STAGE 1: Builder ──────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata git

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /build/simak-server cmd/server/main.go

# ─── STAGE 2: Minimal Production Runner ──────────────────────────────────────
FROM alpine:3.20

# Install runtime dependencies (SSL certificates & timezone)
RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone

WORKDIR /app

# Create non-root user for security compliance
RUN addgroup -S simak && adduser -S -G simak -u 10001 simak

# Copy compiled binary and necessary runtime assets
COPY --from=builder /build/simak-server /app/simak-server
COPY --from=builder /build/migrations /app/migrations
COPY --from=builder /build/docs /app/docs

# Set file permissions for non-root execution
RUN chown -R simak:simak /app
USER simak

EXPOSE 8080

# Healthcheck probe using internal health endpoint
HEALTHCHECK --interval=20s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/simak-server"]
