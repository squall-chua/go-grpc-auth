package templates

import (
	"strings"
	"testing"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

func TestMFAEmailOTPRenders(t *testing.T) {
	reg := notification.NewTemplateRegistry()
	reg.RegisterEmail(MFAEmailOTP)
	got, err := reg.RenderEmail("ns", "mfa_email_otp", nil, map[string]any{
		"Code": "123456", "TTLMinutes": 5, "AppName": "TestApp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Subject, "TestApp") && !strings.Contains(got.Subject, "verification") {
		t.Errorf("Subject=%q", got.Subject)
	}
	if !strings.Contains(got.HTMLBody, "123456") {
		t.Errorf("HTMLBody missing code: %q", got.HTMLBody)
	}
	if !strings.Contains(got.TextBody, "123456") {
		t.Errorf("TextBody missing code: %q", got.TextBody)
	}
}

func TestMFASMSOTPRenders(t *testing.T) {
	reg := notification.NewTemplateRegistry()
	reg.RegisterSMS(MFASMSOTP)
	got, err := reg.RenderSMS("ns", "mfa_sms_otp", nil, map[string]any{
		"Code": "987654", "TTLMinutes": 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "987654") {
		t.Errorf("body missing code: %q", got)
	}
}
