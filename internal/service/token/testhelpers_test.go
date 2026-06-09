package token

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// newTestTokenService returns the unexported *tokenService so tests in this
// package can call package-private methods (e.g. GenerateIDTokenWithClaims).
// Returning the interface would force every claim-injection test to go through
// the full pair-generation path, which depends on a non-nil tokenRepo.
func newTestTokenService(t *testing.T) *tokenService {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return newTokenServiceForTest(priv, "test-kid", "https://issuer.test", nil, nil, nil, nil)
}

func newTestUser() *domain.User {
	return &domain.User{
		ID:        bson.NewObjectID(),
		Email:     "test@example.com",
		Username:  "testuser",
		Namespace: "default",
		Status:    domain.UserStatusActive,
	}
}
