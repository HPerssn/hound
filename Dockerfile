# Use a Go image that matches your version
FROM golang:1.25-alpine AS builder

# Install dependencies for building with CGO
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go.mod first and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Enable CGO for SQLite
ENV CGO_ENABLED=1

# Build the binary
RUN go build -o hound ./cmd/server

# Final minimal image
FROM alpine:3.18
RUN apk add --no-cache sqlite
COPY --from=builder /app/hound /usr/local/bin/hound
COPY --from=builder /app/static ./static

EXPOSE 8080
ENV HOUND_DB_DSN="file:/data/hound.db?cache=shared&_foreign_keys=1"

ENTRYPOINT ["/usr/local/bin/hound"]
