package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
)

type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *domain.User, audience string, scopes []string) (*domain.TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*domain.Principal, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	RevokeTokens(ctx context.Context, accessToken, refreshToken string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	GenerateIDToken(ctx context.Context, user *domain.User, audience, nonce string) (string, error)
	GenerateClientToken(ctx context.Context, client *domain.OIDCClient, scopes []string) (*domain.TokenPair, error)
}

type tokenService struct {
	tokenRepo            repository.TokenRepository
	userRepo             repository.UserRepository
	clientRepo           repository.ClientRepository
	roleRepo             repository.RoleRepository
	privateKey           *rsa.PrivateKey
	kid                  string
	issuer               string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewTokenService(
	tokenRepo repository.TokenRepository,
	userRepo repository.UserRepository,
	clientRepo repository.ClientRepository,
	roleRepo repository.RoleRepository,
	privateKey *rsa.PrivateKey,
	kid, issuer string,
	accessTokenDuration, refreshTokenDuration time.Duration,
) TokenService {
	return &tokenService{
		tokenRepo:            tokenRepo,
		userRepo:             userRepo,
		clientRepo:           clientRepo,
		roleRepo:             roleRepo,
		privateKey:           privateKey,
		kid:                  kid,
		issuer:               issuer,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, user *domain.User, audience string, scopes []string) (*domain.TokenPair, error) {
	accessToken, accessHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	accessExp := time.Now().UTC().Add(s.accessTokenDuration)
	refreshExp := time.Now().UTC().Add(s.refreshTokenDuration)

	err = s.tokenRepo.Create(ctx, &domain.Token{
		TokenHash: accessHash,
		UserID:    user.ID.Hex(),
		Namespace: user.Namespace,
		Type:      domain.TokenTypeAccess,
		Audience:  audience,
		Scopes:    scopes,
		ExpiresAt: accessExp,
	})
	if err != nil {
		return nil, err
	}

	err = s.tokenRepo.Create(ctx, &domain.Token{
		TokenHash: refreshHash,
		UserID:    user.ID.Hex(),
		Namespace: user.Namespace,
		Type:      domain.TokenTypeRefresh,
		Audience:  audience,
		Scopes:    scopes,
		ExpiresAt: refreshExp,
	})
	if err != nil {
		return nil, err
	}

	aud := audience
	if aud == "" {
		aud = s.issuer
	}
	idToken, err := s.GenerateIDToken(ctx, user, aud, "")
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresIn:    int(s.accessTokenDuration.Seconds()),
	}, nil
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*domain.Principal, error) {
	hash := s.hashToken(token)
	t, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if t.Type != domain.TokenTypeAccess || t.ExpiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("token expired or invalid type")
	}

	// Try User
	user, err := s.userRepo.GetByID(ctx, t.UserID)
	if err == nil {
		permissions := s.resolvePermissions(ctx, user.Namespace, user.Roles, user.Permissions)
		return &domain.Principal{
			UserID:      user.ID.Hex(),
			Namespace:   user.Namespace,
			Roles:       user.Roles,
			Permissions: permissions,
			ExpiresAt:   t.ExpiresAt.Unix(),
		}, nil
	}

	// Try Client
	client, err := s.clientRepo.GetByID(ctx, t.UserID)
	if err == nil {
		return &domain.Principal{
			UserID:      client.ClientID,
			Namespace:   client.Namespace,
			Roles:       []string{"client"},
			Permissions: client.AllowedScopes,
			ExpiresAt:   t.ExpiresAt.Unix(),
		}, nil
	}

	return nil, fmt.Errorf("principal not found for token")
}

func (s *tokenService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	hash := s.hashToken(refreshToken)
	t, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if t.Type != domain.TokenTypeRefresh || t.ExpiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("refresh token expired or invalid type")
	}

	user, err := s.userRepo.GetByID(ctx, t.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for refresh: %w", err)
	}

	if err := s.tokenRepo.DeleteByHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	pair, err := s.GenerateTokenPair(ctx, user, t.Audience, t.Scopes)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *tokenService) RevokeTokens(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		if err := s.tokenRepo.DeleteByHash(ctx, s.hashToken(accessToken)); err != nil {
			return err
		}
	}
	if refreshToken != "" {
		if err := s.tokenRepo.DeleteByHash(ctx, s.hashToken(refreshToken)); err != nil {
			return err
		}
	}
	return nil
}

func (s *tokenService) RevokeAllForUser(ctx context.Context, userID string) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

func (s *tokenService) GenerateIDToken(ctx context.Context, user *domain.User, audience, nonce string) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("OIDC not configured: private key missing")
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":       s.issuer,
		"sub":       user.ID,
		"aud":       audience,
		"exp":       now.Add(s.accessTokenDuration).Unix(),
		"iat":       now.Unix(),
		"namespace": user.Namespace,
		"email":     user.Email,
		"username":  user.Username,
		"roles":     user.Roles,
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.kid

	return token.SignedString(s.privateKey)
}

func (s *tokenService) GenerateClientToken(ctx context.Context, client *domain.OIDCClient, scopes []string) (*domain.TokenPair, error) {
	accessToken, accessHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	accessExp := time.Now().UTC().Add(s.accessTokenDuration)

	err = s.tokenRepo.Create(ctx, &domain.Token{
		TokenHash: accessHash,
		UserID:    client.ClientID,
		Namespace: client.Namespace,
		Type:      domain.TokenTypeAccess,
		Scopes:    scopes,
		ExpiresAt: accessExp,
	})
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken: accessToken,
		ExpiresIn:   3600,
	}, nil
}

func (s *tokenService) generateOpaqueToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	return token, s.hashToken(token), nil
}

func (s *tokenService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *tokenService) resolvePermissions(ctx context.Context, namespace string, roles []string, userPermissions []string) []string {
	seen := make(map[string]bool)
	for _, p := range userPermissions {
		seen[p] = true
	}

	for _, roleName := range roles {
		role, err := s.roleRepo.GetByName(ctx, namespace, roleName)
		if err != nil {
			continue
		}
		for _, p := range role.Permissions {
			seen[p] = true
		}
	}

	result := make([]string, 0, len(seen))
	for p := range seen {
		result = append(result, p)
	}
	return result
}
