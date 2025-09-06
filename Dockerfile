FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache --no-install-recommends tzdata curl mysql-client && \
    addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser && \
    mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/config.yaml /app/config.yaml
COPY --from=builder /app/db/schema /app/db/schema

USER appuser
WORKDIR /app
EXPOSE 9876

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9876/health || exit 1

ENTRYPOINT ["/app/server"]
