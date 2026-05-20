package notification

import (
	"context"
	"testing"
)

// Compile-time checks: a struct implementing the right method satisfies the interface.
type fakeEmail struct{ called bool }

func (f *fakeEmail) Send(ctx context.Context, msg EmailMessage) error {
	f.called = true
	return nil
}

type fakeSMS struct{ called bool }

func (f *fakeSMS) Send(ctx context.Context, msg SMSMessage) error {
	f.called = true
	return nil
}

func TestEmailSenderInterface(t *testing.T) {
	var s EmailSender = &fakeEmail{}
	if err := s.Send(context.Background(), EmailMessage{To: "a@b", Subject: "s", HTMLBody: "<p>x</p>"}); err != nil {
		t.Fatal(err)
	}
}

func TestSMSSenderInterface(t *testing.T) {
	var s SMSSender = &fakeSMS{}
	if err := s.Send(context.Background(), SMSMessage{To: "+15551234567", Body: "hi"}); err != nil {
		t.Fatal(err)
	}
}
