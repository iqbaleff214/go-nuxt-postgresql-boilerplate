package service

import (
	"context"
	"fmt"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/jobs"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationService handles CRUD for notifications and enqueues delivery jobs.
type NotificationService struct {
	q      *repository.Queries
	client *asynq.Client
}

func NewNotificationService(db *pgxpool.Pool, client *asynq.Client) *NotificationService {
	return &NotificationService{
		q:      repository.New(db),
		client: client,
	}
}

// Create inserts a notification and enqueues a WebSocket delivery job.
func (s *NotificationService) Create(ctx context.Context, userID uuid.UUID, notifType, title, body string) (*repository.Notification, error) {
	n, err := s.q.CreateNotification(ctx, repository.CreateNotificationParams{
		UserID: core.UUIDToPg(userID),
		Type:   notifType,
		Title:  title,
		Body:   core.TextToPg(&body),
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	task, err := jobs.NewSendNotificationTask(userID, core.PgToUUID(n.ID), notifType, title, body)
	if err == nil {
		_, _ = s.client.EnqueueContext(ctx, task)
	}
	return &n, nil
}

func (s *NotificationService) ListForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]repository.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)

	items, err := s.q.ListNotificationsForUser(ctx, repository.ListNotificationsForUserParams{
		UserID: core.UUIDToPg(userID),
		Limit:  int32(pageSize),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	total, err := s.q.CountNotificationsForUser(ctx, core.UUIDToPg(userID))
	if err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}
	return items, total, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notifID uuid.UUID) error {
	_, err := s.q.MarkNotificationRead(ctx, repository.MarkNotificationReadParams{
		ID:     core.UUIDToPg(notifID),
		UserID: core.UUIDToPg(userID),
	})
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.q.MarkAllNotificationsRead(ctx, core.UUIDToPg(userID))
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.q.CountUnreadNotifications(ctx, core.UUIDToPg(userID))
}
