package auth

import (
	"crypto/ecdsa"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
)

// Test data: a deterministic EIP-4361 message + signature.
func mustKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	// Hard-coded test key (do not use in production).
	k, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("HexToECDSA: %v", err)
	}
	return k
}

func signMessage(t *testing.T, message string, key *ecdsa.PrivateKey) string {
	t.Helper()
	prefix := "\x19Ethereum Signed Message:\n" + itoa(len(message))
	hash := crypto.Keccak256Hash([]byte(prefix + message))
	sig, err := crypto.Sign(hash.Bytes(), key)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	sig[64] += 27
	return "0x" + common.Bytes2Hex(sig)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func buildSIWE(t *testing.T, address common.Address, nonce string) string {
	t.Helper()
	return "example.com wants you to sign in with your Ethereum account:\n" +
		address.Hex() + "\n\n" +
		"Sign in with Ethereum to the app.\n\n" +
		"URI: https://example.com/login\n" +
		"Version: 1\n" +
		"Chain ID: 1\n" +
		"Nonce: " + nonce + "\n" +
		"Issued At: " + time.Now().UTC().Format(time.RFC3339)
}

func TestSIWE_ParseAndRecoverAddress(t *testing.T) {
	key := mustKey(t)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	nonce := "abc12345"
	msgStr := buildSIWE(t, addr, nonce)
	sig := signMessage(t, msgStr, key)

	msg, err := siwe.ParseMessage(msgStr)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}

	gotAddr := msg.GetAddress()
	if gotAddr != addr {
		t.Fatalf("recovered address mismatch: got %s want %s", gotAddr.Hex(), addr.Hex())
	}

	if _, err := msg.VerifyEIP191(sig); err != nil {
		t.Fatalf("VerifyEIP191: %v", err)
	}
}

func TestSIWE_RejectsTamperedMessage(t *testing.T) {
	key := mustKey(t)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	nonce := "abc12345"
	msgStr := buildSIWE(t, addr, nonce)
	sig := signMessage(t, msgStr, key)

	tampered := strings.Replace(msgStr, "Sign in with Ethereum to the app.", "Sign in to take over.", 1)
	msg, err := siwe.ParseMessage(tampered)
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}

	if _, err := msg.VerifyEIP191(sig); err == nil {
		t.Fatalf("expected VerifyEIP191 to fail on tampered message")
	}
}

func TestIsEOA(t *testing.T) {
	// A real EOA address (Vitalik).
	eoa := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	if !IsEOA(eoa) {
		t.Fatalf("expected %s to be EOA", eoa.Hex())
	}
}
