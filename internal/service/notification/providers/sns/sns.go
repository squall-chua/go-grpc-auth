package sns

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

type Config struct {
	Region         string
	SenderID       string // optional
	AccessKeyID    string // optional; uses default credential chain when empty
	SecretAccessKey string
}

type snsAPI interface {
	Publish(ctx context.Context, in *sns.PublishInput, opts ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type Provider struct {
	api snsAPI
	cfg Config
}

// New constructs a Provider. Uses static credentials when AccessKeyID is
// provided; otherwise falls back to the default AWS credential chain.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("sns: region required")
	}
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("sns: load AWS config: %w", err)
	}
	return &Provider{api: sns.NewFromConfig(awsCfg), cfg: cfg}, nil
}

// NewWithAPI is exposed for tests.
func NewWithAPI(api snsAPI, cfg Config) *Provider {
	return &Provider{api: api, cfg: cfg}
}

// Send delivers an SMS via Amazon SNS.
func (p *Provider) Send(ctx context.Context, msg notification.SMSMessage) error {
	if msg.To == "" {
		return fmt.Errorf("sns: empty recipient: %w", notification.ErrInvalidRecipient)
	}
	in := &sns.PublishInput{
		PhoneNumber: &msg.To,
		Message:     &msg.Body,
	}
	if p.cfg.SenderID != "" {
		in.MessageAttributes = map[string]snstypes.MessageAttributeValue{
			"AWS.SNS.SMS.SenderID": {DataType: ptr("String"), StringValue: &p.cfg.SenderID},
		}
	}
	if _, err := p.api.Publish(ctx, in); err != nil {
		return classifyError(err)
	}
	return nil
}

func ptr(s string) *string { return &s }

// classifyError maps known AWS error codes to package sentinels.
func classifyError(err error) error {
	type apiErr interface{ ErrorCode() string }
	if ae, ok := err.(apiErr); ok {
		switch ae.ErrorCode() {
		case "Throttled", "ThrottledException":
			return fmt.Errorf("sns throttled: %w", notification.ErrRateLimited)
		case "InvalidParameter", "InvalidParameterValue":
			return fmt.Errorf("sns invalid: %w", notification.ErrInvalidRecipient)
		}
	}
	return fmt.Errorf("sns publish: %w", err)
}
