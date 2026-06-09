package auth

import (
	"context"
	"time"
)

// NonceStore stores SIWE nonces bound to (namespace, wallet).
// Nonces are single-use and have a TTL.
type NonceStore interface {
	// Save persists a nonce for the given (namespace, wallet) with the given TTL.
	Save(ctx context.Context, namespace, wallet, nonce string, ttl time.Duration) error
	// Consume atomically reads-and-deletes the nonce. Returns (true, nil) if the
	// nonce existed and matched; (false, nil) if it didn't (missing, expired, or
	// already consumed).
	Consume(ctx context.Context, namespace, wallet, nonce string) (bool, error)
}
