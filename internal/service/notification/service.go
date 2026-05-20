package notification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"go.uber.org/zap"
)

// NamespaceNotificationView is the package-local projection of the
// per-namespace notification settings (mirrors domain.NamespaceNotificationConfig).
type NamespaceNotificationView struct {
	EmailProvider  string
	SMSProvider    string
	EmailTemplates map[string]EmailTemplateOverride
	SMSTemplates   map[string]SMSTemplateOverride
}

// NamespaceResolver resolves the notification config for a namespace.
// Implemented by the wiring layer over repository.NamespaceRepository.
type NamespaceResolver interface {
	NotificationConfig(ctx context.Context, namespace string) (NamespaceNotificationView, error)
}

// AuditEmitter receives notification events. Implemented by the wiring
// layer over audit.AuditService.
type AuditEmitter interface {
	LogNotification(ctx context.Context, event, userID, namespace string, metadata any)
}

// Audit event names (mirror the constants added to internal/domain/audit.go).
const (
	eventEmailSent   = "NOTIFICATION_EMAIL_SENT"
	eventEmailFailed = "NOTIFICATION_EMAIL_FAILED"
	eventSMSSent     = "NOTIFICATION_SMS_SENT"
	eventSMSFailed   = "NOTIFICATION_SMS_FAILED"
)

// Service is the facade callers use to send notifications.
type Service interface {
	SendEmail(ctx context.Context, namespace, userID, templateName, to string, data any) error
	SendSMS(ctx context.Context, namespace, userID, templateName, to string, data any) error
}

type service struct {
	reg   *Registry
	tpl   *TemplateRegistry
	ns    NamespaceResolver
	audit AuditEmitter
}

func NewService(reg *Registry, tpl *TemplateRegistry, ns NamespaceResolver, audit AuditEmitter) Service {
	return &service{reg: reg, tpl: tpl, ns: ns, audit: audit}
}

func (s *service) SendEmail(ctx context.Context, namespace, userID, templateName, to string, data any) error {
	start := time.Now()
	nsCfg, err := s.ns.NotificationConfig(ctx, namespace)
	if err != nil {
		zap.L().Warn("notification: namespace resolver error, falling back to defaults",
			zap.String("namespace", namespace),
			zap.Error(err))
		nsCfg = NamespaceNotificationView{}
	}

	sender, providerName, err := s.reg.PickEmail(nsCfg.EmailProvider)
	if err != nil {
		s.auditFailure(ctx, eventEmailFailed, userID, namespace, "", templateName, to, time.Since(start), err)
		return err
	}

	override := overrideFor(nsCfg.EmailTemplates, templateName)
	rendered, err := s.tpl.RenderEmail(namespace, templateName, override, data)
	if err != nil {
		s.auditFailure(ctx, eventEmailFailed, userID, namespace, providerName, templateName, to, time.Since(start), err)
		return err
	}

	msg := EmailMessage{To: to, Subject: rendered.Subject, HTMLBody: rendered.HTMLBody, TextBody: rendered.TextBody}
	err = sender.Send(ctx, msg)
	dur := time.Since(start)
	if err != nil {
		s.auditFailure(ctx, eventEmailFailed, userID, namespace, providerName, templateName, to, dur, err)
		return err
	}
	s.auditSuccess(ctx, eventEmailSent, userID, namespace, providerName, templateName, to, dur)
	return nil
}

func (s *service) SendSMS(ctx context.Context, namespace, userID, templateName, to string, data any) error {
	start := time.Now()
	nsCfg, err := s.ns.NotificationConfig(ctx, namespace)
	if err != nil {
		zap.L().Warn("notification: namespace resolver error, falling back to defaults",
			zap.String("namespace", namespace),
			zap.Error(err))
		nsCfg = NamespaceNotificationView{}
	}

	sender, providerName, err := s.reg.PickSMS(nsCfg.SMSProvider)
	if err != nil {
		s.auditFailure(ctx, eventSMSFailed, userID, namespace, "", templateName, to, time.Since(start), err)
		return err
	}

	override := overrideFor(nsCfg.SMSTemplates, templateName)
	body, err := s.tpl.RenderSMS(namespace, templateName, override, data)
	if err != nil {
		s.auditFailure(ctx, eventSMSFailed, userID, namespace, providerName, templateName, to, time.Since(start), err)
		return err
	}

	msg := SMSMessage{To: to, Body: body}
	err = sender.Send(ctx, msg)
	dur := time.Since(start)
	if err != nil {
		s.auditFailure(ctx, eventSMSFailed, userID, namespace, providerName, templateName, to, dur, err)
		return err
	}
	s.auditSuccess(ctx, eventSMSSent, userID, namespace, providerName, templateName, to, dur)
	return nil
}

func (s *service) auditSuccess(ctx context.Context, event, userID, namespace, provider, template, recipient string, dur time.Duration) {
	meta := s.metadata(provider, template, recipient, dur, nil)
	s.audit.LogNotification(ctx, event, userID, namespace, meta)
	zap.L().Info("notification sent",
		zap.String("event", event),
		zap.String("namespace", namespace),
		zap.String("provider", provider),
		zap.String("template", template),
		zap.String("recipient_hash", recipientHash(recipient)),
		zap.Duration("duration", dur),
	)
}

func (s *service) auditFailure(ctx context.Context, event, userID, namespace, provider, template, recipient string, dur time.Duration, err error) {
	meta := s.metadata(provider, template, recipient, dur, err)
	s.audit.LogNotification(ctx, event, userID, namespace, meta)
	zap.L().Warn("notification failed",
		zap.String("event", event),
		zap.String("namespace", namespace),
		zap.String("provider", provider),
		zap.String("template", template),
		zap.String("recipient_hash", recipientHash(recipient)),
		zap.Duration("duration", dur),
		zap.Error(err),
	)
}

func (s *service) metadata(provider, template, recipient string, dur time.Duration, err error) map[string]any {
	m := map[string]any{
		"provider":       provider,
		"template":       template,
		"recipient_hash": recipientHash(recipient),
		"duration_ms":    dur.Milliseconds(),
	}
	if err != nil {
		m["error"] = err.Error()
	}
	return m
}

func recipientHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func overrideFor[T any](m map[string]T, name string) *T {
	if v, ok := m[name]; ok {
		return &v
	}
	return nil
}
