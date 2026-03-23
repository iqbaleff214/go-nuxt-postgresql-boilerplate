# Product Requirements Document (PRD)
## Full-Stack Boilerplate: Nuxt.js + Golang (Gin)

**Version:** 1.2.0
**Date:** 2026-03-24
**Status:** Draft

---

## 1. Overview

This document outlines the requirements for a full-stack boilerplate application that provides a production-ready foundation for web application development. The stack consists of a Nuxt.js frontend with Tailwind CSS and a Golang (Gin) backend with PostgreSQL, complete with authentication (including 2FA), profile management, user management, email integration, real-time in-app notifications via WebSocket, background job processing, cloud storage integration, and Docker support.

---

## 2. Goals

- Provide a reusable, scalable boilerplate that can be forked and adapted for any web application project.
- Implement a secure, complete authentication system with 2FA support out of the box.
- Enable superadmin-level user management for operational control.
- Support transactional and notification emails via an email service integration.
- Deliver real-time in-app notifications via WebSocket.
- Offload long-running or scheduled tasks to a background worker system.
- Support cloud storage (S3-compatible) for file uploads with a local fallback for development.
- Ensure the entire stack runs consistently across environments via Docker.

---

## 3. Tech Stack

| Layer           | Technology                                                                 |
|-----------------|----------------------------------------------------------------------------|
| Frontend        | Nuxt.js 3 (Vue 3)                                                          |
| UI Framework    | Tailwind CSS + shadcn-vue / Headless UI                                    |
| Backend         | Gin (Golang 1.22+)                                                         |
| Database        | PostgreSQL 16+                                                             |
| ORM             | sqlc                                                  |
| Auth            | JWT (access + refresh tokens) + TOTP 2FA                                   |
| Real-time       | WebSocket (Gorilla WebSocket / Fiber WS optional)                          |
| Background Jobs | Asynq (Redis-based task queue)                                             |
| Cloud Storage   | S3-compatible (AWS S3 / Cloudflare R2 / MinIO)                             |
| Email           | SMTP / SendGrid / Resend (configurable)                                    |
| Container       | Docker + Docker Compose                                                    |
| Package Mgr     | pnpm (frontend), go modules (backend)                                      |

---

## 4. System Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                          Docker Network                            │
│                                                                    │
│  ┌──────────────┐  HTTP  ┌──────────────┐       ┌───────────────┐  │
│  │  Nuxt.js FE  │───────▶│   Go (Gin)   │──────▶│  PostgreSQL   │  │
│  │  :3000       │  WS    │   :8080      │       │  :5432        │  │
│  └──────────────┘◀──────▶└──────┬───────┘       └───────────────┘  │
│                                 │                                  │
│              ┌──────────────────┼───────────────────┐              │
│              ▼                  ▼                   ▼              │
│       ┌────────────┐    ┌─────────────┐    ┌────────────────┐      │
│       │  SMTP/Mail │    │    Redis    │    │  Cloud / MinIO │      │
│       │  Service   │    │  :6379      │    │  :9000 (dev)   │      │
│       └────────────┘    └──────┬──────┘    └────────────────┘      │
│                                │                                   │
│                         ┌──────▼───────┐                           │
│                         │  Asynq Worker│                           │
│                         │  (bg jobs)   │                           │
│                         └──────────────┘                           │
└────────────────────────────────────────────────────────────────────┘
```

---

## 5. Features & Requirements

### 5.1 Authentication System

#### 5.1.1 Registration
- User self-registration with email and password.
- Password validation: min 8 characters, at least 1 uppercase, 1 number, 1 special character.
- Email uniqueness check.
- Send email verification link upon registration.
- Account remains inactive until email is verified.

#### 5.1.2 Email Verification
- Verification link is time-limited (expires in 24 hours).
- Token is single-use and invalidated after use.
- Resend verification email option.

#### 5.1.3 Login
- Login with email and password.
- If 2FA is enabled on the account, return a short-lived `mfa_challenge_token` instead of full tokens; require a TOTP code to complete login.
- Return JWT access token (short-lived, 15 min) and refresh token (long-lived, 7 days) only after all auth factors are verified.
- Refresh token is stored as an HTTP-only cookie.
- Failed login attempt tracking (lockout after 5 consecutive failures for 15 minutes).

#### 5.1.4 Logout
- Invalidate/revoke the active refresh token on logout.
- Clear HTTP-only cookie.

#### 5.1.5 Token Refresh
- Silent token refresh via dedicated `/auth/refresh` endpoint using the refresh token cookie.
- Frontend automatically refreshes access token on expiry.

#### 5.1.6 Forgot Password
- User requests password reset by providing their email.
- Backend sends a password reset link with a time-limited token (expires in 1 hour).
- Token is single-use.

#### 5.1.7 Reset Password
- User submits new password using the token from the reset email.
- Notify user by email upon successful password reset.

#### 5.1.8 Change Password
- Authenticated users can change their password by providing current + new password.
- Notify user by email upon successful password change.

---

### 5.2 Profile Management

All features below are accessible to authenticated users for their own profile.

#### 5.2.1 View Profile
- Retrieve full profile info: name, email, avatar, bio, created date, role.

#### 5.2.2 Update Profile
- Update: first name, last name, display name, bio.
- Input validation for all fields.

#### 5.2.3 Upload Avatar
- Upload profile picture (JPEG, PNG, WebP; max 2MB).
- Backend stores the file and returns a URL.
- Old avatar is deleted on replacement.

#### 5.2.4 Update Email
- User requests email change by providing new email.
- Send verification link to the new email address.
- Email is updated only after the new address is verified.
- Old email receives a notification of the change.

#### 5.2.5 Account Deletion (Self)
- Authenticated user can request account deletion.
- Confirmation prompt required on frontend.
- Soft-delete: account marked inactive; hard delete after 30 days.
- User receives a confirmation email with an undo/cancel link (valid for 30 days).

---

### 5.3 User Management (Superadmin)

Only users with the `superadmin` role can access these features.

#### 5.3.1 List Users
- Paginated list of all users.
- Filterable by: role, status (active/inactive/banned), email verification status.
- Searchable by: name, email.
- Sortable by: created date, name, email.

#### 5.3.2 View User Detail
- View full profile of any user including metadata: last login, IP (optional), status history.

#### 5.3.3 Create User
- Admin can create a new user account directly.
- Assign role at creation.
- Send a set-password email to the new user.

#### 5.3.4 Update User
- Edit any user's: name, bio, role, status.
- Cannot change another superadmin's role.

#### 5.3.5 Activate / Deactivate User
- Toggle user active status.
- Deactivated users cannot log in.
- Notify user by email of deactivation/reactivation.

#### 5.3.6 Ban / Unban User
- Permanently ban a user with an optional reason.
- Banned users cannot log in.
- Notify user by email of ban.

#### 5.3.7 Delete User
- Hard delete a user account.
- Cannot delete own account or another superadmin.
- All associated data is removed or anonymized.

#### 5.3.8 Impersonate User (Optional / Phase 2)
- Superadmin can temporarily act as another user for debugging.
- Impersonation is logged and time-limited.

---

### 5.4 Role & Permission System

| Role        | Description                                  |
|-------------|----------------------------------------------|
| `superadmin`| Full access including user management        |
| `user`      | Standard authenticated user                  |

- Roles stored in the database.
- Middleware/guards on both frontend and backend enforce role-based access.
- Default role for self-registration is `user`.

---

### 5.5 Email System

#### 5.5.1 Email Templates
All transactional emails use HTML templates with a consistent brand layout. Required templates:

| Template                  | Trigger                              |
|---------------------------|--------------------------------------|
| Welcome / Verify Email    | Registration                         |
| Email Verified            | After successful verification        |
| Password Reset            | Forgot password request              |
| Password Changed          | After password change/reset          |
| Email Change Verification | User requests email change           |
| Account Deactivated       | Admin deactivates account            |
| Account Banned            | Admin bans account                   |
| Account Deletion Confirm  | User requests account deletion       |
| New Account by Admin      | Admin creates account for user       |

#### 5.5.2 Email Configuration
- Configurable via environment variables.
- Support for: SMTP (any provider), SendGrid, Resend.
- Email sending is dispatched as an ARQ background job (non-blocking).

---

### 5.6 Two-Factor Authentication (2FA)

#### 5.6.1 Setup
- Authenticated users can enable TOTP-based 2FA from the security settings page.
- Backend generates a TOTP secret and returns a `otpauth://` URI and a QR code (as a data URI).
- User scans the QR code with an authenticator app (Google Authenticator, Authy, etc.).
- User confirms setup by submitting a valid TOTP code — 2FA is only activated after successful confirmation.
- Backend generates and returns 8 single-use recovery codes upon activation; stores them as hashes.
- Recovery codes are shown **once** and must be saved by the user.

#### 5.6.2 Login with 2FA
- After valid email/password, if 2FA is enabled:
  - Return a `mfa_challenge_token` (short-lived, 5 min, single-use).
  - Frontend redirects to a 2FA verification page.
  - User submits a 6-digit TOTP code or a recovery code.
  - Backend verifies the code and, if valid, issues the full access + refresh tokens.
- Brute-force protection: max 5 wrong TOTP attempts before the challenge token is invalidated.

#### 5.6.3 Disable 2FA
- User can disable 2FA by providing their current password and a valid TOTP code.
- All existing recovery codes are invalidated.
- User receives an email notification that 2FA was disabled.

#### 5.6.4 Recovery Codes
- User can regenerate recovery codes (requires current password + valid TOTP code).
- Old codes are invalidated immediately on regeneration.
- Each recovery code is single-use and marked as consumed after use.

---

### 5.7 Real-Time In-App Notifications (WebSocket)

#### 5.7.1 Connection
- Authenticated users connect to `ws://<host>/ws/notifications` using their JWT access token (passed as a query param or initial message).
- Backend validates the token on connection; unauthenticated connections are rejected.
- Connection is maintained per user session; multiple tabs/devices are supported.

#### 5.7.2 Notification Types
Initial supported notification types:

| Type                   | Trigger                                      |
|------------------------|----------------------------------------------|
| `account.status_change`| Admin activates, deactivates, or bans user   |
| `account.role_change`  | Admin changes user's role                    |
| `security.2fa_disabled`| 2FA disabled on the account                  |
| `security.password_changed` | Password changed or reset               |
| `system.announcement`  | Superadmin broadcasts a message to all users |

#### 5.7.3 Notification Payload
```json
{
  "id": "uuid",
  "type": "account.status_change",
  "title": "Your account has been deactivated",
  "body": "Contact support for more information.",
  "read": false,
  "created_at": "2026-03-23T10:00:00Z"
}
```

#### 5.7.4 Persistence
- All notifications are persisted in the `notifications` table.
- Users can fetch their notification history via REST (`GET /api/v1/notifications`).
- Mark notification as read: `PATCH /api/v1/notifications/:id/read`.
- Mark all as read: `PATCH /api/v1/notifications/read-all`.
- Unread count accessible via `GET /api/v1/notifications/unread-count`.

#### 5.7.5 Delivery
- When a notification is created, the backend:
  1. Persists it to the database.
  2. Pushes it to the user's active WebSocket connection(s) in real time.
  3. If the user is offline, the notification is delivered on next connection via a "pending" fetch.
- Broadcasting to all users (system announcements) fans out via Redis pub/sub to all backend instances.

---

### 5.8 Background Jobs (Asynq + Redis)

#### 5.8.1 Worker Setup
- Asynq is used as the job queue.
- A dedicated `worker` Docker service consumes jobs from Redis.
- Jobs are defined in `backend/internal/jobs/`.
- Worker is horizontally scalable.

#### 5.8.2 Registered Jobs

| Job                        | Trigger                                      | Description                                   |
|----------------------------|----------------------------------------------|-----------------------------------------------|
| `send_email`               | Any email-triggering event                   | Send transactional email                      |
| `send_notification`        | Any notification-triggering event            | Persist + push WebSocket notification         |
| `cleanup_expired_tokens`   | Scheduled (cron, daily)                      | Delete expired/used tokens                    |
| `hard_delete_accounts`     | Scheduled (cron, daily)                      | Hard delete accounts > 30 days                |
| `broadcast_announcement`   | Superadmin triggers                          | Broadcast system notification                 |

#### 5.8.3 Job Configuration
- Max retries: 3
- Exponential backoff
- Configurable timeout per job
- Failed jobs logged
- Cron jobs via Asynq scheduler

---

### 5.9 Cloud Storage

#### 5.9.1 Backends
- **Local** (default for development): files stored at a configurable local path, served as static files.
- **S3-compatible** (production): supports AWS S3, Cloudflare R2, MinIO; configured via env vars.
- MinIO included as a Docker service in development for local S3-compatible testing.

#### 5.9.2 Stored File Types
| Category      | Max Size | Allowed Types            | Path Pattern                   |
|---------------|----------|--------------------------|--------------------------------|
| User avatars  | 2 MB     | JPEG, PNG, WebP          | `avatars/{user_id}/{filename}` |

#### 5.9.3 Upload Flow
1. Client sends multipart form data to `POST /api/v1/profile/avatar`.
2. Backend validates file type (MIME sniffing) and size.
3. Backend uploads the file to the configured storage backend.
4. Backend returns the public URL and updates `users.avatar_url`.
5. Previous avatar file is deleted from storage on replacement.

#### 5.9.4 Signed URLs (S3 mode)
- Private buckets are supported; the backend generates pre-signed read URLs with short TTLs when serving private files.
- For public-readable buckets (e.g., avatars), the raw CDN URL is stored and returned directly.

#### 5.9.5 Storage Service Interface
The `StorageService` abstraction exposes:
- `upload(file, path) → url`
- `delete(path)`
- `get_signed_url(path, expires_in) → url`

Both local and S3 backends implement this interface, making the storage backend fully swappable via config.

---

### 5.10 API Design (REST + WebSocket)

- RESTful API with versioning (`/api/v1/...`).
- WebSocket endpoint at `/ws/notifications` for real-time events.
- Swagger/OpenAPI via `swaggo/swag` (or similar), available at `/docs` (dev only)
- All protected REST routes require `Authorization: Bearer <access_token>` header.
- WebSocket authentication via `?token=<access_token>` query parameter on connect.
- Consistent JSON response envelope:

```json
{
  "success": true,
  "message": "...",
  "data": { ... },
  "meta": { "page": 1, "total": 100 }
}
```

- Standardized error responses:

```json
{
  "success": false,
  "message": "Validation error",
  "errors": [{ "field": "email", "detail": "Invalid email format" }]
}
```

---

### 5.11 Frontend Structure (Nuxt.js)

#### Pages / Routes

| Route                       | Access       | Description                               |
|-----------------------------|--------------|-------------------------------------------|
| `/`                         | Public       | Landing / redirect to dashboard           |
| `/login`                    | Guest only   | Login form                                |
| `/login/2fa`                | Guest only   | TOTP verification step after login        |
| `/register`                 | Guest only   | Registration form                         |
| `/verify-email`             | Public       | Email verification handler                |
| `/forgot-password`          | Guest only   | Request password reset                    |
| `/reset-password`           | Guest only   | Password reset form (via token)           |
| `/dashboard`                | Auth         | Main dashboard                            |
| `/profile`                  | Auth         | View & edit own profile                   |
| `/profile/security`         | Auth         | Change password, manage 2FA               |
| `/profile/security/2fa`     | Auth         | 2FA setup wizard (QR code + recovery codes) |
| `/notifications`            | Auth         | Full notification history                 |
| `/admin/users`              | Superadmin   | User list                                 |
| `/admin/users/[id]`         | Superadmin   | User detail / edit                        |
| `/admin/users/create`       | Superadmin   | Create new user                           |
| `/admin/announcements`      | Superadmin   | Broadcast system-wide notification        |

#### Frontend Key Modules
- **Auth Store** (Pinia): manage access token, user state, refresh logic, 2FA challenge state.
- **Notification Store** (Pinia): unread count, notification list, real-time push via WebSocket.
- **WebSocket Composable** (`useNotifications`): manages WS connection lifecycle, auto-reconnect, token refresh on reconnect.
- **Route Middleware**: `auth` guard, `guest` guard, `superadmin` guard, `mfa` guard (blocks full app until 2FA step complete).
- **API Composable** (`useApi`): typed HTTP client with auto token refresh.
- **Toast Notifications**: success, error, info system-wide.
- **Form Validation**: using VeeValidate + Zod or valibot.

---

## 6. Database Schema (High Level)

### `users`
| Column               | Type         | Notes                            |
|----------------------|--------------|----------------------------------|
| id                   | UUID (PK)    |                                  |
| email                | VARCHAR      | unique, indexed                  |
| hashed_password      | VARCHAR      |                                  |
| first_name           | VARCHAR      |                                  |
| last_name            | VARCHAR      |                                  |
| display_name         | VARCHAR      |                                  |
| bio                  | TEXT         | nullable                         |
| avatar_url           | VARCHAR      | nullable                         |
| role                 | ENUM         | user, superadmin                 |
| status               | ENUM         | active, inactive, banned         |
| is_email_verified    | BOOLEAN      | default false                    |
| totp_secret          | VARCHAR      | nullable; encrypted at rest      |
| is_2fa_enabled       | BOOLEAN      | default false                    |
| last_login_at        | TIMESTAMP    | nullable                         |
| deleted_at           | TIMESTAMP    | nullable (soft delete)           |
| created_at           | TIMESTAMP    |                                  |
| updated_at           | TIMESTAMP    |                                  |

### `tokens`
| Column       | Type      | Notes                                                                          |
|--------------|-----------|--------------------------------------------------------------------------------|
| id           | UUID (PK) |                                                                                |
| user_id      | UUID (FK) | references users.id                                                            |
| token        | VARCHAR   | hashed token value                                                             |
| type         | ENUM      | refresh, email_verify, password_reset, email_change, delete_cancel, mfa_challenge, totp_recovery |
| expires_at   | TIMESTAMP |                                                                                |
| used_at      | TIMESTAMP | nullable                                                                       |
| created_at   | TIMESTAMP |                                                                                |

### `notifications`
| Column       | Type      | Notes                                         |
|--------------|-----------|-----------------------------------------------|
| id           | UUID (PK) |                                               |
| user_id      | UUID (FK) | references users.id                           |
| type         | VARCHAR   | e.g. `account.status_change`                  |
| title        | VARCHAR   |                                               |
| body         | TEXT      | nullable                                      |
| read_at      | TIMESTAMP | nullable; null = unread                       |
| created_at   | TIMESTAMP |                                               |

---

## 7. Docker Setup

### Services

| Service    | Image                    | Port (host:container)    | Notes                          |
|------------|--------------------------|--------------------------|--------------------------------|
| frontend   | node:22-alpine           | 3000:3000                |                                |
| backend    | golang:1.22-alpine       | 8080:8080                | API + WebSocket                |
| worker     | golang:1.22-alpine       | —                        | Asynq worker                   |
| db         | postgres:16-alpine       | 5432:5432                |                                |
| redis      | redis:7-alpine           | 6379:6379                | Job queue + pub/sub            |
| minio      | minio/minio              | 9000:9000, 9001:9001     | Dev-only S3-compatible storage |
| mailhog    | mailhog/mailhog          | 8025:8025, 1025:1025     | Dev-only SMTP catch-all        |

- `docker-compose.yml` for production (no MailHog, no MinIO — use real cloud providers).
- `docker-compose.dev.yml` for local development with hot reload, MailHog, and MinIO.
- Worker and backend share the same image/codebase; entrypoint differs.
- `.env.example` provided for all services.
- Database migrations run automatically on backend startup.
- Health checks defined for all services.

---

## 8. Security Requirements

- Passwords hashed with bcrypt (cost factor ≥ 12).
- All tokens stored as hashes in the database.
- Refresh token rotated on each use.
- TOTP secret encrypted at rest in the database (AES-256 or Fernet).
- 2FA recovery codes stored as bcrypt hashes; each is single-use.
- CORS configured to allow only the frontend origin.
- Rate limiting on auth endpoints: register, login, forgot-password, TOTP verify.
- HTTP-only, Secure, SameSite=Strict cookie for refresh token.
- No sensitive data in JWT payload (only user ID and role).
- WebSocket connections authenticated on upgrade; unauthenticated connections are rejected.
- Input sanitization on all endpoints.
- SQL injection prevented by ORM parameterized queries.
- File upload validation: type, size, and content (MIME sniffing).
- Cloud storage bucket is private by default; files served via pre-signed URLs where needed.
- SQL injection prevented via **parameterized queries (sqlc) or ORM**
- TOTP secret encrypted using AES-256 (custom Go implementation)
- Rate limiting via middleware (e.g., `golang.org/x/time/rate` or Redis-based)

---

## 9. Project Structure

```
/
├── frontend/                  # Nuxt.js application
│   ├── assets/
│   ├── components/
│   │   ├── auth/
│   │   ├── profile/
│   │   ├── admin/
│   │   └── ui/                # Reusable UI components
│   ├── composables/
│   │   ├── useApi.ts
│   │   ├── useToast.ts
│   │   └── useNotifications.ts    # WebSocket lifecycle
│   ├── layouts/
│   │   ├── default.vue
│   │   ├── auth.vue
│   │   └── admin.vue
│   ├── middleware/
│   │   ├── auth.ts
│   │   ├── guest.ts
│   │   └── superadmin.ts
│   ├── pages/
│   ├── stores/
│   │   ├── auth.ts
│   │   └── notifications.ts
│   ├── types/
│   ├── nuxt.config.ts
│   └── tailwind.config.ts
│
├── backend/                   # Golang application
│   ├── cmd/
│   │   ├── api/               # main entry (HTTP server)
│   │   └── worker/            # Asynq worker entry
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/
│   │   │   │   ├── auth.go
│   │   │   │   ├── profile.go
│   │   │   │   ├── users.go
│   │   │   │   └── notifications.go
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   ├── ws/
│   │   │   └── notifications.go
│   │   ├── core/
│   │   │   ├── config.go
│   │   │   ├── security.go
│   │   │   ├── database.go
│   │   │   └── redis.go
│   │   ├── model/
│   │   ├── repository/        # sqlc / db layer
│   │   ├── service/
│   │   │   ├── auth_service.go
│   │   │   ├── totp_service.go
│   │   │   ├── user_service.go
│   │   │   ├── email_service.go
│   │   │   ├── notification_service.go
│   │   │   └── storage_service.go
│   │   ├── jobs/
│   │   │   ├── email_jobs.go
│   │   │   ├── notification_jobs.go
│   │   │   └── maintenance_jobs.go
│   │   ├── templates/
│   │   │   └── email/
│   ├── migrations/            # SQL migrations (golang-migrate / atlas)
│   ├── go.mod
│   └── Dockerfile
│
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
└── PRD.md
```

---

## 10. Environment Variables

### Backend
```env
# App
APP_NAME=MyApp
APP_ENV=development
SECRET_KEY=your-secret-key
TOTP_ENCRYPTION_KEY=your-fernet-key   # for encrypting TOTP secrets at rest
FRONTEND_URL=http://localhost:3000

# Database
DATABASE_URL=postgres://user:pass@db:5432/dbname?sslmode=disable

# JWT
ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_DAYS=7
MFA_CHALLENGE_TOKEN_EXPIRE_MINUTES=5

# Redis
REDIS_URL=redis://redis:6379

# Email
MAIL_PROVIDER=smtp          # smtp | sendgrid | resend
MAIL_FROM=noreply@myapp.com
MAIL_FROM_NAME=MyApp
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SENDGRID_API_KEY=
RESEND_API_KEY=

# Storage
STORAGE_BACKEND=local       # local | s3
STORAGE_PATH=./uploads
S3_ENDPOINT_URL=            # e.g. https://s3.amazonaws.com or MinIO URL
S3_BUCKET_NAME=
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_REGION=
S3_PUBLIC_URL=              # CDN or public base URL (optional)
```

### Frontend
```env
NUXT_PUBLIC_API_BASE=http://localhost:8000/api/v1
NUXT_PUBLIC_WS_BASE=ws://localhost:8000
```

---

## 11. Out of Scope (v1.0)

- OAuth / Social login (Google, GitHub, etc.) — Phase 2
- Multi-tenancy
- Audit log UI — Phase 2
- Admin impersonation — Phase 2
- Push notifications (FCM / APNs) — Phase 2
- SMS-based 2FA — Phase 2 (TOTP only in v1.0)

---

## 12. Success Criteria

- [ ] All auth flows (register, verify, login, logout, refresh, forgot/reset password) work end-to-end.
- [ ] 2FA setup, login challenge, disable, and recovery codes work end-to-end.
- [ ] Profile CRUD fully functional including avatar upload to cloud/local storage.
- [ ] Superadmin can list, filter, create, update, activate, ban, and delete users.
- [ ] All defined email templates are dispatched as background jobs at the correct trigger points.
- [ ] Real-time notifications are delivered via WebSocket and persisted in the database.
- [ ] Unread notification count updates live in the UI without page refresh.
- [ ] Background worker processes jobs from Redis and scheduled cron tasks run correctly.
- [ ] File uploads work with both local and S3-compatible backends.
- [ ] Entire stack (frontend, backend, worker, db, redis, minio, mailhog) starts with a single `docker compose up` command.
- [ ] Swagger docs accessible in development mode.
- [ ] Frontend route guards enforce authentication, 2FA, and role requirements correctly.
- [ ] All auth endpoints have rate limiting applied.
- [ ] Database migrations run automatically on startup.
- [ ] Swagger docs available via Go tooling (`swaggo`)
- [ ] Background jobs handled by Asynq worker