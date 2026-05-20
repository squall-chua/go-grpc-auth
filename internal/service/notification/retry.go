package notification

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
)

// isRetryable returns true for transient failures and unknown errors,
// false for definite client errors and context cancellation/deadline.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, ErrInvalidRecipient) {
		return false
	}
	if errors.Is(err, ErrTemplateNotFound) {
		return false
	}
	// Defensive: if a future provider returns ErrProviderUnavailable from inside Send,
	// retrying would loop pointlessly since the provider itself is down.
	if errors.Is(err, ErrProviderUnavailable) {
		return false
	}
	return true
}

// retryOpts returns the standard retry options for notification delivery.
func retryOpts(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Context(ctx),
		retry.Attempts(3),
		retry.Delay(200 * time.Millisecond),
		retry.MaxDelay(2 * time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(isRetryable),
		retry.LastErrorOnly(true),
	}
}

// WithRetryEmail wraps an EmailSender with the default retry policy.
func WithRetryEmail(inner EmailSender) EmailSender {
	return &retryEmail{inner: inner}
}

type retryEmail struct {
	inner EmailSender
}

func (s *retryEmail) Send(ctx context.Context, msg EmailMessage) error {
	return retry.Do(func() error {
		return s.inner.Send(ctx, msg)
	}, retryOpts(ctx)...)
}

// WithRetrySMS wraps an SMSSender with the default retry policy.
func WithRetrySMS(inner SMSSender) SMSSender {
	return &retrySMS{inner: inner}
}

type retrySMS struct {
	inner SMSSender
}

func (s *retrySMS) Send(ctx context.Context, msg SMSMessage) error {
	return retry.Do(func() error {
		return s.inner.Send(ctx, msg)
	}, retryOpts(ctx)...)
}
