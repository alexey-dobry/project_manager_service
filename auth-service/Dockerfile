# syntax=docker/dockerfile:1.7

# ---------- builder ----------
FROM golang:1.22-alpine AS builder

WORKDIR /src

# отдельный слой для модулей — кэш не инвалидируется при правке исходников
COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/auth-service \
    ./cmd

# ---------- runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/auth-service /app/auth-service
COPY migrations /app/migrations

USER app
EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8081/health || exit 1

ENTRYPOINT ["/app/auth-service"]
