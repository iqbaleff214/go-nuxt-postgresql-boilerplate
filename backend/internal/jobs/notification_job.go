package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeSendNotification = "notification:send"

// SendNotificationPayload is the task payload for pushing a notification to a user.
type SendNotificationPayload struct {
	UserID   string `json:"user_id"`
	NotifID  string `json:"notif_id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

func NewSendNotificationTask(userID, notifID uuid.UUID, notifType, title, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendNotificationPayload{
		UserID:  userID.String(),
		NotifID: notifID.String(),
		Type:    notifType,
		Title:   title,
		Body:    body,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal notification payload: %w", err)
	}
	return asynq.NewTask(TypeSendNotification, payload), nil
}

// NotificationPusher is the minimal interface needed to push to the WebSocket hub.
type NotificationPusher interface {
	BroadcastToUser(ctx context.Context, userID uuid.UUID, payload []byte)
}

// NotificationJobHandler handles TypeSendNotification tasks.
type NotificationJobHandler struct {
	pusher NotificationPusher
}

func NewNotificationJobHandler(pusher NotificationPusher) *NotificationJobHandler {
	return &NotificationJobHandler{pusher: pusher}
}

func (h *NotificationJobHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal notification payload: %w", err)
	}
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id in notification payload: %w", err)
	}

	// Build the WebSocket message envelope
	msg := map[string]any{
		"id":    p.NotifID,
		"type":  p.Type,
		"title": p.Title,
		"body":  p.Body,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal ws message: %w", err)
	}

	h.pusher.BroadcastToUser(ctx, userID, data)
	return nil
}
