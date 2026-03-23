# Task Breakdown

Based on [PRD.md](PRD.md) v1.2.0. Tasks are grouped into phases intended to be completed in order, as later phases depend on earlier ones.

---

## Phase 0 — Project Scaffolding & Infrastructure

Goal: establish the repository skeleton, Docker environment, and tooling so all subsequent work has a consistent base.

---

### T0.1 — Initialize Backend Go Module

**Location:** `backend/`

**Instructions:**
1. Run `go mod init github.com/<your-org>/go-nuxt-starter-kit/backend` inside `backend/`.
2. Create `backend/cmd/api/main.go` — minimal entry point that starts an HTTP server on `:8080`.
3. Create `backend/cmd/worker/main.go` — minimal entry point (just `fmt.Println("worker starting")` for now).
4. Install core dependencies:
   ```
   go get github.com/gin-gonic/gin
   go get github.com/joho/godotenv
   go get github.com/jackc/pgx/v5
   go get github.com/redis/go-redis/v9
   go get github.com/golang-jwt/jwt/v5
   go get golang.org/x/crypto
   go get github.com/google/uuid
   ```
5. Verify `go build ./...` succeeds.

**Done when:** `go build ./...` in `backend/` compiles without errors.

---

### T0.2 — Initialize Frontend Nuxt Project

**Location:** `frontend/`

**Instructions:**
1. Run `pnpm dlx nuxi@latest init frontend` (or scaffold inside the empty `frontend/` dir with `pnpm dlx nuxi@latest init .`).
2. Select: Nuxt 3, TypeScript, pnpm.
3. Install additional dependencies:
   ```
   pnpm add @pinia/nuxt pinia
   pnpm add @vee-validate/nuxt vee-validate zod
   pnpm add -D tailwindcss @tailwindcss/vite
   ```
4. Install shadcn-vue: follow `shadcn-vue.com/docs/installation/nuxt` — run `pnpm dlx shadcn-vue@latest init`.
5. Configure `nuxt.config.ts`:
   - Add `@pinia/nuxt` and `@vee-validate/nuxt` to `modules`.
   - Set `runtimeConfig.public.apiBase` and `runtimeConfig.public.wsBase` (read from env vars `NUXT_PUBLIC_API_BASE`, `NUXT_PUBLIC_WS_BASE`).
6. Create `tailwind.config.ts` at `frontend/` root following shadcn-vue's recommended config.
7. Verify `pnpm dev` starts the dev server on `:3000`.

**Done when:** `pnpm dev` serves the default Nuxt welcome page on `:3000`.

---

### T0.3 — Create Root Makefile

**Location:** `/Makefile` (repo root)

**Instructions:**
Create a `Makefile` with the following targets:
```makefile
dev:          # docker compose -f docker-compose.dev.yml up --build
down:         # docker compose -f docker-compose.dev.yml down
prod:         # docker compose up --build
backend-build: # cd backend && go build -o api ./cmd/api && go build -o worker ./cmd/worker
backend-test:  # cd backend && go test ./...
backend-lint:  # cd backend && golangci-lint run
fe-install:    # cd frontend && pnpm install
fe-dev:        # cd frontend && pnpm dev
fe-build:      # cd frontend && pnpm build
fe-lint:       # cd frontend && pnpm lint
migrate-up:    # migrate -path ./backend/migrations -database $$DATABASE_URL up
migrate-down:  # migrate -path ./backend/migrations -database $$DATABASE_URL down 1
```

**Done when:** `make dev` is a valid command (even if Docker isn't yet fully configured).

---

### T0.4 — Docker Compose Setup

**Location:** `/docker-compose.dev.yml` and `/docker-compose.yml`

**Instructions:**

Create `docker-compose.dev.yml` with the following services (all on a shared `app-network` bridge network):

| Service | Image | Ports | Volumes / Notes |
|---|---|---|---|
| `frontend` | `node:22-alpine` | `3000:3000` | Mount `./frontend` → `/app`; cmd: `pnpm dev` |
| `backend` | `golang:1.22-alpine` | `8080:8080` | Mount `./backend` → `/app`; cmd: `go run ./cmd/api`; `env_file: .env` |
| `worker` | `golang:1.22-alpine` | — | Same mount as backend; cmd: `go run ./cmd/worker`; `env_file: .env` |
| `db` | `postgres:16-alpine` | `5432:5432` | Named volume `pg_data`; env: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` |
| `redis` | `redis:7-alpine` | `6379:6379` | Named volume `redis_data` |
| `minio` | `minio/minio` | `9000:9000`, `9001:9001` | Named volume `minio_data`; cmd: `server /data --console-address ":9001"` |
| `mailhog` | `mailhog/mailhog` | `8025:8025`, `1025:1025` | No volumes |

Add `depends_on` so `backend` and `worker` depend on `db` and `redis`, and `frontend` depends on `backend`.

Add `healthcheck` for `db` (use `pg_isready`) and `redis` (use `redis-cli ping`).

Create `docker-compose.yml` (production) — identical but **omit** `minio` and `mailhog`, remove source-code volume mounts, and use pre-built images instead of `golang:` base with `go run`.

Create `.env.example` at repo root with all variables from PRD.md §10.

**Done when:** `docker compose -f docker-compose.dev.yml up` starts all 7 services without errors.

---

### T0.5 — Backend Directory Structure & Core Config

**Location:** `backend/internal/`

**Instructions:**
1. Create the full directory skeleton (empty files or minimal stubs):
   ```
   backend/internal/
   ├── api/
   │   ├── handler/
   │   ├── middleware/
   │   └── router.go
   ├── core/
   │   ├── config.go
   │   ├── database.go
   │   ├── redis.go
   │   └── security.go
   ├── model/
   ├── repository/
   ├── service/
   ├── jobs/
   ├── ws/
   └── templates/email/
   ```
2. Implement `internal/core/config.go`:
   - Define a `Config` struct with all fields from PRD.md §10 (App, Database, JWT, Redis, Email, Storage sections).
   - Load from environment using `os.Getenv` (call `godotenv.Load()` before reading in non-production envs).
   - Expose a `Load() (*Config, error)` function.
3. Implement `internal/core/database.go`:
   - Connect to PostgreSQL using `pgx/v5` connection pool.
   - Expose `NewPool(dsn string) (*pgxpool.Pool, error)`.
4. Implement `internal/core/redis.go`:
   - Create a Redis client using `go-redis/v9`.
   - Expose `NewRedisClient(url string) *redis.Client`.
5. Update `cmd/api/main.go` to call `config.Load()`, `database.NewPool()`, and `redis.NewRedisClient()` on startup and log fatal if any fails.

**Done when:** Backend starts up and logs successful DB and Redis connections.

---

### T0.6 — Database Migrations Setup

**Location:** `backend/migrations/`

**Instructions:**
1. Install `golang-migrate` CLI: `brew install golang-migrate` (or download binary).
2. Install Go driver: `go get -tags 'postgres' github.com/golang-migrate/migrate/v4`.
3. Create migration runner in `internal/core/database.go`: add a `RunMigrations(dsn, migrationsPath string) error` function that uses `golang-migrate` to apply all pending migrations on startup.
4. Call `RunMigrations` from `cmd/api/main.go` before starting the HTTP server.
5. Create the first migration file pair:
   - `backend/migrations/000001_create_users_table.up.sql`
   - `backend/migrations/000001_create_users_table.down.sql`

   **Up migration content:**
   ```sql
   CREATE TYPE user_role AS ENUM ('user', 'superadmin');
   CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned');

   CREATE TABLE users (
     id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     email            VARCHAR(255) NOT NULL UNIQUE,
     hashed_password  VARCHAR(255) NOT NULL,
     first_name       VARCHAR(100) NOT NULL DEFAULT '',
     last_name        VARCHAR(100) NOT NULL DEFAULT '',
     display_name     VARCHAR(100) NOT NULL DEFAULT '',
     bio              TEXT,
     avatar_url       VARCHAR(500),
     role             user_role NOT NULL DEFAULT 'user',
     status           user_status NOT NULL DEFAULT 'inactive',
     is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
     totp_secret      VARCHAR(500),
     is_2fa_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
     last_login_at    TIMESTAMPTZ,
     deleted_at       TIMESTAMPTZ,
     created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_users_email ON users(email);
   CREATE INDEX idx_users_status ON users(status);
   CREATE INDEX idx_users_deleted_at ON users(deleted_at);
   ```

   **Down migration content:**
   ```sql
   DROP TABLE IF EXISTS users;
   DROP TYPE IF EXISTS user_status;
   DROP TYPE IF EXISTS user_role;
   ```

6. Create migration `000002_create_tokens_table`:
   ```sql
   -- up
   CREATE TYPE token_type AS ENUM (
     'refresh', 'email_verify', 'password_reset',
     'email_change', 'delete_cancel', 'mfa_challenge', 'totp_recovery'
   );

   CREATE TABLE tokens (
     id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     token      VARCHAR(500) NOT NULL,
     type       token_type NOT NULL,
     expires_at TIMESTAMPTZ NOT NULL,
     used_at    TIMESTAMPTZ,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_tokens_user_id ON tokens(user_id);
   CREATE INDEX idx_tokens_type ON tokens(type);
   CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);
   ```

7. Create migration `000003_create_notifications_table`:
   ```sql
   -- up
   CREATE TABLE notifications (
     id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     type       VARCHAR(100) NOT NULL,
     title      VARCHAR(255) NOT NULL,
     body       TEXT,
     read_at    TIMESTAMPTZ,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_notifications_user_id ON notifications(user_id);
   CREATE INDEX idx_notifications_read_at ON notifications(read_at);
   ```

**Done when:** Running `make migrate-up` applies all 3 migrations and the tables exist in the DB.

---

### T0.7 — sqlc Setup

**Location:** `backend/`

**Instructions:**
1. Install sqlc CLI: `brew install sqlc`.
2. Install Go sqlc package: `go get github.com/sqlc-dev/sqlc/cmd/sqlc`.
3. Create `backend/sqlc.yaml`:
   ```yaml
   version: "2"
   sql:
     - engine: "postgresql"
       queries: "internal/repository/queries/"
       schema: "migrations/"
       gen:
         go:
           package: "repository"
           out: "internal/repository"
           emit_json_tags: true
           emit_interface: true
   ```
4. Create `backend/internal/repository/queries/` directory — leave empty for now. Query files will be added in each feature task.
5. Run `sqlc generate` (or `make sqlc`) after adding each query file to regenerate typed Go code.

**Done when:** `sqlc generate` runs without errors (even with empty queries dir).

---

## Phase 1 — Backend: Authentication

Goal: implement all auth endpoints, JWT logic, and password/token utilities.

---

### T1.1 — Security Utilities

**Location:** `backend/internal/core/security.go`

**Instructions:**
Implement the following functions:

1. **`HashPassword(password string) (string, error)`**
   - Use `bcrypt.GenerateFromPassword` with cost `12`.

2. **`CheckPassword(hash, password string) error`**
   - Use `bcrypt.CompareHashAndPassword`.

3. **`GenerateRandomToken(byteLength int) (string, error)`**
   - Use `crypto/rand` to generate random bytes, encode as hex string.

4. **`HashToken(raw string) string`**
   - Use `sha256.Sum256` on the raw token, return hex-encoded hash.
   - All tokens are stored hashed; this is used before inserting and before lookup.

5. **`GenerateAccessToken(userID, role string, secret string, expireMinutes int) (string, error)`**
   - Create a JWT with claims: `sub` (userID), `role`, `iat`, `exp`.
   - Sign with `HS256`.

6. **`ParseAccessToken(tokenStr, secret string) (*Claims, error)`**
   - Parse and validate JWT; return custom `Claims` struct with `UserID` and `Role`.

7. **`EncryptAES(plaintext, key string) (string, error)`** and **`DecryptAES(ciphertext, key string) (string, error)`**
   - Use `crypto/aes` with GCM mode.
   - Key must be 32 bytes (pad/trim if needed, or enforce exact length).
   - Used exclusively for TOTP secrets at rest.

**Done when:** Unit tests for all 7 functions pass.

---

### T1.2 — sqlc Queries: Users & Tokens

**Location:** `backend/internal/repository/queries/`

**Instructions:**
Create `users.sql` and `tokens.sql` with the following named queries (sqlc will generate Go functions from these):

**users.sql:**
```sql
-- name: CreateUser :one
-- name: GetUserByID :one
-- name: GetUserByEmail :one
-- name: UpdateUserProfile :one   (first_name, last_name, display_name, bio, updated_at)
-- name: UpdateUserEmail :one
-- name: UpdateUserPassword :one
-- name: UpdateUserStatus :one
-- name: UpdateUserRole :one
-- name: UpdateUserAvatarURL :one
-- name: SetEmailVerified :one
-- name: Enable2FA :one           (totp_secret, is_2fa_enabled=true)
-- name: Disable2FA :one          (totp_secret=null, is_2fa_enabled=false)
-- name: UpdateLastLogin :one
-- name: SoftDeleteUser :one      (deleted_at=NOW(), status='inactive')
-- name: HardDeleteUser :exec
-- name: ListUsers :many          (with filters: role, status, is_email_verified, search; pagination: limit, offset)
-- name: CountUsers :one          (same filters as ListUsers)
-- name: ListUsersScheduledForHardDelete :many  (deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '30 days')
```

**tokens.sql:**
```sql
-- name: CreateToken :one
-- name: GetTokenByHash :one      (filter by token hash + type)
-- name: MarkTokenUsed :one       (used_at=NOW())
-- name: DeleteTokensByUserAndType :exec
-- name: DeleteExpiredTokens :exec  (expires_at < NOW())
-- name: ListRecoveryTokensByUser :many  (type='totp_recovery', used_at IS NULL)
```

Run `sqlc generate` after writing each file. Fix any SQL syntax errors surfaced by sqlc.

**Done when:** `sqlc generate` produces Go files in `internal/repository/` with typed methods for all queries above.

---

### T1.3 — Token DB Service

**Location:** `backend/internal/service/token_service.go`

**Instructions:**
Create a `TokenService` struct wrapping the sqlc repository. Implement:

1. **`CreateToken(ctx, userID, tokenType string, ttl time.Duration) (rawToken string, err error)`**
   - Generate a random token (`GenerateRandomToken(32)`).
   - Hash it (`HashToken`).
   - Insert via `repository.CreateToken` with `expires_at = now + ttl`.
   - Return the **raw** (unhashed) token to the caller (this is what goes in emails/cookies).

2. **`ValidateToken(ctx, rawToken, tokenType string) (*repository.Token, error)`**
   - Hash the raw token.
   - Fetch via `repository.GetTokenByHash`.
   - Return error if: not found, type mismatch, `expires_at` is past, or `used_at` is not null.
   - Do **not** mark used here — callers must call `ConsumeToken` explicitly.

3. **`ConsumeToken(ctx, tokenID uuid.UUID) error`**
   - Call `repository.MarkTokenUsed`.

4. **`RevokeUserTokensByType(ctx, userID uuid.UUID, tokenType string) error`**
   - Call `repository.DeleteTokensByUserAndType`.

**Done when:** Functions compile and logic is correct (write table-driven unit tests with a mock repository).

---

### T1.4 — Auth Service

**Location:** `backend/internal/service/auth_service.go`

**Instructions:**
Create `AuthService` with a dependency on `TokenService`, `UserRepository`, and `Config`. Implement:

1. **`Register(ctx, email, password, firstName, lastName string) error`**
   - Validate password strength (min 8 chars, ≥1 uppercase, ≥1 digit, ≥1 special char). Return validation error if fails.
   - Hash password with `HashPassword`.
   - Insert user via `repository.CreateUser` with `status='inactive'`, `is_email_verified=false`.
   - Create an `email_verify` token (24-hour TTL) via `TokenService.CreateToken`.
   - Enqueue `send_email` job (pass token raw value — job details in Phase 4). For now, just log.
   - Return error if email already exists (check for unique constraint violation from pgx).

2. **`VerifyEmail(ctx, rawToken string) error`**
   - Validate token via `TokenService.ValidateToken(rawToken, "email_verify")`.
   - Set `is_email_verified=true` and `status='active'` on the user.
   - Consume the token.

3. **`ResendVerificationEmail(ctx, email string) error`**
   - Find user by email. Return generic success (do not reveal whether email exists).
   - If user exists and is not verified: revoke existing `email_verify` tokens, create new one, enqueue email job.

4. **`Login(ctx, email, password string) (result LoginResult, err error)`**
   - Find user by email. Check password. Track failed attempts (store count+timestamp in Redis key `login_attempts:<email>`).
   - If 5+ consecutive failures within 15 min → return lockout error.
   - On success reset attempt counter.
   - If user `status != 'active'` → return appropriate error.
   - If `is_2fa_enabled` → create `mfa_challenge` token (5-min TTL), return `LoginResult{MFAChallengeToken: rawToken}`.
   - Otherwise → issue access + refresh tokens. Update `last_login_at`.
   - Return `LoginResult{AccessToken, RefreshToken}`.

5. **`RefreshToken(ctx, rawRefreshToken string) (accessToken, newRefreshToken string, err error)`**
   - Validate `refresh` token. Consume it (rotation). Revoke old `refresh` tokens for user.
   - Create new `refresh` token (7-day TTL). Generate new access token.

6. **`Logout(ctx, rawRefreshToken string) error`**
   - Validate and consume the refresh token (ignore expiry errors — just delete).

7. **`ForgotPassword(ctx, email string) error`**
   - Find user. Revoke existing `password_reset` tokens. Create new one (1-hour TTL). Enqueue email job. Return generic success regardless.

8. **`ResetPassword(ctx, rawToken, newPassword string) error`**
   - Validate `password_reset` token. Validate password strength. Hash and update password. Consume token. Enqueue "password changed" email job.

9. **`ChangePassword(ctx, userID uuid.UUID, currentPassword, newPassword string) error`**
   - Fetch user. Verify current password. Validate and hash new password. Update. Enqueue notification email.

**Done when:** All service methods compile. Write unit tests using a mock repository for Register, Login (including lockout), and RefreshToken.

---

### T1.5 — Auth HTTP Handlers & Router

**Location:** `backend/internal/api/handler/auth.go` and `backend/internal/api/router.go`

**Instructions:**

**Handlers** (`AuthHandler` struct, one method per endpoint):

| Method | Path | Handler | Auth |
|---|---|---|---|
| POST | `/api/v1/auth/register` | `Register` | Public |
| POST | `/api/v1/auth/verify-email` | `VerifyEmail` | Public |
| POST | `/api/v1/auth/resend-verification` | `ResendVerification` | Public |
| POST | `/api/v1/auth/login` | `Login` | Public |
| POST | `/api/v1/auth/logout` | `Logout` | Auth |
| POST | `/api/v1/auth/refresh` | `Refresh` | Public (reads cookie) |
| POST | `/api/v1/auth/forgot-password` | `ForgotPassword` | Public |
| POST | `/api/v1/auth/reset-password` | `ResetPassword` | Public |
| POST | `/api/v1/auth/change-password` | `ChangePassword` | Auth |

**Request/response rules:**
- Refresh token is set/cleared via `Set-Cookie` header: `refresh_token=<value>; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=604800`.
- All responses use the JSON envelope from PRD.md §5.10.
- `Login` response body includes `access_token` (and optionally `mfa_challenge_token` if 2FA is pending).
- `Refresh` reads cookie `refresh_token`, not request body.

**Router (`router.go`):**
- Initialize Gin engine.
- Apply global middleware: CORS (allow frontend origin from config), request logger, recovery.
- Register all auth routes.
- Expose `SetupRouter(deps) *gin.Engine` function called from `cmd/api/main.go`.

**Done when:** `POST /api/v1/auth/register` returns 201 and creates a user in DB; `POST /api/v1/auth/login` returns access token.

---

### T1.6 — Auth Middleware

**Location:** `backend/internal/api/middleware/auth.go`

**Instructions:**
1. **`RequireAuth(config *core.Config) gin.HandlerFunc`**
   - Extract `Authorization: Bearer <token>` header.
   - Parse and validate JWT using `security.ParseAccessToken`.
   - On failure → 401. On success → set `userID` and `role` in Gin context.

2. **`RequireSuperadmin() gin.HandlerFunc`**
   - Check `role` in context (set by `RequireAuth`). If not `superadmin` → 403.

3. **`RateLimit(key string, limit int, window time.Duration, redisClient *redis.Client) gin.HandlerFunc`**
   - Use a Redis sliding-window counter keyed by `key:<client-ip>`.
   - If count exceeds `limit` within `window` → return 429 with `Retry-After` header.
   - Apply to: `register` (10/min), `login` (10/min), `forgot-password` (5/min), TOTP verify (5/min).

**Done when:** Requests without a valid token to a protected route return 401; superadmin-only routes return 403 for regular users.

---

## Phase 2 — Backend: Two-Factor Authentication (TOTP)

---

### T2.1 — TOTP Service

**Location:** `backend/internal/service/totp_service.go`

**Instructions:**
Install: `go get github.com/pquerna/otp/totp`

Implement `TOTPService`:

1. **`GenerateSecret(userEmail string) (secret, otpauthURL, qrDataURI string, err error)`**
   - Use `totp.Generate` with `Issuer=config.AppName`, `AccountName=userEmail`.
   - Return raw secret, the `otpauth://` URI, and a base64 PNG QR code data URI (use `github.com/skip2/go-qrcode` or similar: `go get github.com/skip2/go-qrcode`).

2. **`Verify(secret, code string) bool`**
   - Use `totp.Validate(code, secret)` with a 1-step clock drift tolerance.

3. **`GenerateRecoveryCodes(count int) ([]string, error)`**
   - Generate `count` (8) random 10-character alphanumeric codes.
   - Return raw codes. Caller is responsible for hashing before storage.

**Done when:** `GenerateSecret` returns a valid QR data URI and `Verify` returns true for a code generated from the same secret.

---

### T2.2 — sqlc Queries: Recovery Tokens

Recovery codes are stored as `totp_recovery` rows in the `tokens` table (type already exists in the enum). No new migration needed. The queries from T1.2 (`ListRecoveryTokensByUser`, `CreateToken`, `MarkTokenUsed`) cover this.

**Instructions:**
In `AuthService` or a new `TwoFAService`, implement:

1. **`Setup2FA(ctx, userID uuid.UUID) (otpauthURL, qrDataURI string, err error)`**
   - Generate TOTP secret via `TOTPService.GenerateSecret`.
   - Encrypt the raw secret with `security.EncryptAES(secret, config.TOTPEncryptionKey)`.
   - Store the encrypted secret on the user row via `repository.Enable2FA` (but set `is_2fa_enabled=false` yet — activation happens on confirmation).
   - Return the QR URI and data URI to the frontend.

2. **`Confirm2FA(ctx, userID uuid.UUID, code string) (recoveryCodes []string, err error)`**
   - Fetch user, decrypt TOTP secret, verify code via `TOTPService.Verify`.
   - If invalid → error.
   - Set `is_2fa_enabled=true` on the user.
   - Generate 8 recovery codes. Hash each with bcrypt (cost 10). Store each as a `totp_recovery` token with a very far expiry (e.g., 10 years).
   - Return the raw recovery codes (shown once to the user).

3. **`Disable2FA(ctx, userID uuid.UUID, password, code string) error`**
   - Verify password and TOTP code. Clear `totp_secret`, set `is_2fa_enabled=false`. Revoke all `totp_recovery` tokens for user. Enqueue "2FA disabled" email + notification.

4. **`RegenerateRecoveryCodes(ctx, userID uuid.UUID, password, code string) ([]string, error)`**
   - Verify password and TOTP code. Revoke existing `totp_recovery` tokens. Generate and store 8 new ones. Return raw codes.

5. **`VerifyMFAChallenge(ctx, rawChallengeToken, codeOrRecovery string) (accessToken, refreshToken string, err error)`**
   - Validate `mfa_challenge` token (5-min TTL, single-use). Fetch user.
   - Try TOTP verify. If fails, try matching against stored recovery codes (bcrypt compare each). If a recovery code matches, mark it used.
   - Track failed TOTP attempts via Redis key `mfa_attempts:<challengeTokenID>`. If ≥ 5 failures → consume (invalidate) the challenge token and return error.
   - On success → consume challenge token, issue access + refresh tokens, update `last_login_at`.

**Done when:** Full 2FA flow works end-to-end via `curl` tests: setup → confirm → login with TOTP code → disable.

---

### T2.3 — 2FA HTTP Handlers

**Location:** `backend/internal/api/handler/auth.go` (extend existing) or `twofa.go`

| Method | Path | Handler | Auth |
|---|---|---|---|
| POST | `/api/v1/auth/2fa/setup` | `Setup2FA` | Auth |
| POST | `/api/v1/auth/2fa/confirm` | `Confirm2FA` | Auth |
| POST | `/api/v1/auth/2fa/disable` | `Disable2FA` | Auth |
| POST | `/api/v1/auth/2fa/verify` | `VerifyMFAChallenge` | Public (uses challenge token) |
| POST | `/api/v1/auth/2fa/recovery-codes/regenerate` | `RegenerateRecoveryCodes` | Auth |

Register routes in `router.go`.

**Done when:** Postman/curl can complete the full 2FA setup-and-login flow.

---

## Phase 3 — Backend: Profile & User Management

---

### T3.1 — Profile Service

**Location:** `backend/internal/service/user_service.go`

**Instructions:**
Implement `UserService` with dependency on sqlc repository, `StorageService` (stub for now — implement in Phase 7), and `TokenService`.

1. **`GetProfile(ctx, userID uuid.UUID) (*repository.User, error)`**
   - Fetch user by ID.

2. **`UpdateProfile(ctx, userID uuid.UUID, firstName, lastName, displayName, bio string) (*repository.User, error)`**
   - Validate field lengths (display_name max 100 chars, bio max 500 chars).
   - Update via `repository.UpdateUserProfile`.

3. **`UploadAvatar(ctx, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (avatarURL string, err error)`**
   - Validate: allowed MIME types (JPEG, PNG, WebP) via `http.DetectContentType`, max size 2MB.
   - If user has an existing `avatar_url`, delete the old file via `StorageService.Delete`.
   - Upload via `StorageService.Upload(file, "avatars/<userID>/<uuid>.<ext>")`.
   - Update `users.avatar_url`. Return new URL.

4. **`RequestEmailChange(ctx, userID uuid.UUID, newEmail string) error`**
   - Validate email format. Check uniqueness.
   - Revoke existing `email_change` tokens. Create new one (24-hour TTL).
   - Enqueue email with verification link sent to **new email**.
   - Enqueue notification email to **old email** informing them of the change request.

5. **`ConfirmEmailChange(ctx, rawToken string) error`**
   - Validate `email_change` token. Fetch user. Update `users.email`. Consume token.

6. **`RequestAccountDeletion(ctx, userID uuid.UUID) error`**
   - Soft-delete: set `deleted_at=NOW()`, `status='inactive'`.
   - Create `delete_cancel` token (30-day TTL).
   - Enqueue confirmation email with cancel link.

7. **`CancelAccountDeletion(ctx, rawToken string) error`**
   - Validate `delete_cancel` token. Clear `deleted_at`, set `status='active'`. Consume token.

**Done when:** All profile endpoints work via curl.

---

### T3.2 — Profile HTTP Handlers

| Method | Path | Handler | Auth |
|---|---|---|---|
| GET | `/api/v1/profile` | `GetProfile` | Auth |
| PATCH | `/api/v1/profile` | `UpdateProfile` | Auth |
| POST | `/api/v1/profile/avatar` | `UploadAvatar` | Auth |
| POST | `/api/v1/profile/email` | `RequestEmailChange` | Auth |
| POST | `/api/v1/profile/email/confirm` | `ConfirmEmailChange` | Public |
| POST | `/api/v1/profile/delete` | `RequestAccountDeletion` | Auth |
| POST | `/api/v1/profile/delete/cancel` | `CancelAccountDeletion` | Public |
| POST | `/api/v1/auth/change-password` | *(already in T1.5)* | Auth |

Register in `router.go`.

**Done when:** Can update profile, upload avatar (saved to local storage for now), and request email change.

---

### T3.3 — Superadmin User Management Service

Extend `UserService` with superadmin-only methods:

1. **`ListUsers(ctx, filters ListUsersFilter, page, pageSize int) ([]repository.User, total int64, err error)`**
   - `ListUsersFilter`: `Role *string`, `Status *string`, `IsEmailVerified *bool`, `Search *string`.
   - Use `repository.ListUsers` and `repository.CountUsers`.

2. **`GetUserByID(ctx, userID uuid.UUID) (*repository.User, error)`**

3. **`AdminCreateUser(ctx, email, firstName, lastName, role string) error`**
   - Generate a random temporary password. Hash it.
   - Insert user with `status='active'`, `is_email_verified=true`.
   - Create a `password_reset` token (72-hour TTL — longer for admin-created accounts).
   - Enqueue "new account created by admin" email with the set-password link.

4. **`AdminUpdateUser(ctx, callerID, targetUserID uuid.UUID, updates AdminUpdateFields) error`**
   - Guard: if target user is `superadmin` and caller is not the same user → forbid role change.
   - Apply updates (name, bio, role, status). If status changes → enqueue notification + email.

5. **`AdminDeleteUser(ctx, callerID, targetUserID uuid.UUID) error`**
   - Guard: cannot delete own account; cannot delete another superadmin.
   - Hard delete immediately (remove row and all related data).

**Done when:** Superadmin can list users with filters via curl.

---

### T3.4 — Superadmin HTTP Handlers

| Method | Path | Handler | Auth |
|---|---|---|---|
| GET | `/api/v1/admin/users` | `ListUsers` | Superadmin |
| GET | `/api/v1/admin/users/:id` | `GetUser` | Superadmin |
| POST | `/api/v1/admin/users` | `CreateUser` | Superadmin |
| PATCH | `/api/v1/admin/users/:id` | `UpdateUser` | Superadmin |
| DELETE | `/api/v1/admin/users/:id` | `DeleteUser` | Superadmin |
| POST | `/api/v1/admin/users/:id/activate` | `ActivateUser` | Superadmin |
| POST | `/api/v1/admin/users/:id/deactivate` | `DeactivateUser` | Superadmin |
| POST | `/api/v1/admin/users/:id/ban` | `BanUser` | Superadmin |
| POST | `/api/v1/admin/users/:id/unban` | `UnbanUser` | Superadmin |
| POST | `/api/v1/admin/announcements` | `BroadcastAnnouncement` | Superadmin |

Register routes in `router.go` under a group with `RequireAuth` + `RequireSuperadmin` middleware.

**Done when:** All endpoints return correct responses. `GET /api/v1/admin/users` returns paginated list.

---

## Phase 4 — Backend: Email System

---

### T4.1 — Email Service

**Location:** `backend/internal/service/email_service.go`

**Instructions:**
Define interface:
```go
type EmailSender interface {
    Send(ctx context.Context, to, subject, htmlBody string) error
}
```

Implement three structs: `SMTPSender`, `SendGridSender`, `ResendSender`.

- **`SMTPSender`**: use `net/smtp` standard library. Build MIME multipart message.
- **`SendGridSender`**: use `go get github.com/sendgrid/sendgrid-go`. Call `mail/send` API.
- **`ResendSender`**: use `go get github.com/resendlabs/resend-go`. Call `emails.Send`.

Create factory `NewEmailSender(config *core.Config) EmailSender` that selects the implementation based on `MAIL_PROVIDER`.

**Done when:** Dev environment sends test email visible in MailHog at `http://localhost:8025`.

---

### T4.2 — Email Templates

**Location:** `backend/internal/templates/email/`

**Instructions:**
Create Go HTML templates (using `html/template`) for each email type from PRD.md §5.5.1. Each template is a separate `.html` file:

| File | Subject |
|---|---|
| `welcome_verify.html` | Welcome — Verify your email |
| `email_verified.html` | Your email has been verified |
| `password_reset.html` | Reset your password |
| `password_changed.html` | Your password was changed |
| `email_change_verify.html` | Confirm your new email address |
| `account_deactivated.html` | Your account has been deactivated |
| `account_banned.html` | Your account has been banned |
| `account_deletion_confirm.html` | Confirm account deletion |
| `new_account_admin.html` | Your account has been created |

Each template receives a data struct via `{{ .FieldName }}`. Define matching Go structs in `templates/email/data.go`.

Create `TemplateRenderer` service with method `Render(templateName string, data interface{}) (string, error)` that parses and executes templates using Go's `html/template`.

**Done when:** `TemplateRenderer.Render("welcome_verify.html", data)` returns a valid HTML string.

---

### T4.3 — Email Job Registration (Asynq)

**Location:** `backend/internal/jobs/email_jobs.go`

**Instructions:**
Install: `go get github.com/hibiken/asynq`

1. Define task type constants:
   ```go
   const TypeSendEmail = "email:send"
   ```

2. Define `SendEmailPayload` struct with fields: `To`, `TemplateName`, `Data` (map or typed struct).

3. **`NewSendEmailTask(payload SendEmailPayload) (*asynq.Task, error)`**
   - JSON-marshal payload into task.

4. **`HandleSendEmailTask(ctx context.Context, t *asynq.Task) error`**
   - Unmarshal payload. Call `TemplateRenderer.Render`. Call `EmailSender.Send`.

5. Register handler in `cmd/worker/main.go`:
   ```go
   mux.HandleFunc(jobs.TypeSendEmail, emailJobHandler.Handle)
   ```

6. Replace the "log only" stubs in `AuthService` with actual `asynq.Client.Enqueue(jobs.NewSendEmailTask(...))` calls.

**Done when:** Register a new user → email appears in MailHog.

---

## Phase 5 — Backend: Real-Time Notifications

---

### T5.1 — Notification Service

**Location:** `backend/internal/service/notification_service.go`

**Instructions:**
Implement `NotificationService`:

1. **`Create(ctx, userID uuid.UUID, notifType, title, body string) (*repository.Notification, error)`**
   - Insert into `notifications` table.
   - Enqueue `send_notification` Asynq job (see T5.3).

2. **`ListForUser(ctx, userID uuid.UUID, page, pageSize int) ([]repository.Notification, int64, error)`**

3. **`MarkRead(ctx, userID, notifID uuid.UUID) error`**
   - Set `read_at=NOW()` only if `user_id` matches (prevent reading other users' notifications).

4. **`MarkAllRead(ctx, userID uuid.UUID) error`**

5. **`UnreadCount(ctx, userID uuid.UUID) (int64, error)`**

Add sqlc queries to `notifications.sql`:
```sql
-- name: CreateNotification :one
-- name: ListNotificationsForUser :many
-- name: CountNotificationsForUser :one
-- name: GetNotificationByID :one
-- name: MarkNotificationRead :one
-- name: MarkAllNotificationsRead :exec
-- name: CountUnreadNotifications :one
```
Run `sqlc generate`.

**Done when:** Notifications are inserted and queryable.

---

### T5.2 — WebSocket Hub

**Location:** `backend/internal/ws/notifications.go`

**Instructions:**
Install: `go get github.com/gorilla/websocket`

1. Define `Client` struct:
   ```go
   type Client struct {
       UserID uuid.UUID
       conn   *websocket.Conn
       send   chan []byte
   }
   ```

2. Define `Hub` struct:
   ```go
   type Hub struct {
       clients    map[uuid.UUID][]*Client  // supports multiple connections per user
       register   chan *Client
       unregister chan *Client
       broadcast  chan Message             // message with UserID target
       redis      *redis.Client
   }
   ```

3. **`Hub.Run()`** — goroutine that:
   - Handles register/unregister.
   - On `broadcast`: finds clients by `UserID`, writes JSON to their `send` channel.
   - Also subscribes to a Redis pub/sub channel `ws:notifications` and re-broadcasts messages from other backend instances.

4. **`ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request)`**
   - Upgrade HTTP to WebSocket using `websocket.Upgrader`.
   - Extract `?token=<access_token>` query param. Validate JWT. Reject if invalid.
   - Register client in hub. Start `writePump` and `readPump` goroutines.
   - `readPump`: reads incoming messages (ignore or handle ping/pong).
   - `writePump`: writes from `client.send` channel to WebSocket.

5. Register route in `router.go`: `GET /ws/notifications` → `ServeWS`.

**Done when:** Can connect to WebSocket with a valid token and receive a test message pushed via the hub.

---

### T5.3 — Notification Job

**Location:** `backend/internal/jobs/notification_jobs.go`

**Instructions:**
1. Define `TypeSendNotification = "notification:send"`.
2. Define `SendNotificationPayload{UserID, Type, Title, Body}`.
3. **`HandleSendNotificationTask`**:
   - Persist notification via `NotificationService.Create` (pass `skipEnqueue=true` flag or restructure to avoid recursion).
   - Push to hub via `hub.broadcast` channel or publish to Redis pub/sub channel `ws:notifications`.
4. Register handler in worker `main.go`.
5. Call `NotificationService.Create` from all relevant places (status change, role change, 2FA disabled, etc. — add these calls in the service methods from Phase 3).

**Done when:** Changing a user's status via admin API delivers a live WebSocket message to the affected user's open browser tab.

---

### T5.4 — Notification HTTP Handlers

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/notifications` | Auth |
| PATCH | `/api/v1/notifications/:id/read` | Auth |
| PATCH | `/api/v1/notifications/read-all` | Auth |
| GET | `/api/v1/notifications/unread-count` | Auth |

Register in `router.go`.

---

## Phase 6 — Backend: Background Worker & Cron

---

### T6.1 — Worker Main Entry

**Location:** `backend/cmd/worker/main.go`

**Instructions:**
Replace the stub with:
1. Load config.
2. Connect to Redis.
3. Create `asynq.Server` with concurrency (e.g., 10).
4. Create `asynq.ServeMux` and register all job handlers from jobs package.
5. Call `server.Run(mux)`.

**Done when:** Worker process starts and logs "asynq: starting processing".

---

### T6.2 — Scheduled Cron Jobs

**Location:** `backend/cmd/worker/main.go` (or a separate `scheduler.go`)

**Instructions:**
Use `asynq.Scheduler`:

1. **`cleanup_expired_tokens`** — daily at 02:00 UTC.
   - Handler calls `repository.DeleteExpiredTokens`.

2. **`hard_delete_accounts`** — daily at 03:00 UTC.
   - Handler calls `repository.ListUsersScheduledForHardDelete`, then for each user calls `repository.HardDeleteUser` (cascades via FK to tokens and notifications).

Register both with the scheduler. Start scheduler in a goroutine alongside the worker server.

**Done when:** Running the worker for 1 minute and triggering manually (via `asynq` CLI or test) confirms both jobs execute.

---

## Phase 7 — Backend: Storage Service

---

### T7.1 — Storage Service Interface & Local Backend

**Location:** `backend/internal/service/storage_service.go`

**Instructions:**
Define interface:
```go
type StorageService interface {
    Upload(ctx context.Context, file io.Reader, path string) (publicURL string, err error)
    Delete(ctx context.Context, path string) error
    GetSignedURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)
}
```

Implement `LocalStorageService`:
- `Upload`: write file bytes to `config.StoragePath/<path>`. Return `/files/<path>` as public URL.
- `Delete`: `os.Remove(config.StoragePath/<path>)`.
- `GetSignedURL`: for local storage, just return the same public URL (signed URLs are S3-only).

Register a static file route in Gin: `router.Static("/files", config.StoragePath)`.

**Done when:** Avatar upload (from T3.1) stores to disk and the returned URL serves the image from Gin.

---

### T7.2 — S3-Compatible Storage Backend

**Location:** `backend/internal/service/storage_service.go`

**Instructions:**
Install: `go get github.com/aws/aws-sdk-go-v2/service/s3 github.com/aws/aws-sdk-go-v2/config`

Implement `S3StorageService`:
- Configure with `S3_ENDPOINT_URL` (allows MinIO), `S3_BUCKET_NAME`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_REGION`.
- `Upload`: use `s3.PutObject`. If `S3_PUBLIC_URL` is set, return `<S3_PUBLIC_URL>/<path>`, else pre-sign a read URL.
- `Delete`: use `s3.DeleteObject`.
- `GetSignedURL`: use `s3.PresignGetObject` with the given `expiresIn`.

Create factory `NewStorageService(config) StorageService` → returns `LocalStorageService` if `STORAGE_BACKEND=local`, else `S3StorageService`.

Update `T3.1`'s `UploadAvatar` to use the injected `StorageService`.

**Done when:** Setting `STORAGE_BACKEND=s3` with MinIO credentials routes uploads to MinIO, visible in the MinIO console at `:9001`.

---

## Phase 8 — Frontend: Foundation

---

### T8.1 — Nuxt Configuration & Global Types

**Location:** `frontend/nuxt.config.ts`, `frontend/types/`

**Instructions:**
1. Configure `nuxt.config.ts`:
   - Modules: `@pinia/nuxt`, `@vee-validate/nuxt`.
   - `runtimeConfig.public`: `{ apiBase: '', wsBase: '' }` (populated from `NUXT_PUBLIC_*` env vars).
   - Enable Tailwind CSS via `@tailwindcss/vite`.
   - Configure `shadcn-vue` components path if needed.

2. Create `frontend/types/api.ts` — TypeScript interfaces matching API response shapes:
   ```ts
   interface ApiResponse<T> { success: boolean; message: string; data: T; meta?: PaginationMeta }
   interface PaginationMeta { page: number; total: number; page_size: number }
   interface ApiError { success: false; message: string; errors?: FieldError[] }
   interface FieldError { field: string; detail: string }
   ```

3. Create `frontend/types/user.ts`:
   ```ts
   interface User { id, email, firstName, lastName, displayName, bio, avatarUrl, role, status, isEmailVerified, is2faEnabled, lastLoginAt, createdAt }
   interface Notification { id, type, title, body, readAt, createdAt }
   ```

**Done when:** TypeScript compiler finds no errors in these files.

---

### T8.2 — `useApi` Composable

**Location:** `frontend/composables/useApi.ts`

**Instructions:**
Implement a typed HTTP client composable:
1. Use Nuxt's `$fetch` (or `useFetch`) under the hood.
2. Automatically inject `Authorization: Bearer <access_token>` header from the auth store.
3. On 401 response: call `authStore.refresh()` to get a new access token, then retry the original request once.
4. If refresh also fails (401) → call `authStore.logout()` and redirect to `/login`.
5. Expose methods: `get<T>(path)`, `post<T>(path, body)`, `patch<T>(path, body)`, `delete<T>(path)`, `upload<T>(path, formData)`.
6. Throw errors using the `ApiError` interface so callers can access `errors[]` for field-level validation feedback.

**Done when:** Can call `useApi().get<User>('/profile')` from a component and receive a typed response.

---

### T8.3 — Auth Store (Pinia)

**Location:** `frontend/stores/auth.ts`

**Instructions:**
Define a Pinia store with:

**State:**
- `accessToken: string | null`
- `user: User | null`
- `mfaChallengeToken: string | null` — set after login when 2FA is pending

**Actions:**
- `login(email, password)` → call `POST /auth/login`. If response has `mfa_challenge_token` → save it, return `{ requires2fa: true }`. Otherwise save `access_token`, call `fetchMe()`.
- `fetchMe()` → call `GET /profile`, set `user`.
- `refresh()` → call `POST /auth/refresh` (cookie-based). Update `accessToken`.
- `logout()` → call `POST /auth/logout`, clear state, redirect to `/login`.
- `verify2fa(code)` → call `POST /auth/2fa/verify` with `mfa_challenge_token` and code. On success, save access token, call `fetchMe()`.

**Persistence:** use `pinia-plugin-persistedstate` for `accessToken` only (not `user`, re-fetch on mount).

Install: `pnpm add pinia-plugin-persistedstate`

**Done when:** Login form stores token and re-fetching page keeps the user logged in.

---

### T8.4 — Notification Store (Pinia)

**Location:** `frontend/stores/notifications.ts`

**Instructions:**
Define a Pinia store with:

**State:**
- `notifications: Notification[]`
- `unreadCount: number`
- `wsConnected: boolean`

**Actions:**
- `fetchUnreadCount()` → `GET /notifications/unread-count`.
- `fetchNotifications(page)` → `GET /notifications`.
- `markRead(id)` → `PATCH /notifications/:id/read`. Decrement `unreadCount`.
- `markAllRead()` → `PATCH /notifications/read-all`. Set `unreadCount = 0`.
- `addFromWebSocket(notification)` → prepend to list, increment `unreadCount`.

**Done when:** Store actions return correct data from the API.

---

### T8.5 — `useNotifications` Composable (WebSocket)

**Location:** `frontend/composables/useNotifications.ts`

**Instructions:**
1. On mount (when user is authenticated), open WebSocket to `${config.public.wsBase}/ws/notifications?token=<accessToken>`.
2. On message: parse JSON, call `notificationStore.addFromWebSocket(notification)`.
3. On close/error: set `wsConnected = false`. Attempt reconnect with exponential backoff (1s, 2s, 4s, max 30s).
4. Before reconnecting, call `authStore.refresh()` to get a fresh access token (the old one may have expired).
5. On unmount (logout): close connection cleanly.
6. Export `{ wsConnected }` as readonly ref.

**Done when:** With a connected user, triggering a status change via the admin API delivers a live notification in the browser without page refresh.

---

### T8.6 — Route Middleware

**Location:** `frontend/middleware/`

Create four middleware files:

1. **`auth.ts`** — if `!authStore.user && !authStore.accessToken` → redirect to `/login`.
2. **`guest.ts`** — if `authStore.accessToken` → redirect to `/dashboard`.
3. **`superadmin.ts`** — if `authStore.user?.role !== 'superadmin'` → redirect to `/dashboard` with a 403 toast.
4. **`mfa.ts`** — if `authStore.mfaChallengeToken` exists (2FA step pending) → redirect to `/login/2fa` unless already there.

**Done when:** Navigating to `/dashboard` without being logged in redirects to `/login`.

---

### T8.7 — Layouts

**Location:** `frontend/layouts/`

1. **`auth.vue`** — centered card layout, no sidebar. Used by login, register, forgot/reset password, verify-email pages.
2. **`default.vue`** — app shell with: top navbar (user avatar, unread notification badge), optional sidebar, `<slot/>`. Initialize `useNotifications` composable here (WebSocket connection lives in the layout). Apply `auth` middleware.
3. **`admin.vue`** — extends `default.vue` or is a separate layout with an admin sidebar. Apply `superadmin` middleware.

**Done when:** Navigating to `/dashboard` shows the default layout with navbar; navigating to `/login` shows the auth layout.

---

## Phase 9 — Frontend: Authentication Pages

---

### T9.1 — Register Page

**Location:** `frontend/pages/register.vue`

**Instructions:**
- Layout: `auth`.
- Middleware: `guest`.
- Form fields: `firstName` (required), `lastName` (required), `email` (required, email format), `password` (required, min 8, pattern), `confirmPassword` (must match `password`).
- Use VeeValidate + Zod schema for validation.
- On submit: call `POST /auth/register`. On success → show "check your email" message. On field errors → display inline under each input using `errors[]` from `ApiError`.

---

### T9.2 — Verify Email Page

**Location:** `frontend/pages/verify-email.vue`

**Instructions:**
- Layout: `auth`.
- On mount: read `?token=` from query params. Call `POST /auth/verify-email` with the token.
- Show loading spinner while pending.
- On success → show "Email verified! You can now log in." with a link to `/login`.
- On error → show error message and a "Resend verification" button that calls `POST /auth/resend-verification`.

---

### T9.3 — Login Page

**Location:** `frontend/pages/login.vue`

**Instructions:**
- Layout: `auth`.
- Middleware: `guest`.
- Form fields: `email`, `password`.
- On submit: call `authStore.login(email, password)`.
  - If `requires2fa` is true → redirect to `/login/2fa`.
  - Otherwise → redirect to `/dashboard`.
- Show lockout message if API returns 429.

---

### T9.4 — 2FA Login Page

**Location:** `frontend/pages/login/2fa.vue`

**Instructions:**
- Layout: `auth`.
- Middleware: redirect to `/login` if `authStore.mfaChallengeToken` is null.
- Show a numeric OTP input (6 digits) and a separate "use recovery code" toggle that shows a text input.
- On submit: call `authStore.verify2fa(code)`. On success → redirect to `/dashboard`. On error → show "Invalid code" and increment attempt counter display (warn at 4 attempts).

---

### T9.5 — Forgot Password & Reset Password Pages

**Location:** `frontend/pages/forgot-password.vue` and `frontend/pages/reset-password.vue`

**Forgot Password:**
- Layout: `auth`. Middleware: `guest`.
- Single email input. On submit: `POST /auth/forgot-password`. Always show success message (do not reveal if email exists).

**Reset Password:**
- Layout: `auth`.
- On mount: read `?token=` from URL.
- Form: `newPassword`, `confirmPassword`. Validate strength.
- On submit: `POST /auth/reset-password` with token + new password. On success → redirect to `/login` with toast.

---

## Phase 10 — Frontend: Profile, Security & Notifications

---

### T10.1 — Profile Page

**Location:** `frontend/pages/profile.vue`

**Instructions:**
- Layout: `default`. Middleware: `auth`.
- Sections:
  1. **Avatar** — display current avatar. Click to open file picker (JPEG/PNG/WebP, max 2MB client-side check). On file select → preview and `POST /profile/avatar` via `useApi().upload`. Show progress. Update `authStore.user.avatarUrl` on success.
  2. **Profile Info** — editable form with `firstName`, `lastName`, `displayName`, `bio`. On submit: `PATCH /profile`. Show inline errors.
  3. **Email** — display current email with an "Change email" button → opens a modal with a new email input. On confirm: `POST /profile/email`. Show "check your new email inbox" message.
  4. **Account Deletion** — "Delete account" button → opens confirmation modal. On confirm: `POST /profile/delete`. Show "An email has been sent to confirm deletion" and log user out.

---

### T10.2 — Security Page

**Location:** `frontend/pages/profile/security.vue`

**Instructions:**
- Layout: `default`. Middleware: `auth`.
- Sections:
  1. **Change Password** — form: `currentPassword`, `newPassword`, `confirmNewPassword`. On submit: `POST /auth/change-password`.
  2. **Two-Factor Authentication** — display 2FA status badge. If disabled: "Enable 2FA" button → navigate to `/profile/security/2fa`. If enabled: "Disable 2FA" button → opens modal asking for password + TOTP code → `POST /auth/2fa/disable`.
  3. **Recovery Codes** (shown only if 2FA enabled) — "Regenerate recovery codes" button → opens confirmation modal (requires password + TOTP code) → `POST /auth/2fa/recovery-codes/regenerate` → shows new codes in a copyable box.

---

### T10.3 — 2FA Setup Wizard

**Location:** `frontend/pages/profile/security/2fa.vue`

**Instructions:**
- Layout: `default`. Middleware: `auth`.
- Multi-step wizard:

  **Step 1 — Setup:**
  - On mount: call `POST /auth/2fa/setup`. Display the returned QR code (`<img :src="qrDataURI" />`). Also show the raw `otpauth://` key for manual entry.
  - "Next" button → go to Step 2.

  **Step 2 — Confirm:**
  - Prompt user to enter the 6-digit code from their authenticator app.
  - On submit: `POST /auth/2fa/confirm` with the code.
  - On success → go to Step 3.

  **Step 3 — Recovery Codes:**
  - Display the 8 recovery codes in a grid.
  - Show warning: "Save these now — they will not be shown again."
  - Provide a "Copy all" button and a "Download as .txt" button.
  - "Done" button → navigate to `/profile/security`.

---

### T10.4 — Notifications Page

**Location:** `frontend/pages/notifications.vue`

**Instructions:**
- Layout: `default`. Middleware: `auth`.
- On mount: call `notificationStore.fetchNotifications(1)`.
- Display a list of notification cards. Unread notifications have a highlighted background.
- "Mark all read" button at top → `notificationStore.markAllRead()`.
- Click a notification → `notificationStore.markRead(id)`, render it as read visually.
- Infinite scroll or pagination to load more.

---

## Phase 11 — Frontend: Admin Pages

---

### T11.1 — User List Page

**Location:** `frontend/pages/admin/users/index.vue`

**Instructions:**
- Layout: `admin`. Middleware: `superadmin`.
- Fetch `GET /admin/users` with query params: `page`, `pageSize=20`, `search`, `role`, `status`, `isEmailVerified`.
- Display a table with columns: Name, Email, Role, Status, Email Verified, Created At, Actions.
- Filters bar: search input (debounced 300ms), role dropdown, status dropdown, email-verified toggle.
- Sortable column headers (pass `sort` and `order` query params).
- Pagination controls.
- Actions column: Edit (link to `/admin/users/:id`), Activate/Deactivate/Ban (inline buttons with confirmation modal), Delete (confirmation modal).

---

### T11.2 — User Detail / Edit Page

**Location:** `frontend/pages/admin/users/[id].vue`

**Instructions:**
- Layout: `admin`. Middleware: `superadmin`.
- On mount: `GET /admin/users/:id`.
- Display full user profile.
- Editable fields: `firstName`, `lastName`, `displayName`, `bio`, `role`, `status`.
- Save button → `PATCH /admin/users/:id`.
- Separate action buttons: Activate, Deactivate, Ban, Unban (each calls the appropriate endpoint with a confirmation dialog).

---

### T11.3 — Create User Page

**Location:** `frontend/pages/admin/users/create.vue`

**Instructions:**
- Layout: `admin`. Middleware: `superadmin`.
- Form: `firstName`, `lastName`, `email`, `role` (dropdown: user / superadmin).
- On submit: `POST /admin/users`. On success → redirect to `/admin/users` with success toast.

---

### T11.4 — System Announcements Page

**Location:** `frontend/pages/admin/announcements.vue`

**Instructions:**
- Layout: `admin`. Middleware: `superadmin`.
- Form: `title` (required), `body` (optional textarea).
- On submit: `POST /admin/announcements`. This broadcasts a `system.announcement` notification to all connected users via the worker job.
- On success → show "Announcement sent to all users" toast.

---

## Phase 12 — Polish, Docs & Final Verification

---

### T12.1 — Swagger/OpenAPI Docs

**Location:** `backend/`

**Instructions:**
1. Install: `go get github.com/swaggo/swag/cmd/swag github.com/swaggo/gin-swagger github.com/swaggo/files`.
2. Add swag annotations to all handler functions (see swaggo docs for format).
3. Run `swag init -g cmd/api/main.go -o docs/`.
4. Register the Swagger UI route in `router.go` (dev only — check `config.AppEnv == "development"`):
   ```go
   router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   ```
5. Add `swag init` to the Makefile as `make docs`.

**Done when:** `GET http://localhost:8080/docs/index.html` renders the Swagger UI with all endpoints documented.

---

### T12.2 — Health Check Endpoints

**Location:** `backend/internal/api/handler/health.go`

**Instructions:**
Add `GET /health` (public):
- Returns 200 with `{ "status": "ok", "db": "ok", "redis": "ok" }`.
- Pings the DB connection pool and Redis client. Returns 503 if either fails.

This endpoint is used by Docker Compose `healthcheck` for the backend service.

**Done when:** `curl localhost:8080/health` returns 200.

---

### T12.3 — Backend Dockerfile

**Location:** `backend/Dockerfile`

**Instructions:**
Multi-stage build:
```dockerfile
# Stage 1: build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api ./cmd/api && go build -o worker ./cmd/worker

# Stage 2: run
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/migrations ./migrations
```

The Docker Compose service uses `CMD ["./api"]` for the backend and `CMD ["./worker"]` for the worker.

**Done when:** `docker build -t backend ./backend` succeeds and `docker run backend ./api` starts the server.

---

### T12.4 — Frontend Dockerfile

**Location:** `frontend/Dockerfile`

**Instructions:**
Multi-stage build:
```dockerfile
# Stage 1: build
FROM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

# Stage 2: run
FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/.output ./.output
CMD ["node", ".output/server/index.mjs"]
```

**Done when:** `docker build -t frontend ./frontend` succeeds.

---

### T12.5 — End-to-End Verification Checklist

Manually verify each item from PRD.md §12 Success Criteria:

- [ ] Register → verify email → login flow works.
- [ ] Login lockout after 5 failed attempts (test with wrong password 5 times).
- [ ] 2FA: setup → confirm → login with TOTP code → disable.
- [ ] 2FA: use recovery code to log in; verify it is consumed (cannot reuse).
- [ ] Profile: edit info, upload avatar, change email (verify via new email link), delete account + cancel deletion.
- [ ] Admin: create user, list with filters, activate/deactivate/ban/unban, delete.
- [ ] Email: verify all 9 templates render in MailHog at the correct trigger.
- [ ] Real-time: open two browser tabs (same user); ban user from admin → both tabs receive notification instantly.
- [ ] Storage: switch `STORAGE_BACKEND=s3` with MinIO → avatar uploads appear in MinIO console.
- [ ] `docker compose -f docker-compose.dev.yml up` starts all 7 services cleanly from a fresh clone.
- [ ] `/docs` renders Swagger UI with all endpoints.
- [ ] Rate limiting: more than 10 login attempts per minute returns 429.
- [ ] Migrations run automatically on backend startup (verify in fresh DB).
- [ ] Worker cron: `cleanup_expired_tokens` and `hard_delete_accounts` are registered and logged by scheduler.
- [ ] Superadmin route guards: regular user cannot access `/admin/*` (returns 403 from API and redirects on frontend).

---

## Phase Summary

| Phase | Scope | Depends On |
|---|---|---|
| 0 | Scaffolding, Docker, DB migrations, sqlc | — |
| 1 | Backend auth (register, login, JWT, tokens) | 0 |
| 2 | Backend TOTP 2FA | 1 |
| 3 | Backend profile & admin user management | 1 |
| 4 | Email service + templates + Asynq email jobs | 1, 3 |
| 5 | Real-time notifications (WebSocket + jobs) | 3, 4 |
| 6 | Background worker & cron jobs | 4, 5 |
| 7 | Storage service (local + S3) | 3 |
| 8 | Frontend foundation (config, stores, middleware, layouts) | 1 |
| 9 | Frontend auth pages | 8 |
| 10 | Frontend profile, security, notifications | 8, 9 |
| 11 | Frontend admin pages | 8, 10 |
| 12 | Swagger, Dockerfiles, final verification | All |
