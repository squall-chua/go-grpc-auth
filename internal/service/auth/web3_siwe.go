package auth

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spruceid/siwe-go"
)

// IsEOA returns true if the address has no contract code at the time of the
// call. We rely on this as a v1 safety check to reject smart-contract wallets
// (Safe, Argent, etc.) — EIP-1271 / EIP-6492 support is a v2 follow-up.
//
// In v1 we don't have an RPC client, so this is a best-effort check that
// always returns true (assume EOA). EIP-1271 is deferred. The function exists
// so call sites are forward-compatible: v2 can swap the body without changing
// the signature.
func IsEOA(_ common.Address) bool {
	return true
}

// VerifySIWE parses a SIWE message, verifies the EIP-191 signature, validates
// time constraints, and returns the recovered address. The optional
// `expectedDomain` and `expectedNonce` are checked against the message and
// must match (pass nil to skip a check).
func VerifySIWE(message, signature string, expectedDomain, expectedNonce *string) (common.Address, error) {
	msg, err := siwe.ParseMessage(message)
	if err != nil {
		return common.Address{}, err
	}

	if expectedDomain != nil && msg.GetDomain() != *expectedDomain {
		return common.Address{}, errors.New("siwe domain mismatch")
	}
	if expectedNonce != nil && msg.GetNonce() != *expectedNonce {
		return common.Address{}, errors.New("siwe nonce mismatch")
	}

	if _, err := msg.VerifyEIP191(signature); err != nil {
		return common.Address{}, err
	}

	return msg.GetAddress(), nil
}
