package sns

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

type fakeSNS struct {
	in  *sns.PublishInput
	out *sns.PublishOutput
	err error
}

func (f *fakeSNS) Publish(ctx context.Context, in *sns.PublishInput, opts ...func(*sns.Options)) (*sns.PublishOutput, error) {
	f.in = in
	return f.out, f.err
}

func TestSNSProviderSendsSMS(t *testing.T) {
	api := &fakeSNS{out: &sns.PublishOutput{}}
	p := NewWithAPI(api, Config{SenderID: "MyApp"})
	err := p.Send(context.Background(), notification.SMSMessage{To: "+15551234567", Body: "Code 42"})
	if err != nil {
		t.Fatal(err)
	}
	if api.in == nil || api.in.PhoneNumber == nil || *api.in.PhoneNumber != "+15551234567" {
		t.Errorf("phone = %v", api.in.PhoneNumber)
	}
	if api.in.Message == nil || *api.in.Message != "Code 42" {
		t.Errorf("message = %v", api.in.Message)
	}
}

func TestSNSProviderRejectsEmptyTo(t *testing.T) {
	api := &fakeSNS{}
	p := NewWithAPI(api, Config{})
	err := p.Send(context.Background(), notification.SMSMessage{To: "", Body: "hi"})
	if !errors.Is(err, notification.ErrInvalidRecipient) {
		t.Fatalf("err=%v", err)
	}
}
