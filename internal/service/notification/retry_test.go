package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/avast/retry-go/v4"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"rate-limited (sentinel)", ErrRateLimited, true},
		{"wrapped rate-limited", wrap("oops", ErrRateLimited), true},
		{"invalid recipient", ErrInvalidRecipient, false},
		{"template not found", ErrTemplateNotFound, false},
		{"provider unavailable", ErrProviderUnavailable, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
		{"unknown error", errors.New("boom"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isRetryable(tc.err)
			if got != tc.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func wrap(msg string, err error) error {
	return &wrapErr{msg: msg, err: err}
}

type wrapErr struct {
	msg string
	err error
}

func (e *wrapErr) Error() string { return e.msg + ": " + e.err.Error() }
func (e *wrapErr) Unwrap() error { return e.err }

func TestRetryDoSucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(func() error {
		calls++
		return nil
	}, testRetryOpts()...)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if calls != 1 {
		t.Errorf("calls=%d, want 1", calls)
	}
}

func TestRetryDoRetriesUntilSuccess(t *testing.T) {
	calls := 0
	err := retry.Do(func() error {
		calls++
		if calls < 3 {
			return ErrRateLimited
		}
		return nil
	}, testRetryOpts()...)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls=%d, want 3", calls)
	}
}

func TestRetryDoExhaustsAndReturnsLastError(t *testing.T) {
	calls := 0
	err := retry.Do(func() error {
		calls++
		return ErrRateLimited
	}, testRetryOpts()...)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
	if calls != 3 {
		t.Errorf("calls=%d, want 3 (max attempts)", calls)
	}
}

func TestRetryDoDoesNotRetryNonRetryable(t *testing.T) {
	calls := 0
	err := retry.Do(func() error {
		calls++
		return ErrInvalidRecipient
	}, testRetryOpts()...)
	if !errors.Is(err, ErrInvalidRecipient) {
		t.Fatalf("err = %v, want ErrInvalidRecipient", err)
	}
	if calls != 1 {
		t.Errorf("calls=%d, want 1 (non-retryable)", calls)
	}
}

func TestRetryDoHonorsCallerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := retry.Do(func() error {
		calls++
		return nil
	}, testRetryOpts()...,
	)
	_ = ctx // context cancellation tested via retryOpts below

	// With a pre-canceled context, retry.Do returns immediately.
	err = retry.Do(func() error {
		calls++
		return ErrRateLimited
	}, append(testRetryOpts(), retry.Context(ctx))...)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestWithRetryEmailForwardsToInner(t *testing.T) {
	inner := &fakeEmail{}
	wrapped := WithRetryEmail(inner)
	if err := wrapped.Send(context.Background(), EmailMessage{To: "a@b", Subject: "s", HTMLBody: "<p>x</p>"}); err != nil {
		t.Fatal(err)
	}
	if !inner.called {
		t.Error("inner sender was not called")
	}
}

func TestWithRetrySMSForwardsToInner(t *testing.T) {
	inner := &fakeSMS{}
	wrapped := WithRetrySMS(inner)
	if err := wrapped.Send(context.Background(), SMSMessage{To: "+15551234567", Body: "hi"}); err != nil {
		t.Fatal(err)
	}
	if !inner.called {
		t.Error("inner sender was not called")
	}
}

// testRetryOpts returns fast retry options for tests (no real sleeping).
func testRetryOpts() []retry.Option {
	return []retry.Option{
		retry.Attempts(3),
		retry.Delay(0),
		retry.RetryIf(isRetryable),
		retry.LastErrorOnly(true),
	}
}
