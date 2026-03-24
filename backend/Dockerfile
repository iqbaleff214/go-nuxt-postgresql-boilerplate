# Stage 1: build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

RUN go build -o api ./cmd/api && go build -o worker ./cmd/worker

# Stage 2: runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrations ./migrations
# Email templates are embedded in the binary via go:embed — no copy needed

EXPOSE 8080
