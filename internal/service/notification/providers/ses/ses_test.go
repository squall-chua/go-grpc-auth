package ses

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

type fakeSES struct {
	in  *sesv2.SendEmailInput
	out *sesv2.SendEmailOutput
	err error
}

func (f *fakeSES) SendEmail(ctx context.Context, in *sesv2.SendEmailInput, opts ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.in = in
	return f.out, f.err
}

func TestSESProviderSendsEmail(t *testing.T) {
	api := &fakeSES{out: &sesv2.SendEmailOutput{}}
	p := NewWithAPI(api, Config{FromAddress: "noreply@x", FromName: "X"})
	err := p.Send(context.Background(), notification.EmailMessage{
		To:       "a@b",
		Subject:  "Hello",
		HTMLBody: "<p>x</p>",
		TextBody: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.in == nil {
		t.Fatal("SendEmail not invoked")
	}
	if api.in.FromEmailAddress == nil || !strings.Contains(*api.in.FromEmailAddress, "noreply@x") {
		t.Errorf("from = %v", api.in.FromEmailAddress)
	}
	if api.in.Destination == nil || len(api.in.Destination.ToAddresses) != 1 || api.in.Destination.ToAddresses[0] != "a@b" {
		t.Errorf("to = %v", api.in.Destination)
	}
}

func TestSESProviderClassifiesThrottling(t *testing.T) {
	api := &fakeSES{err: &throttleErr{}}
	p := NewWithAPI(api, Config{FromAddress: "noreply@x"})
	err := p.Send(context.Background(), notification.EmailMessage{To: "a@b", Subject: "s", HTMLBody: "<p>x</p>"})
	if !errors.Is(err, notification.ErrRateLimited) {
		t.Fatalf("err=%v want ErrRateLimited", err)
	}
}

// throttleErr mimics a smithy.APIError with ErrorCode "ThrottlingException".
type throttleErr struct{}

func (e *throttleErr) Error() string        { return "ThrottlingException" }
func (e *throttleErr) ErrorCode() string    { return "ThrottlingException" }
func (e *throttleErr) ErrorMessage() string { return "throttled" }
func (e *throttleErr) ErrorFault() int      { return 0 }
