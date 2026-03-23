CREATE TYPE user_role AS ENUM ('user', 'superadmin');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned');

CREATE TABLE users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email             VARCHAR(255) NOT NULL UNIQUE,
    hashed_password   VARCHAR(255) NOT NULL,
    first_name        VARCHAR(100) NOT NULL DEFAULT '',
    last_name         VARCHAR(100) NOT NULL DEFAULT '',
    display_name      VARCHAR(100) NOT NULL DEFAULT '',
    bio               TEXT,
    avatar_url        VARCHAR(500),
    role              user_role   NOT NULL DEFAULT 'user',
    status            user_status NOT NULL DEFAULT 'inactive',
    is_email_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    totp_secret       VARCHAR(500),
    is_2fa_enabled    BOOLEAN     NOT NULL DEFAULT FALSE,
    last_login_at     TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email      ON users(email);
CREATE INDEX idx_users_status     ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
