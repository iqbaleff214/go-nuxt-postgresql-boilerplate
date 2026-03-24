package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TypeBroadcastAnnouncement = "admin:broadcast_announcement"

type BroadcastAnnouncementPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewBroadcastAnnouncementTask(title, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(BroadcastAnnouncementPayload{Title: title, Body: body})
	if err != nil {
		return nil, fmt.Errorf("marshal broadcast payload: %w", err)
	}
	return asynq.NewTask(TypeBroadcastAnnouncement, payload), nil
}

// BroadcastAnnouncementHandler fans out a notification to all active users.
type BroadcastAnnouncementHandler struct {
	q      *repository.Queries
	pusher NotificationPusher
}

func NewBroadcastAnnouncementHandler(db *pgxpool.Pool, pusher NotificationPusher) *BroadcastAnnouncementHandler {
	return &BroadcastAnnouncementHandler{q: repository.New(db), pusher: pusher}
}

func (h *BroadcastAnnouncementHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p BroadcastAnnouncementPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal broadcast payload: %w", err)
	}

	users, err := h.q.ListActiveUsers(ctx)
	if err != nil {
		return fmt.Errorf("list active users: %w", err)
	}

	msg := map[string]any{
		"type":  "system.announcement",
		"title": p.Title,
		"body":  p.Body,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal ws message: %w", err)
	}

	for _, u := range users {
		h.pusher.BroadcastToUser(ctx, core.PgToUUID(u.ID), data)
	}
	log.Printf("broadcast announcement to %d users", len(users))
	return nil
}
