package service

import (
	"context"
	"fmt"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// TwoFAService handles TOTP 2FA setup, verification, and recovery.
type TwoFAService struct {
	cfg    *core.Config
	q      *repository.Queries
	tokens *TokenService
	totp   *TOTPService
	auth   *AuthService
	rdb    *redis.Client
}

func NewTwoFAService(cfg *core.Config, db *pgxpool.Pool, rdb *redis.Client, auth *AuthService) *TwoFAService {
	return &TwoFAService{
		cfg:    cfg,
		q:      repository.New(db),
		tokens: NewTokenService(db),
		totp:   NewTOTPService(cfg.AppName),
		auth:   auth,
		rdb:    rdb,
	}
}

// Setup2FA generates a TOTP secret, encrypts it, persists it (is_2fa_enabled stays false),
// and returns the QR data URI and otpauth URL for display.
func (s *TwoFAService) Setup2FA(ctx context.Context, userID string) (otpauthURL, qrDataURI string, err error) {
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(mustParseUUID(userID)))
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}
	if user.Is2faEnabled {
		return "", "", fmt.Errorf("2FA is already enabled")
	}

	secret, otpauthURL, qrDataURI, err := s.totp.GenerateSecret(user.Email)
	if err != nil {
		return "", "", err
	}

	encrypted, err := core.EncryptAES(secret, s.cfg.TOTPEncryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("encrypt secret: %w", err)
	}

	encText := core.TextToPg(&encrypted)
	_, err = s.q.Enable2FA(ctx, repository.Enable2FAParams{
		ID:          core.UUIDToPg(mustParseUUID(userID)),
		TotpSecret:  encText,
		Is2faEnabled: false, // not enabled until confirmed
	})
	if err != nil {
		return "", "", fmt.Errorf("save totp secret: %w", err)
	}

	return otpauthURL, qrDataURI, nil
}

// Confirm2FA verifies the user's first TOTP code, activates 2FA, and returns 8 recovery codes.
func (s *TwoFAService) Confirm2FA(ctx context.Context, userID, code string) ([]string, error) {
	uid := mustParseUUID(userID)
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(uid))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.Is2faEnabled {
		return nil, fmt.Errorf("2FA already active")
	}
	if !user.TotpSecret.Valid {
		return nil, fmt.Errorf("2FA setup not started")
	}

	secret, err := core.DecryptAES(user.TotpSecret.String, s.cfg.TOTPEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}
	if !s.totp.Verify(secret, code) {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	// Activate 2FA
	encText := core.TextToPg(&user.TotpSecret.String)
	if _, err := s.q.Enable2FA(ctx, repository.Enable2FAParams{
		ID:          core.UUIDToPg(uid),
		TotpSecret:  encText,
		Is2faEnabled: true,
	}); err != nil {
		return nil, fmt.Errorf("activate 2fa: %w", err)
	}

	// Generate and store recovery codes
	rawCodes, err := s.totp.GenerateRecoveryCodes(8)
	if err != nil {
		return nil, fmt.Errorf("generate recovery codes: %w", err)
	}

	for _, raw := range rawCodes {
		hashed, err := bcrypt.GenerateFromPassword([]byte(raw), 10)
		if err != nil {
			return nil, err
		}
		hashedStr := string(hashed)
		_, err = s.tokens.q.CreateToken(ctx, repository.CreateTokenParams{
			UserID:    core.UUIDToPg(uid),
			Token:     hashedStr,
			Type:      repository.TokenTypeTotpRecovery,
			ExpiresAt: core.TimeToPg(time.Now().Add(10 * 365 * 24 * time.Hour)),
		})
		if err != nil {
			return nil, fmt.Errorf("store recovery code: %w", err)
		}
	}

	// TODO Phase 4: enqueue "2FA enabled" notification
	return rawCodes, nil
}

// Disable2FA verifies credentials, clears the TOTP secret, and revokes recovery codes.
func (s *TwoFAService) Disable2FA(ctx context.Context, userID, password, code string) error {
	uid := mustParseUUID(userID)
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(uid))
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if !user.Is2faEnabled {
		return fmt.Errorf("2FA is not enabled")
	}
	if err := core.CheckPassword(user.HashedPassword, password); err != nil {
		return fmt.Errorf("incorrect password")
	}
	secret, err := core.DecryptAES(user.TotpSecret.String, s.cfg.TOTPEncryptionKey)
	if err != nil {
		return fmt.Errorf("decrypt secret: %w", err)
	}
	if !s.totp.Verify(secret, code) {
		return fmt.Errorf("invalid TOTP code")
	}

	if _, err := s.q.Disable2FA(ctx, core.UUIDToPg(uid)); err != nil {
		return fmt.Errorf("disable 2fa: %w", err)
	}
	_ = s.tokens.RevokeUserTokensByType(ctx, uid, repository.TokenTypeTotpRecovery)

	// TODO Phase 4: enqueue "2FA disabled" email + notification
	return nil
}

// RegenerateRecoveryCodes verifies credentials, revokes old codes, and creates new ones.
func (s *TwoFAService) RegenerateRecoveryCodes(ctx context.Context, userID, password, code string) ([]string, error) {
	uid := mustParseUUID(userID)
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(uid))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if !user.Is2faEnabled {
		return nil, fmt.Errorf("2FA is not enabled")
	}
	if err := core.CheckPassword(user.HashedPassword, password); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}
	secret, err := core.DecryptAES(user.TotpSecret.String, s.cfg.TOTPEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}
	if !s.totp.Verify(secret, code) {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	_ = s.tokens.RevokeUserTokensByType(ctx, uid, repository.TokenTypeTotpRecovery)

	rawCodes, err := s.totp.GenerateRecoveryCodes(8)
	if err != nil {
		return nil, err
	}
	for _, raw := range rawCodes {
		hashed, err := bcrypt.GenerateFromPassword([]byte(raw), 10)
		if err != nil {
			return nil, err
		}
		hashedStr := string(hashed)
		_, err = s.tokens.q.CreateToken(ctx, repository.CreateTokenParams{
			UserID:    core.UUIDToPg(uid),
			Token:     hashedStr,
			Type:      repository.TokenTypeTotpRecovery,
			ExpiresAt: core.TimeToPg(time.Now().Add(10 * 365 * 24 * time.Hour)),
		})
		if err != nil {
			return nil, fmt.Errorf("store recovery code: %w", err)
		}
	}
	return rawCodes, nil
}

// VerifyMFAChallenge completes the 2FA login step and issues full tokens.
func (s *TwoFAService) VerifyMFAChallenge(ctx context.Context, rawChallengeToken, codeOrRecovery string) (accessToken, refreshToken string, err error) {
	t, err := s.tokens.ValidateToken(ctx, rawChallengeToken, repository.TokenTypeMfaChallenge)
	if err != nil {
		return "", "", fmt.Errorf("invalid or expired challenge")
	}

	uid := core.PgToUUID(t.UserID)
	user, err := s.q.GetUserByID(ctx, t.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	attemptsKey := fmt.Sprintf("mfa_attempts:%s", core.PgToUUID(t.ID))
	attempts, _ := s.rdb.Get(ctx, attemptsKey).Int()
	if attempts >= 5 {
		_ = s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
		return "", "", fmt.Errorf("too many failed attempts, please log in again")
	}

	// Try TOTP
	secret, decErr := core.DecryptAES(user.TotpSecret.String, s.cfg.TOTPEncryptionKey)
	verified := decErr == nil && s.totp.Verify(secret, codeOrRecovery)

	// Try recovery code if TOTP failed
	if !verified {
		verified = s.verifyRecoveryCode(ctx, uid, codeOrRecovery)
	}

	if !verified {
		s.rdb.Incr(ctx, attemptsKey)
		s.rdb.Expire(ctx, attemptsKey, 10*time.Minute)
		return "", "", fmt.Errorf("invalid code")
	}

	// Success: consume challenge, clean up attempt counter, issue tokens
	_ = s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
	s.rdb.Del(ctx, attemptsKey)

	result, err := s.auth.issueTokens(ctx, user)
	if err != nil {
		return "", "", err
	}
	return result.AccessToken, result.RefreshToken, nil
}

func (s *TwoFAService) verifyRecoveryCode(ctx context.Context, userID uuid.UUID, code string) bool {
	pgUID := core.UUIDToPg(userID)
	codes, err := s.tokens.q.ListRecoveryTokensByUser(ctx, pgUID)
	if err != nil {
		return false
	}
	for _, rc := range codes {
		if bcrypt.CompareHashAndPassword([]byte(rc.Token), []byte(code)) == nil {
			_ = s.tokens.ConsumeToken(ctx, core.PgToUUID(rc.ID))
			return true
		}
	}
	return false
}
