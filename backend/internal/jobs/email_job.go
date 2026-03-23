package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeSendEmail = "email:send"

// Sender is the minimal interface the email job handler needs.
// The concrete implementations live in service/email_service.go and satisfy
// this interface via structural typing.
type Sender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

// SendEmailPayload is the Asynq task payload for sending an email.
type SendEmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// NewSendEmailTask creates a new Asynq task for sending an email.
func NewSendEmailTask(to, subject, html string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailPayload{To: to, Subject: subject, HTML: html})
	if err != nil {
		return nil, fmt.Errorf("marshal send email payload: %w", err)
	}
	return asynq.NewTask(TypeSendEmail, payload), nil
}

// EmailJobHandler handles TypeSendEmail tasks.
type EmailJobHandler struct {
	sender Sender
}

func NewEmailJobHandler(sender Sender) *EmailJobHandler {
	return &EmailJobHandler{sender: sender}
}

func (h *EmailJobHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal send email payload: %w", err)
	}
	if err := h.sender.Send(ctx, p.To, p.Subject, p.HTML); err != nil {
		return fmt.Errorf("send email to %s: %w", p.To, err)
	}
	return nil
}
