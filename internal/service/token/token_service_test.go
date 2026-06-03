package token

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateTokenPairWithClaims_IncludesCustomClaims verifies that the
// claim-injection path used by GenerateTokenPairWithClaims actually embeds
// caller-supplied claims into the signed ID token.
//
// We exercise the underlying GenerateIDTokenWithClaims directly (rather than
// the public GenerateTokenPairWithClaims entry point) because the pair-level
// method also writes to tokenRepo, which the unit test deliberately leaves nil.
// The claim-injection logic itself lives entirely in GenerateIDTokenWithClaims,
// so this is a faithful coverage of the new feature.
func TestGenerateTokenPairWithClaims_IncludesCustomClaims(t *testing.T) {
	svc := newTestTokenService(t)

	user := newTestUser()
	claims := map[string]any{
		"web3_address":  "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
		"web3_chain_id": int64(1),
	}
	idToken, err := svc.GenerateIDTokenWithClaims(context.Background(), user, "test-aud", "test-nonce", claims)
	if err != nil {
		t.Fatalf("GenerateIDTokenWithClaims: %v", err)
	}
	if idToken == "" {
		t.Fatalf("expected ID token to be set")
	}

	parsed, err := jwt.Parse(idToken, func(t *jwt.Token) (any, error) {
		return &svc.privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("parse id token: %v", err)
	}
	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected jwt.MapClaims, got %T", parsed.Claims)
	}
	if got, want := mapClaims["web3_address"], "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"; got != want {
		t.Errorf("web3_address claim: got %v, want %v", got, want)
	}
	// JWT decodes JSON numbers as float64; compare as float64 to avoid a
	// spurious type mismatch on otherwise-equal values.
	if got, want := mapClaims["web3_chain_id"], float64(1); got != want {
		t.Errorf("web3_chain_id claim: got %v, want %v", got, want)
	}
}
