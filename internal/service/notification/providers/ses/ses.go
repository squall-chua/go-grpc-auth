package ses

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

// Config holds SES provider configuration.
type Config struct {
	Region         string
	FromAddress    string
	FromName       string
	AccessKeyID    string // optional; uses default credential chain when empty
	SecretAccessKey string
}

// sesAPI is the minimal subset of SESv2 we depend on.
type sesAPI interface {
	SendEmail(ctx context.Context, in *sesv2.SendEmailInput, opts ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// Provider sends email via Amazon SES v2.
type Provider struct {
	api sesAPI
	cfg Config
}

// New constructs a Provider. Uses static credentials when AccessKeyID is
// provided; otherwise falls back to the default AWS credential chain.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("ses: region required")
	}
	if cfg.FromAddress == "" {
		return nil, fmt.Errorf("ses: from address required")
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
		return nil, fmt.Errorf("ses: load AWS config: %w", err)
	}
	return &Provider{api: sesv2.NewFromConfig(awsCfg), cfg: cfg}, nil
}

// NewWithAPI is exposed for tests.
func NewWithAPI(api sesAPI, cfg Config) *Provider {
	return &Provider{api: api, cfg: cfg}
}

// Send delivers a single email message via SES.
func (p *Provider) Send(ctx context.Context, msg notification.EmailMessage) error {
	from := msg.From
	if from == "" {
		from = formatFrom(p.cfg.FromName, p.cfg.FromAddress)
	}
	body := &types.Body{}
	if msg.HTMLBody != "" {
		body.Html = &types.Content{Data: ptr(msg.HTMLBody)}
	}
	if msg.TextBody != "" {
		body.Text = &types.Content{Data: ptr(msg.TextBody)}
	}
	in := &sesv2.SendEmailInput{
		FromEmailAddress: ptr(from),
		Destination:      &types.Destination{ToAddresses: []string{msg.To}},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: ptr(msg.Subject)},
				Body:    body,
			},
		},
	}
	if _, err := p.api.SendEmail(ctx, in); err != nil {
		return classifyError(err)
	}
	return nil
}

func formatFrom(name, addr string) string {
	if name == "" {
		return addr
	}
	return name + " <" + addr + ">"
}

func ptr(s string) *string { return &s }

// classifyError maps known AWS error codes to package sentinels so
// the WithRetry decorator can decide whether to retry.
func classifyError(err error) error {
	type apiErr interface {
		ErrorCode() string
	}
	if ae, ok := err.(apiErr); ok {
		switch ae.ErrorCode() {
		case "ThrottlingException", "TooManyRequestsException":
			return fmt.Errorf("ses throttled: %w", notification.ErrRateLimited)
		case "MailFromDomainNotVerifiedException", "MessageRejected", "AccountSendingPausedException":
			return fmt.Errorf("ses rejected: %w", notification.ErrInvalidRecipient)
		}
	}
	return fmt.Errorf("ses send: %w", err)
}
