package keys

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

type JWK struct {
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	Kid string   `json:"kid"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func GenerateJWKS(publicKey *rsa.PublicKey, kid string) *JWKS {
	return &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Alg: "RS256",
				Use: "sig",
				Kid: kid,
				N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
			},
		},
	}
}
