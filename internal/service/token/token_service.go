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
	GenerateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*domain.Principal, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	RevokeTokens(ctx context.Context, accessToken, refreshToken string) error
	GenerateIDToken(ctx context.Context, user *domain.User, nonce string) (string, error)
}

type tokenService struct {
	tokenRepo  repository.TokenRepository
	userRepo   repository.UserRepository
	privateKey *rsa.PrivateKey
	kid        string
	issuer     string
}

func NewTokenService(tokenRepo repository.TokenRepository, userRepo repository.UserRepository, privateKey *rsa.PrivateKey, kid, issuer string) TokenService {
	return &tokenService{
		tokenRepo:  tokenRepo,
		userRepo:   userRepo,
		privateKey: privateKey,
		kid:        kid,
		issuer:     issuer,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessToken, accessHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	accessExp := time.Now().Add(15 * time.Minute)
	refreshExp := time.Now().Add(24 * 7 * time.Hour)

	err = s.tokenRepo.Create(ctx, &domain.Token{
		TokenHash: accessHash,
		UserID:    user.ID,
		Namespace: user.Namespace,
		Type:      "access",
		ExpiresAt: accessExp,
	})
	if err != nil {
		return nil, err
	}

	err = s.tokenRepo.Create(ctx, &domain.Token{
		TokenHash: refreshHash,
		UserID:    user.ID,
		Namespace: user.Namespace,
		Type:      "refresh",
		ExpiresAt: refreshExp,
	})
	if err != nil {
		return nil, err
	}

	idToken, err := s.GenerateIDToken(ctx, user, "")
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresIn:    900,
	}, nil
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*domain.Principal, error) {
	hash := s.hashToken(token)
	t, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if t.Type != "access" || t.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired or invalid type")
	}

	user, err := s.userRepo.GetByID(ctx, t.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for token: %w", err)
	}

	return &domain.Principal{
		UserID:      user.ID,
		Namespace:   user.Namespace,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		ExpiresAt:   t.ExpiresAt.Unix(),
	}, nil
}

func (s *tokenService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	hash := s.hashToken(refreshToken)
	t, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if t.Type != "refresh" || t.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired or invalid type")
	}

	user, err := s.userRepo.GetByID(ctx, t.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for refresh: %w", err)
	}

	// Revoke old refresh token
	_ = s.tokenRepo.DeleteByHash(ctx, hash)

	return s.GenerateTokenPair(ctx, user)
}

func (s *tokenService) RevokeTokens(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		_ = s.tokenRepo.DeleteByHash(ctx, s.hashToken(accessToken))
	}
	if refreshToken != "" {
		_ = s.tokenRepo.DeleteByHash(ctx, s.hashToken(refreshToken))
	}
	return nil
}

func (s *tokenService) GenerateIDToken(ctx context.Context, user *domain.User, nonce string) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("OIDC not configured: private key missing")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       s.issuer,
		"sub":       user.ID,
		"aud":       "auth-service", // Default audience, should be client_id in real OIDC
		"exp":       now.Add(1 * time.Hour).Unix(),
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
