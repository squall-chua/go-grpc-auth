package notification

import (
	"errors"
	"strings"
	"testing"
)

func TestRegisterAndRenderEmail(t *testing.T) {
	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("hello", "Hi {{.Name}}", "<p>Hi {{.Name}}</p>", "Hi {{.Name}}"))

	got, err := tr.RenderEmail("ns1", "hello", nil, map[string]any{"Name": "Squall"})
	if err != nil {
		t.Fatalf("RenderEmail: %v", err)
	}
	if got.Subject != "Hi Squall" {
		t.Errorf("Subject=%q", got.Subject)
	}
	if !strings.Contains(got.HTMLBody, "Hi Squall") {
		t.Errorf("HTMLBody=%q", got.HTMLBody)
	}
	if got.TextBody != "Hi Squall" {
		t.Errorf("TextBody=%q", got.TextBody)
	}
}

func TestRenderEmailUnknownTemplate(t *testing.T) {
	tr := NewTemplateRegistry()
	_, err := tr.RenderEmail("ns1", "missing", nil, nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("err = %v, want ErrTemplateNotFound", err)
	}
}

func TestRenderEmailWholeOverride(t *testing.T) {
	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("hello", "Default {{.Name}}", "<p>Default {{.Name}}</p>", "Default {{.Name}}"))

	override := &EmailTemplateOverride{
		Subject:  "Custom {{.Name}}",
		HTMLBody: "<h1>Custom {{.Name}}</h1>",
		TextBody: "Custom {{.Name}}",
	}
	got, err := tr.RenderEmail("ns1", "hello", override, map[string]any{"Name": "X"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "Custom X" || !strings.Contains(got.HTMLBody, "Custom X") || got.TextBody != "Custom X" {
		t.Errorf("override not applied: %+v", got)
	}
}

func TestRenderEmailPartialOverrideFallsBackPerField(t *testing.T) {
	tr := NewTemplateRegistry()
	tr.RegisterEmail(MustEmailTemplate("hello", "Default {{.Name}}", "<p>Default {{.Name}}</p>", "Default {{.Name}}"))

	override := &EmailTemplateOverride{
		Subject: "Custom Subject",
	}
	got, err := tr.RenderEmail("ns1", "hello", override, map[string]any{"Name": "X"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "Custom Subject" {
		t.Errorf("Subject=%q want %q", got.Subject, "Custom Subject")
	}
	if !strings.Contains(got.HTMLBody, "Default X") {
		t.Errorf("HTMLBody should fall back: %q", got.HTMLBody)
	}
	if got.TextBody != "Default X" {
		t.Errorf("TextBody should fall back: %q", got.TextBody)
	}
}

func TestRenderSMS(t *testing.T) {
	tr := NewTemplateRegistry()
	tr.RegisterSMS(MustSMSTemplate("code", "Your code: {{.Code}}"))

	got, err := tr.RenderSMS("ns1", "code", nil, map[string]any{"Code": "123456"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Your code: 123456" {
		t.Errorf("got %q", got)
	}
}

func TestRenderSMSOverride(t *testing.T) {
	tr := NewTemplateRegistry()
	tr.RegisterSMS(MustSMSTemplate("code", "Default {{.Code}}"))

	got, err := tr.RenderSMS("ns1", "code", &SMSTemplateOverride{Body: "Override {{.Code}}"}, map[string]any{"Code": "9"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Override 9" {
		t.Errorf("got %q", got)
	}
}
