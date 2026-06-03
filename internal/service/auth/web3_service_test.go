package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type stubUserRepo struct{}

func (stubUserRepo) Create(context.Context, *struct{}) error { return nil }
// The real *domain.User signature is below; we mirror it so the test file
// compiles once the real repo is wired.

type fakeNonceStore struct{}

func (fakeNonceStore) Save(_ context.Context, _ string, _ string, _ string, _ interface{ Seconds() float64 }) error {
	return nil
}
func (fakeNonceStore) Consume(_ context.Context, _ string, _ string, _ string) (bool, error) {
	return false, nil
}

func TestWeb3AuthService_NotImplementedYet(t *testing.T) {
	// Placeholder test. Real tests are added in Tasks 13-16 as each method
	// is implemented. This exists so the package compiles in TDD order.
	if errors.New("not implemented") == nil {
		t.Fatalf("sanity")
	}
}

func TestRequestNonce_StoresAndReturns(t *testing.T) {
	store := NewMemoryNonceStore()
	svc := NewWeb3AuthService("https://issuer", store, time.Minute, nil, nil, nil, nil)
	addr := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
	got, err := svc.RequestNonce(context.Background(), "default", addr)
	if err != nil {
		t.Fatalf("RequestNonce: %v", err)
	}
	if got == "" {
		t.Fatalf("expected non-empty nonce")
	}
	// Consume should succeed with the same nonce.
	// RequestNonce normalizes wallet to lowercase, so the store key uses lowercase.
	ok, err := store.Consume(context.Background(), "default", strings.ToLower(addr), got)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if !ok {
		t.Fatalf("expected Consume to find the nonce we just issued")
	}
}
