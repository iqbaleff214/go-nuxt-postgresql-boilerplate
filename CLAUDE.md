# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

This repository is **in the specification phase**. The `backend/` and `frontend/` directories are currently empty. The PRD.md file defines the full requirements and is the authoritative source for implementation decisions.

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Nuxt.js 3 (Vue 3), Tailwind CSS, shadcn-vue |
| Frontend State | Pinia |
| Frontend Validation | VeeValidate + Zod/valibot |
| Package Manager (FE) | pnpm |
| Backend | Go 1.25+ with Gin framework |
| Database | PostgreSQL 16+ via sqlc (type-safe SQL, no ORM) |
| Auth | JWT (access + refresh tokens) + TOTP 2FA |
| Real-time | Gorilla WebSocket |
| Background Jobs | Asynq (Redis-backed task queue) |
| Storage | S3-compatible (AWS S3 / Cloudflare R2 / MinIO) or local |
| Email | SMTP / SendGrid / Resend (configurable via env) |
| Containers | Docker + Docker Compose |

## Commands (Planned)

### Backend
```bash
cd backend
go run ./cmd/api          # Run API server
go run ./cmd/worker       # Run Asynq background worker
go build -o api ./cmd/api
go build -o worker ./cmd/worker
go test ./...
golangci-lint run
```

### Frontend
```bash
cd frontend
pnpm install
pnpm dev                  # Dev server on :3000
pnpm build
pnpm lint
```

### Docker (full stack)
```bash
docker compose -f docker-compose.dev.yml up   # Local dev (includes MailHog, MinIO)
docker compose up                             # Production-like
```

Database migrations run automatically on backend startup.

## Architecture

```
Nuxt.js FE (:3000) ←→ Go/Gin API (:8080) ←→ PostgreSQL (:5432)
                              ↕
                         Redis (:6379)
                         ├── Asynq Worker (bg jobs)
                         └── Pub/sub (WebSocket fan-out)
                              ↕
                    Cloud/MinIO Storage + SMTP/Mail
```

- Frontend communicates with backend via REST (`/api/v1/...`) and WebSocket (`/ws/notifications`)
- WebSocket auth uses `?token=<access_token>` query param on connect
- Background jobs (email, notifications, maintenance) run in a separate `worker` process sharing the same codebase as `backend`, with a different entrypoint (`cmd/worker`)
- Swagger/OpenAPI available at `/docs` in dev mode (via `swaggo`)

## Backend Structure

```
backend/
├── cmd/api/          # HTTP server entrypoint
├── cmd/worker/       # Asynq worker entrypoint
├── internal/
│   ├── api/
│   │   ├── handler/ # auth.go, profile.go, users.go, notifications.go
│   │   ├── middleware/
│   │   └── router.go
│   ├── ws/           # WebSocket notification hub
│   ├── core/         # config, database, redis, security helpers
│   ├── model/        # domain models
│   ├── repository/   # sqlc-generated DB layer
│   ├── service/      # auth, totp, user, email, notification, storage services
│   ├── jobs/         # Asynq job definitions
│   └── templates/email/
└── migrations/       # SQL migrations (golang-migrate or atlas)
```

## Frontend Structure

```
frontend/
├── composables/
│   ├── useApi.ts           # Typed HTTP client with auto token refresh
│   └── useNotifications.ts # WebSocket lifecycle + auto-reconnect
├── stores/
│   ├── auth.ts             # Access token, user state, 2FA challenge state
│   └── notifications.ts    # Unread count, notification list
├── middleware/             # auth, guest, superadmin, mfa route guards
├── layouts/                # default, auth, admin
└── pages/                  # Route components
```

## API Response Envelope

```json
{ "success": true, "message": "...", "data": {}, "meta": { "page": 1, "total": 100 } }
{ "success": false, "message": "Validation error", "errors": [{ "field": "email", "detail": "..." }] }
```

## Database Schema

Three core tables: `users`, `tokens`, `notifications`. Key design notes:
- All tokens stored **hashed** (never plaintext)
- `tokens.type` enum: `refresh`, `email_verify`, `password_reset`, `email_change`, `delete_cancel`, `mfa_challenge`, `totp_recovery`
- `users.totp_secret` encrypted at rest (AES-256)
- Soft delete on users via `deleted_at`; hard delete scheduled 30 days later via cron job
- `notifications.read_at` is null when unread

## Key Security Constraints

- bcrypt cost ≥ 12 for passwords; bcrypt for 2FA recovery codes
- Refresh token rotated on each use; stored as HTTP-only, Secure, SameSite=Strict cookie
- JWT payload contains only user ID and role (no sensitive data)
- Rate limiting required on: register, login, forgot-password, TOTP verify endpoints
- CORS restricted to frontend origin only
- File uploads: validate MIME type (sniffing), size ≤ 2MB for avatars

## Background Jobs (Asynq)

| Job | Trigger |
|---|---|
| `send_email` | Any email event |
| `send_notification` | Any notification event |
| `cleanup_expired_tokens` | Daily cron |
| `hard_delete_accounts` | Daily cron (30-day soft-delete window) |
| `broadcast_announcement` | Superadmin action |

Max 3 retries with exponential backoff. System announcements fan out via Redis pub/sub.

## Environment Variables

See PRD.md §10 for the full list. Key variables:
- `DATABASE_URL`, `REDIS_URL`
- `SECRET_KEY`, `TOTP_ENCRYPTION_KEY`
- `MAIL_PROVIDER` (smtp | sendgrid | resend)
- `STORAGE_BACKEND` (local | s3)
- `NUXT_PUBLIC_API_BASE`, `NUXT_PUBLIC_WS_BASE`

## Dev Services (docker-compose.dev.yml only)

- **MailHog** (:8025 UI, :1025 SMTP) — catches all outgoing emails
- **MinIO** (:9000 S3 API, :9001 console) — local S3-compatible storage

## Out of Scope (v1.0)

OAuth/social login, multi-tenancy, audit log UI, admin impersonation, push notifications (FCM/APNs), SMS-based 2FA.
