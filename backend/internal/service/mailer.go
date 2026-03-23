package service

import (
	"context"
	"fmt"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/jobs"
	tmpl "github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/templates"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/templates/email"
	"github.com/hibiken/asynq"
)

// Mailer builds email HTML from templates and enqueues them via Asynq.
type Mailer struct {
	renderer    *tmpl.Renderer
	client      *asynq.Client
	appName     string
	supportURL  string
	frontendURL string
}

func NewMailer(renderer *tmpl.Renderer, client *asynq.Client, appName, frontendURL string) *Mailer {
	return &Mailer{
		renderer:    renderer,
		client:      client,
		appName:     appName,
		supportURL:  frontendURL + "/support",
		frontendURL: frontendURL,
	}
}

func (m *Mailer) enqueue(ctx context.Context, to, templateName string, data tmpl.Data) error {
	subject, html, err := m.renderer.Render(templateName, data)
	if err != nil {
		return fmt.Errorf("render %s: %w", templateName, err)
	}
	task, err := jobs.NewSendEmailTask(to, subject, html)
	if err != nil {
		return err
	}
	_, err = m.client.EnqueueContext(ctx, task)
	return err
}

func (m *Mailer) SendWelcomeVerify(ctx context.Context, to, firstName, rawToken string) error {
	return m.enqueue(ctx, to, "welcome_verify", email.WelcomeVerifyData{
		Base:      tmpl.Base{AppName: m.appName, Subject: "Verify your email address"},
		FirstName: firstName,
		VerifyURL: m.frontendURL + "/verify-email?token=" + rawToken,
	})
}

func (m *Mailer) SendPasswordReset(ctx context.Context, to, firstName, rawToken string) error {
	return m.enqueue(ctx, to, "password_reset", email.PasswordResetData{
		Base:      tmpl.Base{AppName: m.appName, Subject: "Reset your password"},
		FirstName: firstName,
		ResetURL:  m.frontendURL + "/reset-password?token=" + rawToken,
	})
}

func (m *Mailer) SendPasswordChanged(ctx context.Context, to, firstName string) error {
	return m.enqueue(ctx, to, "password_changed", email.PasswordChangedData{
		Base:      tmpl.Base{AppName: m.appName, Subject: "Your password was changed"},
		FirstName: firstName,
		ChangedAt: time.Now().UTC().Format("2006-01-02 15:04 UTC"),
		ResetURL:  m.frontendURL + "/forgot-password",
	})
}

func (m *Mailer) SendEmailChangeVerify(ctx context.Context, toNewEmail, firstName, newEmail, rawToken string) error {
	return m.enqueue(ctx, toNewEmail, "email_change_verify", email.EmailChangeVerifyData{
		Base:       tmpl.Base{AppName: m.appName, Subject: "Confirm your new email address"},
		FirstName:  firstName,
		NewEmail:   newEmail,
		ConfirmURL: m.frontendURL + "/confirm-email?token=" + rawToken,
	})
}

func (m *Mailer) SendAccountDeletionConfirm(ctx context.Context, to, firstName, rawToken string) error {
	return m.enqueue(ctx, to, "account_deletion_confirm", email.AccountDeletionConfirmData{
		Base:       tmpl.Base{AppName: m.appName, Subject: "Account deletion requested"},
		FirstName:  firstName,
		CancelURL:  m.frontendURL + "/cancel-deletion?token=" + rawToken,
		SupportURL: m.supportURL,
	})
}

func (m *Mailer) SendNewAccountAdmin(ctx context.Context, to, firstName, tempPassword, rawToken string) error {
	return m.enqueue(ctx, to, "new_account_admin", email.NewAccountAdminData{
		Base:         tmpl.Base{AppName: m.appName, Subject: "Your account has been created"},
		FirstName:    firstName,
		Email:        to,
		TempPassword: tempPassword,
		LoginURL:     m.frontendURL + "/reset-password?token=" + rawToken,
	})
}

func (m *Mailer) SendAccountDeactivated(ctx context.Context, to, firstName string) error {
	return m.enqueue(ctx, to, "account_deactivated", email.AccountDeactivatedData{
		Base:       tmpl.Base{AppName: m.appName, Subject: "Your account has been deactivated"},
		FirstName:  firstName,
		SupportURL: m.supportURL,
	})
}

func (m *Mailer) SendAccountBanned(ctx context.Context, to, firstName, reason string) error {
	return m.enqueue(ctx, to, "account_banned", email.AccountBannedData{
		Base:       tmpl.Base{AppName: m.appName, Subject: "Your account has been banned"},
		FirstName:  firstName,
		Reason:     reason,
		SupportURL: m.supportURL,
	})
}
