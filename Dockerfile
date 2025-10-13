FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o tinyurl-server ./cmd/server/

RUN go build -o tinyurl-cli ./cmd/cli/


FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/tinyurl-server .
COPY --from=builder /app/tinyurl-cli .

VOLUME /data

ENV TINYURL_DB_PATH="file:/data/tinyurl.db?cache=shared&mode=rwc&_fk=1"
ENV PORT="8080"

EXPOSE 8080
CMD ["./tinyurl-server"]