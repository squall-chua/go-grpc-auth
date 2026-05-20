package logp

import (
	"context"
	"sync"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	"go.uber.org/zap"
)

const defaultMaxCapture = 100

// Provider logs sends via zap and optionally captures the last N messages
// for test inspection. In production the capture is bounded so the slice
// cannot grow indefinitely.
type Provider struct {
	mu         sync.Mutex
	emails     []notification.EmailMessage
	sms        []notification.SMSMessage
	maxCapture int
}

// New creates a Provider that logs sends and retains the last 100 messages
// per channel for diagnostic inspection.
func New() *Provider {
	return &Provider{maxCapture: defaultMaxCapture}
}

// Send records and logs an email.
func (p *Provider) Send(ctx context.Context, msg notification.EmailMessage) error {
	p.mu.Lock()
	p.emails = append(p.emails, msg)
	if len(p.emails) > p.maxCapture {
		p.emails = p.emails[len(p.emails)-p.maxCapture:]
	}
	p.mu.Unlock()
	zap.L().Info("logp email", zap.String("to", msg.To), zap.String("subject", msg.Subject))
	return nil
}

// SendSMS records and logs an SMS.
func (p *Provider) SendSMS(ctx context.Context, msg notification.SMSMessage) error {
	p.mu.Lock()
	p.sms = append(p.sms, msg)
	if len(p.sms) > p.maxCapture {
		p.sms = p.sms[len(p.sms)-p.maxCapture:]
	}
	p.mu.Unlock()
	zap.L().Info("logp sms", zap.String("to", msg.To), zap.String("body", msg.Body))
	return nil
}

// SentEmails returns a copy of all captured emails.
func (p *Provider) SentEmails() []notification.EmailMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]notification.EmailMessage, len(p.emails))
	copy(out, p.emails)
	return out
}

// SentSMS returns a copy of all captured SMS messages.
func (p *Provider) SentSMS() []notification.SMSMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]notification.SMSMessage, len(p.sms))
	copy(out, p.sms)
	return out
}

// EmailAdapter exposes Provider as a notification.EmailSender.
type EmailAdapter struct{ P *Provider }

func (a EmailAdapter) Send(ctx context.Context, msg notification.EmailMessage) error {
	return a.P.Send(ctx, msg)
}

// SMSAdapter exposes Provider as a notification.SMSSender.
type SMSAdapter struct{ P *Provider }

func (a SMSAdapter) Send(ctx context.Context, msg notification.SMSMessage) error {
	return a.P.SendSMS(ctx, msg)
}
