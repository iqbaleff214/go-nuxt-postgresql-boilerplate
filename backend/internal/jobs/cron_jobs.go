package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TypeCleanupExpiredTokens = "cron:cleanup_expired_tokens"
	TypeHardDeleteAccounts   = "cron:hard_delete_accounts"
)

// ─── Cleanup expired tokens ───────────────────────────────────────────────────

type CleanupExpiredTokensHandler struct {
	q *repository.Queries
}

func NewCleanupExpiredTokensHandler(db *pgxpool.Pool) *CleanupExpiredTokensHandler {
	return &CleanupExpiredTokensHandler{q: repository.New(db)}
}

func (h *CleanupExpiredTokensHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	if err := h.q.DeleteExpiredTokens(ctx); err != nil {
		return fmt.Errorf("delete expired tokens: %w", err)
	}
	log.Println("cron: expired tokens cleaned up")
	return nil
}

// ─── Hard delete accounts ─────────────────────────────────────────────────────

type HardDeleteAccountsHandler struct {
	q *repository.Queries
}

func NewHardDeleteAccountsHandler(db *pgxpool.Pool) *HardDeleteAccountsHandler {
	return &HardDeleteAccountsHandler{q: repository.New(db)}
}

func (h *HardDeleteAccountsHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	users, err := h.q.ListUsersScheduledForHardDelete(ctx)
	if err != nil {
		return fmt.Errorf("list users for hard delete: %w", err)
	}
	for _, u := range users {
		if err := h.q.HardDeleteUser(ctx, u.ID); err != nil {
			log.Printf("cron: hard delete user %v: %v", u.ID, err)
		}
	}
	log.Printf("cron: hard deleted %d accounts", len(users))
	return nil
}
