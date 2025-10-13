FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o tinyurl-migrate ./cmd/migrate/

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/tinyurl-migrate .
COPY ./internal/storage/sqlite/migrations ./migrations

VOLUME /data

ENV DATABASE_DSN="file:/data/tinyurl.db?_journal_mode=WAL&_busy_timeout=5000"

ENTRYPOINT ["./tinyurl-migrate"]