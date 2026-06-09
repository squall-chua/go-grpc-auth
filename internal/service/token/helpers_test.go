package token

import (
	"crypto/rsa"
	"time"

	"github.com/squall-chua/go-grpc-auth/internal/repository"
)

// newTokenServiceForTest builds a tokenService with all repos nil; the methods
// exercised by the unit tests must not touch those repos.
//
// The repo parameters are accepted (and ignored) so the test helper mirrors the
// production NewTokenService signature; this keeps us honest if we later wire
// real repos for higher-fidelity tests.
func newTokenServiceForTest(priv *rsa.PrivateKey, kid, issuer string, _ repository.TokenRepository, _ repository.UserRepository, _ repository.ClientRepository, _ repository.RoleRepository) *tokenService {
	return &tokenService{
		tokenRepo:            nil,
		userRepo:             nil,
		clientRepo:           nil,
		roleRepo:             nil,
		privateKey:           priv,
		kid:                  kid,
		issuer:               issuer,
		accessTokenDuration:  15 * time.Minute,
		refreshTokenDuration: 24 * time.Hour,
	}
}
