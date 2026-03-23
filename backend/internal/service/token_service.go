package service

import (
	"context"
	"fmt"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenService struct {
	q *repository.Queries
}

func NewTokenService(db *pgxpool.Pool) *TokenService {
	return &TokenService{q: repository.New(db)}
}

// CreateToken generates a random token, stores it hashed, and returns the raw value.
func (s *TokenService) CreateToken(ctx context.Context, userID uuid.UUID, tokenType repository.TokenType, ttl time.Duration) (string, error) {
	raw, err := core.GenerateRandomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	_, err = s.q.CreateToken(ctx, repository.CreateTokenParams{
		UserID:    core.UUIDToPg(userID),
		Token:     core.HashToken(raw),
		Type:      tokenType,
		ExpiresAt: core.TimeToPg(time.Now().Add(ttl)),
	})
	if err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	return raw, nil
}

// ValidateToken looks up a raw token by hash, verifies type/expiry/single-use.
// Does NOT consume the token — callers must call ConsumeToken explicitly.
func (s *TokenService) ValidateToken(ctx context.Context, raw string, tokenType repository.TokenType) (*repository.Token, error) {
	t, err := s.q.GetTokenByHash(ctx, repository.GetTokenByHashParams{
		Token: core.HashToken(raw),
		Type:  tokenType,
	})
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}
	if t.UsedAt.Valid {
		return nil, fmt.Errorf("token already used")
	}
	if time.Now().After(t.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}
	return &t, nil
}

// ConsumeToken marks a token as used.
func (s *TokenService) ConsumeToken(ctx context.Context, tokenID uuid.UUID) error {
	_, err := s.q.MarkTokenUsed(ctx, core.UUIDToPg(tokenID))
	return err
}

// RevokeUserTokensByType deletes all tokens of a given type for a user.
func (s *TokenService) RevokeUserTokensByType(ctx context.Context, userID uuid.UUID, tokenType repository.TokenType) error {
	return s.q.DeleteTokensByUserAndType(ctx, repository.DeleteTokensByUserAndTypeParams{
		UserID: core.UUIDToPg(userID),
		Type:   tokenType,
	})
}
