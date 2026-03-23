.PHONY: dev down prod backend-build backend-test backend-lint fe-install fe-dev fe-build fe-lint migrate-up migrate-down docs sqlc

dev:
	docker compose -f docker-compose.dev.yml up --build

down:
	docker compose -f docker-compose.dev.yml down

prod:
	docker compose up --build

## Backend
backend-build:
	cd backend && go build -o api ./cmd/api && go build -o worker ./cmd/worker

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && golangci-lint run

docs:
	cd backend && swag init -g cmd/api/main.go -o docs/

sqlc:
	cd backend && sqlc generate

migrate-up:
	migrate -path ./backend/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./backend/migrations -database "$(DATABASE_URL)" down 1

## Frontend
fe-install:
	cd frontend && pnpm install

fe-dev:
	cd frontend && pnpm dev

fe-build:
	cd frontend && pnpm build

fe-lint:
	cd frontend && pnpm lint
