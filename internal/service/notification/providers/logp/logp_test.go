package logp

import (
	"context"
	"testing"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

func TestCapturesEmail(t *testing.T) {
	p := New()
	msg := notification.EmailMessage{To: "a@b", Subject: "s", HTMLBody: "<p>x</p>"}
	if err := p.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	got := p.SentEmails()
	if len(got) != 1 || got[0].To != "a@b" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestCapturesSMS(t *testing.T) {
	p := New()
	msg := notification.SMSMessage{To: "+1", Body: "hi"}
	if err := p.SendSMS(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	got := p.SentSMS()
	if len(got) != 1 || got[0].Body != "hi" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestImplementsBothSenderInterfaces(t *testing.T) {
	p := New()
	var _ notification.EmailSender = EmailAdapter{P: p}
	var _ notification.SMSSender = SMSAdapter{P: p}
}
