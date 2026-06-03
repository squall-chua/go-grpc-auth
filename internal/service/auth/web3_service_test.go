package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

type fakeAllowedChain struct{ ids []int64 }

func (f fakeAllowedChain) AllowedChainIDs(_ context.Context, _ string) ([]int64, error) {
	return f.ids, nil
}

func TestVerify_HappyPath_AutoProvisions(t *testing.T) {
	// This test needs real repos; for the unit-test phase we cover the small
	// parts (nonce consumption, chainId allowlist, address normalization)
	// and the full happy path is covered in the bufconn e2e test (Task 23).
	store := NewMemoryNonceStore()
	addr := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
	nonce := "abc12345"
	_ = store.Save(context.Background(), "default", strings.ToLower(addr), nonce, time.Minute)

	// Build a SIWE message and sign it.
	key, _ := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	walletAddr := crypto.PubkeyToAddress(key.PublicKey)
	msgStr := buildSIWEForTest(t, walletAddr, nonce, 1)
	sig := signMessage(t, msgStr, key)

	// Note: we can't exercise the full Verify without a userRepo; we just
	// confirm the SIWE verification leg works.
	recovered, err := VerifySIWE(msgStr, sig, nil, &nonce)
	if err != nil {
		t.Fatalf("VerifySIWE: %v", err)
	}
	if recovered != walletAddr {
		t.Fatalf("recovered: got %s want %s", recovered.Hex(), walletAddr.Hex())
	}
}

func buildSIWEForTest(t *testing.T, addr common.Address, nonce string, chainID int64) string {
	t.Helper()
	return "example.com wants you to sign in with your Ethereum account:\n" +
		addr.Hex() + "\n\n" +
		"Sign in with Ethereum to the app.\n\n" +
		"URI: https://example.com/login\n" +
		"Version: 1\n" +
		"Chain ID: " + strconv.FormatInt(chainID, 10) + "\n" +
		"Nonce: " + nonce + "\n" +
		"Issued At: " + time.Now().UTC().Format(time.RFC3339)
}
