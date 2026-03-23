package service

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/resendlabs/resend-go"
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
)

// EmailSender is the abstraction all email backends implement.
type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

// ─── SMTP ────────────────────────────────────────────────────────────────────

type smtpSender struct {
	host     string
	port     string
	user     string
	pass     string
	fromAddr string
	fromName string
}

func (s *smtpSender) Send(_ context.Context, to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddr)
	msg := strings.Join([]string{
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"",
		htmlBody,
	}, "\r\n")
	addr := s.host + ":" + s.port
	return smtp.SendMail(addr, auth, s.fromAddr, []string{to}, []byte(msg))
}

// ─── SendGrid ────────────────────────────────────────────────────────────────

type sendgridSender struct {
	apiKey   string
	fromAddr string
	fromName string
}

func (s *sendgridSender) Send(_ context.Context, to, subject, htmlBody string) error {
	from := sgmail.NewEmail(s.fromName, s.fromAddr)
	toEmail := sgmail.NewEmail("", to)
	message := sgmail.NewSingleEmail(from, subject, toEmail, "", htmlBody)
	client := sendgrid.NewSendClient(s.apiKey)
	resp, err := client.Send(message)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

// ─── Resend ──────────────────────────────────────────────────────────────────

type resendSender struct {
	client   *resend.Client
	fromAddr string
	fromName string
}

func (s *resendSender) Send(_ context.Context, to, subject, htmlBody string) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddr)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}
	_, err := s.client.Emails.Send(params)
	return err
}

// ─── Factory ─────────────────────────────────────────────────────────────────

// NewEmailSender selects the correct backend from cfg.MailProvider.
func NewEmailSender(cfg *core.Config) EmailSender {
	switch cfg.MailProvider {
	case "sendgrid":
		return &sendgridSender{
			apiKey:   cfg.SendGridAPIKey,
			fromAddr: cfg.MailFrom,
			fromName: cfg.MailFromName,
		}
	case "resend":
		return &resendSender{
			client:   resend.NewClient(cfg.ResendAPIKey),
			fromAddr: cfg.MailFrom,
			fromName: cfg.MailFromName,
		}
	default: // smtp
		return &smtpSender{
			host:     cfg.SMTPHost,
			port:     cfg.SMTPPort,
			user:     cfg.SMTPUser,
			pass:     cfg.SMTPPass,
			fromAddr: cfg.MailFrom,
			fromName: cfg.MailFromName,
		}
	}
}
