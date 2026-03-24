# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

RUN go build -o api ./cmd/api && go build -o worker ./cmd/worker

# Stage 2a: API server runtime
FROM alpine:3.19 AS api

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations
# Email templates are embedded in the binary via go:embed — no copy needed

EXPOSE 8080
CMD ["./api"]

# Stage 2b: Worker runtime
FROM alpine:3.19 AS worker

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/worker .
COPY --from=builder /app/migrations ./migrations

CMD ["./worker"]
