package notification

import (
	"context"
	"testing"
)

type stubEmail struct{ name string }

func (s stubEmail) Send(ctx context.Context, msg EmailMessage) error { return nil }

type stubSMS struct{ name string }

func (s stubSMS) Send(ctx context.Context, msg SMSMessage) error { return nil }

func TestRegistryPicksByName(t *testing.T) {
	r := NewRegistry()
	r.RegisterEmail("log", stubEmail{name: "log"})
	r.RegisterEmail("smtp", stubEmail{name: "smtp"})
	r.SetDefaultEmail("smtp")

	got, ok := r.Email("smtp")
	if !ok {
		t.Fatal("expected smtp")
	}
	if _, ok := got.(stubEmail); !ok {
		t.Errorf("type assertion failed")
	}
}

func TestRegistryPickEmailFallsBackToDefault(t *testing.T) {
	r := NewRegistry()
	r.RegisterEmail("log", stubEmail{name: "log"})
	r.SetDefaultEmail("log")

	sender, name, err := r.PickEmail("")
	if err != nil {
		t.Fatal(err)
	}
	if name != "log" {
		t.Errorf("name=%q want log", name)
	}
	if sender == nil {
		t.Fatal("sender nil")
	}
}

func TestRegistryPickEmailUnknownFallsBackToDefault(t *testing.T) {
	r := NewRegistry()
	r.RegisterEmail("log", stubEmail{name: "log"})
	r.SetDefaultEmail("log")

	_, name, err := r.PickEmail("smtp")
	if err != nil {
		t.Fatal(err)
	}
	if name != "log" {
		t.Errorf("name=%q want log", name)
	}
}

func TestRegistryPickEmailNoDefaultReturnsError(t *testing.T) {
	r := NewRegistry()
	_, _, err := r.PickEmail("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistryPickSMS(t *testing.T) {
	r := NewRegistry()
	r.RegisterSMS("sns", stubSMS{name: "sns"})
	r.SetDefaultSMS("sns")

	_, name, err := r.PickSMS("")
	if err != nil {
		t.Fatal(err)
	}
	if name != "sns" {
		t.Errorf("name=%q", name)
	}
}

func TestRegistryPickEmailReturnsPreferredWhenRegistered(t *testing.T) {
	r := NewRegistry()
	r.RegisterEmail("log", stubEmail{name: "log"})
	r.RegisterEmail("smtp", stubEmail{name: "smtp"})
	r.SetDefaultEmail("log")

	sender, name, err := r.PickEmail("smtp")
	if err != nil {
		t.Fatal(err)
	}
	if name != "smtp" {
		t.Errorf("name=%q want smtp", name)
	}
	if sender == nil {
		t.Fatal("sender nil")
	}
}

func TestRegistryPickSMSReturnsPreferredWhenRegistered(t *testing.T) {
	r := NewRegistry()
	r.RegisterSMS("log", stubSMS{name: "log"})
	r.RegisterSMS("sns", stubSMS{name: "sns"})
	r.SetDefaultSMS("log")

	sender, name, err := r.PickSMS("sns")
	if err != nil {
		t.Fatal(err)
	}
	if name != "sns" {
		t.Errorf("name=%q want sns", name)
	}
	if sender == nil {
		t.Fatal("sender nil")
	}
}
