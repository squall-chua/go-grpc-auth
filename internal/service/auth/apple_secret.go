package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleIssuer    = "https://appleid.apple.com"
	appleAudience  = "https://appleid.apple.com"
	appleJWKSURL   = "https://appleid.apple.com/auth/keys"
	appleSecretTTL = 5 * 24 * time.Hour // JWT lifetime; refresh triggered within 24h of expiry
)

type appleSecretSigner struct {
	teamID     string
	keyID      string
	clientID   string
	privateKey *ecdsa.PrivateKey

	mu        sync.Mutex
	cachedJWT string
	cachedExp time.Time
}

func newAppleSecretSigner(teamID, keyID, clientID, p8Path string) (*appleSecretSigner, error) {
	keyBytes, err := os.ReadFile(p8Path)
	if err != nil {
		return nil, fmt.Errorf("apple: read .p8 key: %w", err)
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("apple: .p8 file is not PEM-encoded")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apple: parse .p8 key: %w", err)
	}
	ecKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("apple: .p8 key is not ECDSA")
	}
	if ecKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("apple: .p8 key must use P-256 curve")
	}
	return &appleSecretSigner{
		teamID:     teamID,
		keyID:      keyID,
		clientID:   clientID,
		privateKey: ecKey,
	}, nil
}

func (s *appleSecretSigner) ClientSecret() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.cachedJWT != "" && now.Before(s.cachedExp.Add(-24*time.Hour)) {
		return s.cachedJWT, nil
	}

	exp := now.Add(appleSecretTTL)
	claims := jwt.MapClaims{
		"iss": s.teamID,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"aud": appleAudience,
		"sub": s.clientID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = s.keyID

	signed, err := tok.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("apple: sign client secret: %w", err)
	}
	s.cachedJWT = signed
	s.cachedExp = exp
	return signed, nil
}
