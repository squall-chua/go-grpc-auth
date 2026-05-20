package notification

import "context"

type SMSMessage struct {
	To   string // E.164
	From string // optional; provider default applies when empty
	Body string
}

type SMSSender interface {
	Send(ctx context.Context, msg SMSMessage) error
}
