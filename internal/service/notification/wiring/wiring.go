package wiring

import (
	"context"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/providers/logp"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/providers/ses"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/providers/smtp"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/providers/sns"
)

// BuildConfig is the subset of server config the registry factory needs.
// Mirrors fields from internal/config.Config to avoid importing that package.
type BuildConfig struct {
	DefaultEmailProvider string
	DefaultSMSProvider   string

	SMTPHost        string
	SMTPPort        int
	SMTPUsername    string
	SMTPPassword    string
	SMTPFromAddress string
	SMTPFromName    string
	SMTPUseTLS      bool

	SESRegion          string
	SESFromAddress     string
	SESFromName        string
	SESAccessKeyID     string
	SESSecretAccessKey string

	SNSRegion          string
	SNSSenderID        string
	SNSAccessKeyID     string
	SNSSecretAccessKey string
}

// BuildRegistry constructs a Registry, registering providers whose config
// blocks are populated. The logp provider is always registered.
// Returns the registry and the logp Provider so callers can use it directly
// in tests or for capture-based diagnostics.
func BuildRegistry(ctx context.Context, cfg BuildConfig) (*notification.Registry, *logp.Provider, error) {
	reg := notification.NewRegistry()

	lp := logp.New()
	reg.RegisterEmail("log", notification.WithRetryEmail(logp.EmailAdapter{P: lp}))
	reg.RegisterSMS("log", notification.WithRetrySMS(logp.SMSAdapter{P: lp}))

	if cfg.SMTPHost != "" {
		p, err := smtp.New(smtp.Config{
			Host:        cfg.SMTPHost,
			Port:        cfg.SMTPPort,
			Username:    cfg.SMTPUsername,
			Password:    cfg.SMTPPassword,
			FromAddress: cfg.SMTPFromAddress,
			FromName:    cfg.SMTPFromName,
			UseTLS:      cfg.SMTPUseTLS,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("notification: smtp: %w", err)
		}
		reg.RegisterEmail("smtp", notification.WithRetryEmail(p))
	}

	if cfg.SESRegion != "" {
		p, err := ses.New(ctx, ses.Config{
			Region:          cfg.SESRegion,
			FromAddress:     cfg.SESFromAddress,
			FromName:        cfg.SESFromName,
			AccessKeyID:     cfg.SESAccessKeyID,
			SecretAccessKey: cfg.SESSecretAccessKey,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("notification: ses: %w", err)
		}
		reg.RegisterEmail("ses", notification.WithRetryEmail(p))
	}

	if cfg.SNSRegion != "" {
		p, err := sns.New(ctx, sns.Config{
			Region:          cfg.SNSRegion,
			SenderID:        cfg.SNSSenderID,
			AccessKeyID:     cfg.SNSAccessKeyID,
			SecretAccessKey: cfg.SNSSecretAccessKey,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("notification: sns: %w", err)
		}
		reg.RegisterSMS("sns", notification.WithRetrySMS(p))
	}

	if cfg.DefaultEmailProvider != "" {
		reg.SetDefaultEmail(cfg.DefaultEmailProvider)
	} else {
		reg.SetDefaultEmail("log")
	}
	if cfg.DefaultSMSProvider != "" {
		reg.SetDefaultSMS(cfg.DefaultSMSProvider)
	} else {
		reg.SetDefaultSMS("log")
	}

	return reg, lp, nil
}
