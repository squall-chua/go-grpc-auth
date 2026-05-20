package smtp

import (
	"context"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	mail "github.com/wneessen/go-mail"
)

type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	UseTLS      bool
}

type Provider struct {
	cfg    Config
	client *mail.Client
}

func New(cfg Config) (*Provider, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("smtp: host required")
	}
	if cfg.FromAddress == "" {
		return nil, fmt.Errorf("smtp: from address required")
	}
	opts := []mail.Option{mail.WithPort(cfg.Port)}
	if cfg.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}
	if cfg.UseTLS {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	}
	c, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("smtp: client: %w", err)
	}
	return &Provider{cfg: cfg, client: c}, nil
}

func (p *Provider) Send(ctx context.Context, msg notification.EmailMessage) error {
	m := mail.NewMsg()
	from := msg.From
	if from == "" {
		from = p.cfg.FromAddress
	}
	if p.cfg.FromName != "" && from == p.cfg.FromAddress {
		if err := m.FromFormat(p.cfg.FromName, from); err != nil {
			return fmt.Errorf("%w: from: %v", notification.ErrInvalidRecipient, err)
		}
	} else {
		if err := m.From(from); err != nil {
			return fmt.Errorf("%w: from: %v", notification.ErrInvalidRecipient, err)
		}
	}
	if err := m.To(msg.To); err != nil {
		return fmt.Errorf("%w: to: %v", notification.ErrInvalidRecipient, err)
	}
	m.Subject(msg.Subject)
	if msg.TextBody != "" {
		m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
		if msg.HTMLBody != "" {
			m.AddAlternativeString(mail.TypeTextHTML, msg.HTMLBody)
		}
	} else if msg.HTMLBody != "" {
		m.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
	}

	if err := p.client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
