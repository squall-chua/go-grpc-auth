package notification

import "context"

type EmailMessage struct {
	To       string
	From     string // optional; provider default applies when empty
	Subject  string
	HTMLBody string
	TextBody string // optional plain-text alternative
}

type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}
