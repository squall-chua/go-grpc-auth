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
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	mongoClient          *mongo.Client
	tokenRepo            repository.TokenRepository
	userRepo             repository.UserRepository
	clientRepo           repository.ClientRepository
	privateKey           *rsa.PrivateKey
	kid                  string
	issuer               string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewTokenService(
	mongoClient *mongo.Client,
	tokenRepo repository.TokenRepository,
	userRepo repository.UserRepository,
	clientRepo repository.ClientRepository,
	privateKey *rsa.PrivateKey,
	kid, issuer string,
	accessTokenDuration, refreshTokenDuration time.Duration,
) TokenService {
	return &tokenService{
		mongoClient:          mongoClient,
		tokenRepo:            tokenRepo,
		userRepo:             userRepo,
		clientRepo:           clientRepo,
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
		return &domain.Principal{
			UserID:      user.ID.Hex(),
			Namespace:   user.Namespace,
			Roles:       user.Roles,
			Permissions: user.Permissions,
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

	session, err := s.mongoClient.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	var pair *domain.TokenPair
	_, err = session.WithTransaction(ctx, func(sc context.Context) (interface{}, error) {
		if err := s.tokenRepo.DeleteByHash(sc, hash); err != nil {
			return nil, fmt.Errorf("failed to revoke refresh token: %w", err)
		}

		var txErr error
		pair, txErr = s.GenerateTokenPair(sc, user, t.Audience, t.Scopes)
		return nil, txErr
	})
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *tokenService) RevokeTokens(ctx context.Context, accessToken, refreshToken string) error {
	session, err := s.mongoClient.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc context.Context) (interface{}, error) {
		if accessToken != "" {
			if err := s.tokenRepo.DeleteByHash(sc, s.hashToken(accessToken)); err != nil {
				return nil, err
			}
		}
		if refreshToken != "" {
			if err := s.tokenRepo.DeleteByHash(sc, s.hashToken(refreshToken)); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
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
