package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// LoginResult is returned by Login.
type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	MFAChallengeToken string
	Requires2FA      bool
}

// AuthService handles all authentication flows.
type AuthService struct {
	cfg    *core.Config
	q      *repository.Queries
	tokens *TokenService
	rdb    *redis.Client
}

func NewAuthService(cfg *core.Config, db *pgxpool.Pool, rdb *redis.Client) *AuthService {
	return &AuthService{
		cfg:    cfg,
		q:      repository.New(db),
		tokens: NewTokenService(db),
		rdb:    rdb,
	}
}

// ─── Registration ─────────────────────────────────────────────────────────────

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	if !emailRe.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	hashed, err := core.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := s.q.CreateUser(ctx, repository.CreateUserParams{
		Email:          email,
		HashedPassword: hashed,
		FirstName:      firstName,
		LastName:       lastName,
		DisplayName:    firstName + " " + lastName,
		Role:           repository.UserRoleUser,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("email already registered")
		}
		return fmt.Errorf("create user: %w", err)
	}

	rawToken, err := s.tokens.CreateToken(ctx, core.PgToUUID(user.ID), repository.TokenTypeEmailVerify, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}

	// TODO Phase 4: enqueue send_email job with rawToken
	_ = rawToken

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	t, err := s.tokens.ValidateToken(ctx, rawToken, repository.TokenTypeEmailVerify)
	if err != nil {
		return fmt.Errorf("invalid or expired verification link")
	}
	if _, err := s.q.SetEmailVerified(ctx, t.UserID); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
}

func (s *AuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil // don't reveal whether email exists
	}
	if user.IsEmailVerified {
		return nil
	}
	_ = s.tokens.RevokeUserTokensByType(ctx, core.PgToUUID(user.ID), repository.TokenTypeEmailVerify)
	rawToken, err := s.tokens.CreateToken(ctx, core.PgToUUID(user.ID), repository.TokenTypeEmailVerify, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	// TODO Phase 4: enqueue send_email job
	_ = rawToken
	return nil
}

// ─── Login ─────────────────────────────────────────────────────────────────────

const maxLoginAttempts = 5
const loginLockoutDuration = 15 * time.Minute

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	lockKey := "login_attempts:" + email
	attempts, _ := s.rdb.Get(ctx, lockKey).Int()
	if attempts >= maxLoginAttempts {
		ttl, _ := s.rdb.TTL(ctx, lockKey).Result()
		return LoginResult{}, fmt.Errorf("account locked, try again in %.0f minutes", ttl.Minutes())
	}

	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil || core.CheckPassword(user.HashedPassword, password) != nil {
		s.rdb.Incr(ctx, lockKey)
		s.rdb.Expire(ctx, lockKey, loginLockoutDuration)
		return LoginResult{}, fmt.Errorf("invalid email or password")
	}

	// Reset attempt counter on success
	s.rdb.Del(ctx, lockKey)

	if user.Status != repository.UserStatusActive {
		return LoginResult{}, fmt.Errorf("account is %s", user.Status)
	}

	// 2FA required
	if user.Is2faEnabled {
		rawChallenge, err := s.tokens.CreateToken(ctx, core.PgToUUID(user.ID), repository.TokenTypeMfaChallenge,
			time.Duration(s.cfg.MFAChallengeExpireMinutes)*time.Minute)
		if err != nil {
			return LoginResult{}, fmt.Errorf("create mfa challenge: %w", err)
		}
		return LoginResult{MFAChallengeToken: rawChallenge, Requires2FA: true}, nil
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) issueTokens(ctx context.Context, user repository.User) (LoginResult, error) {
	userID := core.PgToUUID(user.ID)

	accessToken, err := core.GenerateAccessToken(
		userID.String(), string(user.Role),
		s.cfg.SecretKey, s.cfg.AccessTokenExpireMinutes,
	)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	// Rotate: revoke old refresh tokens, issue new one
	_ = s.tokens.RevokeUserTokensByType(ctx, userID, repository.TokenTypeRefresh)
	rawRefresh, err := s.tokens.CreateToken(ctx, userID, repository.TokenTypeRefresh,
		time.Duration(s.cfg.RefreshTokenExpireDays)*24*time.Hour)
	if err != nil {
		return LoginResult{}, fmt.Errorf("create refresh token: %w", err)
	}

	_, _ = s.q.UpdateLastLogin(ctx, user.ID)
	return LoginResult{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

// ─── Refresh ───────────────────────────────────────────────────────────────────

func (s *AuthService) RefreshToken(ctx context.Context, rawRefresh string) (accessToken, newRefreshToken string, err error) {
	t, err := s.tokens.ValidateToken(ctx, rawRefresh, repository.TokenTypeRefresh)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	user, err := s.q.GetUserByID(ctx, t.UserID)
	if err != nil || user.Status != repository.UserStatusActive {
		return "", "", fmt.Errorf("user not found or inactive")
	}

	// Consume old, issue new (rotation)
	if err := s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID)); err != nil {
		return "", "", err
	}

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		return "", "", err
	}
	return result.AccessToken, result.RefreshToken, nil
}

// ─── Logout ────────────────────────────────────────────────────────────────────

func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	t, err := s.tokens.ValidateToken(ctx, rawRefresh, repository.TokenTypeRefresh)
	if err != nil {
		return nil // already invalid; nothing to do
	}
	return s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
}

// ─── Password reset ─────────────────────────────────────────────────────────────

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil // don't reveal whether email exists
	}
	_ = s.tokens.RevokeUserTokensByType(ctx, core.PgToUUID(user.ID), repository.TokenTypePasswordReset)
	rawToken, err := s.tokens.CreateToken(ctx, core.PgToUUID(user.ID), repository.TokenTypePasswordReset, time.Hour)
	if err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}
	// TODO Phase 4: enqueue send_email job
	_ = rawToken
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	t, err := s.tokens.ValidateToken(ctx, rawToken, repository.TokenTypePasswordReset)
	if err != nil {
		return fmt.Errorf("invalid or expired reset link")
	}
	hashed, err := core.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if _, err := s.q.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID: t.UserID, HashedPassword: hashed,
	}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID)); err != nil {
		return err
	}
	// TODO Phase 4: enqueue "password changed" email
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err := core.CheckPassword(user.HashedPassword, currentPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hashed, err := core.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if _, err := s.q.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID: core.UUIDToPg(userID), HashedPassword: hashed,
	}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	// TODO Phase 4: enqueue notification
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────────

func validatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
