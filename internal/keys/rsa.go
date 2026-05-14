package keys

import (
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}

func GenerateKID(key *rsa.PublicKey) string {
	if key == nil {
		return "default-key"
	}
	// Use SHA-256 hash of public key as KID
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", key.N)))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}
