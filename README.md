# Go + Nuxt Starter Kit

A production-ready full-stack boilerplate with a Go/Gin backend and Nuxt.js 3 frontend. Fork it and build on top.

## Features

- **Auth** — Register, email verification, login, logout, token refresh, forgot/reset password, change password
- **2FA** — TOTP setup via QR code, login challenge, recovery codes, disable
- **Profile** — View/edit, avatar upload, email change, self-service account deletion (30-day soft-delete)
- **Admin** — User list with search/filter/pagination, create/edit/activate/deactivate/ban/unban/delete
- **Real-time notifications** — WebSocket delivery with Redis pub/sub fan-out; persisted in DB
- **Background jobs** — Asynq (Redis-backed) worker for emails, notifications, announcements, daily cron cleanup
- **Storage** — Local (dev) or S3-compatible (AWS S3, Cloudflare R2, MinIO)
- **Email** — SMTP, SendGrid, or Resend; 9 HTML templates
- **Swagger UI** — `/docs` in dev mode (swaggo)
- **Docker** — Single `docker compose up` starts everything

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Nuxt.js 3 (Vue 3), Tailwind CSS |
| State | Pinia |
| Validation | VeeValidate + Zod |
| Backend | Go 1.25+, Gin |
| Database | PostgreSQL 16+, sqlc |
| Auth | JWT (access + refresh) + TOTP 2FA |
| Real-time | Gorilla WebSocket |
| Jobs | Asynq + Redis |
| Storage | S3-compatible or local |
| Containers | Docker + Docker Compose |

## Quick Start

**Prerequisites:** Docker, Docker Compose

```bash
git clone https://github.com/404nfidv2/go-nuxt-starter-kit.git
cd go-nuxt-starter-kit
cp .env.example .env
```

Edit `.env` — at minimum set:
```
SECRET_KEY=<random string>
TOTP_ENCRYPTION_KEY=<exactly 32 characters>
```

```bash
make dev
```

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Swagger UI | http://localhost:8080/docs/index.html |
| MailHog | http://localhost:8025 |
| MinIO console | http://localhost:9001 (minioadmin / minioadmin) |

## Development Commands

```bash
# Full dev stack (hot-reload frontend + backend + all services)
make dev

# Individual services
make fe-dev          # Nuxt dev server only (:3000)
cd backend && go run ./cmd/api     # Go API server only (:8080)
cd backend && go run ./cmd/worker  # Asynq worker only

# Code generation
make docs            # Regenerate Swagger docs (swag init)
make sqlc            # Regenerate sqlc query bindings

# Migrations (set DATABASE_URL first)
make migrate-up
make migrate-down

# Build
make backend-build   # Produces ./backend/api and ./backend/worker
make fe-build        # Nuxt SSR build

# Quality
make backend-test    # go test ./...
make backend-lint    # golangci-lint run
make fe-lint         # pnpm lint
```

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   ├── api/          # HTTP server entrypoint
│   │   └── worker/       # Asynq worker + scheduler
│   └── internal/
│       ├── api/
│       │   ├── handler/  # HTTP handlers (auth, profile, admin, notifications)
│       │   ├── middleware/
│       │   └── router.go
│       ├── core/         # Config, DB pool, Redis, security helpers
│       ├── jobs/         # Asynq job definitions and handlers
│       ├── model/        # Domain models
│       ├── repository/   # sqlc-generated DB layer
│       ├── service/      # Business logic (auth, user, mailer, storage, notification)
│       ├── templates/    # Email HTML templates (go:embed)
│       └── ws/           # WebSocket hub + client
├── frontend/
│   ├── composables/
│   │   ├── useApi.ts           # Typed HTTP client with auto token refresh
│   │   └── useNotifications.ts # WebSocket lifecycle + reconnect
│   ├── stores/
│   │   ├── auth.ts             # Access token, user state, 2FA challenge
│   │   └── notifications.ts    # Unread count, notification list
│   ├── middleware/             # auth, guest, superadmin, mfa route guards
│   ├── layouts/                # default, auth, admin
│   └── pages/                  # All route components
├── migrations/                 # SQL migrations (auto-run on startup)
├── docker-compose.yml          # Production
├── docker-compose.dev.yml      # Development (+ MailHog, MinIO)
├── Makefile
└── .env.example
```

## API Overview

Base path: `/api/v1`

All responses follow the envelope format:
```json
{ "success": true,  "message": "...", "data": {} }
{ "success": false, "message": "...", "errors": [{ "field": "...", "detail": "..." }] }
```

| Group | Endpoints |
|---|---|
| Auth | `POST /auth/register`, `/auth/verify-email`, `/auth/resend-verification`, `/auth/login`, `/auth/logout`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`, `/auth/change-password` |
| 2FA | `POST /auth/2fa/setup`, `/auth/2fa/confirm`, `/auth/2fa/disable`, `/auth/2fa/verify`, `/auth/2fa/recovery-codes/regenerate` |
| Profile | `GET/PATCH /profile`, `POST /profile/avatar`, `/profile/email`, `/profile/email/confirm`, `/profile/delete`, `/profile/delete/cancel` |
| Notifications | `GET /notifications`, `/notifications/unread-count`, `PATCH /notifications/read-all`, `/notifications/:id/read` |
| Admin | `GET/POST /admin/users`, `GET/PATCH/DELETE /admin/users/:id`, `POST /admin/users/:id/{activate,deactivate,ban,unban}`, `POST /admin/announcements` |
| System | `GET /health`, `WS /ws/notifications` |

Rate-limited endpoints: register, login, resend-verification, forgot-password, 2FA verify.

## Environment Variables

See [`.env.example`](.env.example) for the full list. Key variables:

| Variable | Description |
|---|---|
| `SECRET_KEY` | JWT signing secret |
| `TOTP_ENCRYPTION_KEY` | AES-256 key for TOTP secrets — must be exactly 32 chars |
| `DATABASE_URL` | PostgreSQL DSN |
| `REDIS_URL` | Redis URL (`redis://host:port`) |
| `MAIL_PROVIDER` | `smtp` \| `sendgrid` \| `resend` |
| `STORAGE_BACKEND` | `local` \| `s3` |
| `NUXT_PUBLIC_API_BASE` | Frontend → backend API base URL |
| `NUXT_PUBLIC_WS_BASE` | Frontend → backend WebSocket base URL |

## Production Deployment

```bash
cp .env.example .env
# Fill in production values
docker compose up -d
```

The production compose file builds both the `api` and `worker` binaries from the same `backend/Dockerfile` using multi-stage build targets. Migrations run automatically on API server startup.

## Out of Scope (v1.0)

OAuth/social login, multi-tenancy, audit log UI, admin impersonation, push notifications (FCM/APNs), SMS-based 2FA.
