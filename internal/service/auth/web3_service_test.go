package auth

import (
	"context"
	"errors"
	"testing"
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
