FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1

RUN go build -o hound ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache sqlite
COPY --from=builder /app/hound /usr/local/bin/hound
COPY --from=builder /app/static ./static

EXPOSE 8080
ENV HOUND_DB_DSN="file:/data/hound.db?cache=shared&_foreign_keys=1"

ENTRYPOINT ["/usr/local/bin/hound"]
