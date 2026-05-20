package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeNSResolver struct {
	cfg map[string]NamespaceNotificationView
	err error
}

func (f *fakeNSResolver) NotificationConfig(ctx context.Context, namespace string) (NamespaceNotificationView, error) {
	if f.err != nil {
		return NamespaceNotificationView{}, f.err
	}
	v, ok := f.cfg[namespace]
	if !ok {
		return NamespaceNotificationView{}, nil
	}
	return v, nil
}

type fakeAudit struct {
	events []struct {
		event     string
		userID    string
		namespace string
		metadata  any
	}
}

func (f *fakeAudit) LogNotification(ctx context.Context, event, userID, namespace string, metadata any) {
	f.events = append(f.events, struct {
		event     string
		userID    string
		namespace string
		metadata  any
	}{event, userID, namespace, metadata})
}

type fakeEmailSender struct {
	calls []EmailMessage
	err   error
}

func (f *fakeEmailSender) Send(ctx context.Context, msg EmailMessage) error {
	f.calls = append(f.calls, msg)
	return f.err
}

type fakeSMSSender struct {
	calls []SMSMessage
	err   error
}

func (f *fakeSMSSender) Send(ctx context.Context, msg SMSMessage) error {
	f.calls = append(f.calls, msg)
	return f.err
}

func TestServiceSendEmailUsesDefault(t *testing.T) {
	reg := NewRegistry()
	sender := &fakeEmailSender{}
	reg.RegisterEmail("log", sender)
	reg.SetDefaultEmail("log")

	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("welcome", "Hi {{.Name}}", "<p>Hi {{.Name}}</p>", "Hi {{.Name}}"))

	au := &fakeAudit{}
	svc := NewService(reg, tr, &fakeNSResolver{}, au)
	err := svc.SendEmail(context.Background(), "ns1", "user1", "welcome", "a@b", map[string]any{"Name": "X"})
	if err != nil {
		t.Fatal(err)
	}

	if len(sender.calls) != 1 {
		t.Fatalf("calls=%d", len(sender.calls))
	}
	if sender.calls[0].To != "a@b" || !strings.Contains(sender.calls[0].HTMLBody, "Hi X") {
		t.Errorf("bad message: %+v", sender.calls[0])
	}
	if len(au.events) != 1 || au.events[0].event != "NOTIFICATION_EMAIL_SENT" {
		t.Errorf("audit events: %+v", au.events)
	}
}

func TestServiceSendEmailNamespaceOverridesProvider(t *testing.T) {
	reg := NewRegistry()
	defaultSender := &fakeEmailSender{}
	otherSender := &fakeEmailSender{}
	reg.RegisterEmail("log", defaultSender)
	reg.RegisterEmail("smtp", otherSender)
	reg.SetDefaultEmail("log")

	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("welcome", "S", "<p>x</p>", ""))

	resolver := &fakeNSResolver{cfg: map[string]NamespaceNotificationView{
		"ns1": {EmailProvider: "smtp"},
	}}
	svc := NewService(reg, tr, resolver, &fakeAudit{})

	if err := svc.SendEmail(context.Background(), "ns1", "u", "welcome", "a@b", nil); err != nil {
		t.Fatal(err)
	}
	if len(otherSender.calls) != 1 || len(defaultSender.calls) != 0 {
		t.Errorf("expected smtp picked: other=%d default=%d", len(otherSender.calls), len(defaultSender.calls))
	}
}

func TestServiceSendEmailTemplateNotFound(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterEmail("log", &fakeEmailSender{})
	reg.SetDefaultEmail("log")
	svc := NewService(reg, NewTemplateRegistry(), &fakeNSResolver{}, &fakeAudit{})

	err := svc.SendEmail(context.Background(), "ns", "u", "missing", "a@b", nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("err=%v want ErrTemplateNotFound", err)
	}
}

func TestServiceSendEmailEmitsFailureAudit(t *testing.T) {
	reg := NewRegistry()
	sender := &fakeEmailSender{err: ErrInvalidRecipient}
	reg.RegisterEmail("log", sender)
	reg.SetDefaultEmail("log")

	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("welcome", "s", "<p>x</p>", ""))

	au := &fakeAudit{}
	svc := NewService(reg, tr, &fakeNSResolver{}, au)
	err := svc.SendEmail(context.Background(), "ns", "u", "welcome", "bad", nil)
	if !errors.Is(err, ErrInvalidRecipient) {
		t.Fatalf("err=%v", err)
	}
	if len(au.events) != 1 || au.events[0].event != "NOTIFICATION_EMAIL_FAILED" {
		t.Errorf("audit events: %+v", au.events)
	}
}

func TestServiceSendEmailContinuesOnResolverError(t *testing.T) {
	reg := NewRegistry()
	sender := &fakeEmailSender{}
	reg.RegisterEmail("log", sender)
	reg.SetDefaultEmail("log")

	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("welcome", "S", "<p>x</p>", ""))

	resolver := &fakeNSResolver{err: errors.New("db unavailable")}
	svc := NewService(reg, tr, resolver, &fakeAudit{})

	err := svc.SendEmail(context.Background(), "ns1", "u", "welcome", "a@b", nil)
	if err != nil {
		t.Fatalf("expected fallback to default, got err: %v", err)
	}
	if len(sender.calls) != 1 {
		t.Errorf("expected fallback send, got %d sender calls", len(sender.calls))
	}
}

func TestServiceSendSMS(t *testing.T) {
	reg := NewRegistry()
	sender := &fakeSMSSender{}
	reg.RegisterSMS("log", sender)
	reg.SetDefaultSMS("log")

	tr := NewTemplateRegistry()
	tr.RegisterSMS(MustSMSTemplate("code", "Your code: {{.Code}}"))

	svc := NewService(reg, tr, &fakeNSResolver{}, &fakeAudit{})
	if err := svc.SendSMS(context.Background(), "ns", "u", "code", "+1", map[string]any{"Code": "42"}); err != nil {
		t.Fatal(err)
	}
	if len(sender.calls) != 1 || sender.calls[0].Body != "Your code: 42" {
		t.Errorf("bad SMS: %+v", sender.calls)
	}
}
