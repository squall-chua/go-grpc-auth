package oidc

import (
	"context"
	"crypto/rsa"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/keys"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
)

type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

type OIDCService interface {
	GetDiscovery(ctx context.Context) (*OIDCDiscovery, error)
	GetJWKS(ctx context.Context) (*keys.JWKS, error)
	GetUserInfo(ctx context.Context, userID string) (*domain.User, error)
}

type oidcService struct {
	issuer    string
	publicKey *rsa.PublicKey
	kid       string
	userRepo  repository.UserRepository
}

func NewOIDCService(issuer string, publicKey *rsa.PublicKey, kid string, userRepo repository.UserRepository) OIDCService {
	return &oidcService{
		issuer:    issuer,
		publicKey: publicKey,
		kid:       kid,
		userRepo:  userRepo,
	}
}

func (s *oidcService) GetDiscovery(ctx context.Context) (*OIDCDiscovery, error) {
	return &OIDCDiscovery{
		Issuer:                            s.issuer,
		AuthorizationEndpoint:             s.issuer + "/oauth2/authorize",
		TokenEndpoint:                     s.issuer + "/oauth2/token",
		UserinfoEndpoint:                  s.issuer + "/oauth2/userinfo",
		JwksURI:                           s.issuer + "/.well-known/jwks.json",
		RegistrationEndpoint:              s.issuer + "/oauth2/register",
		ScopesSupported:                   []string{"openid", "profile", "email", "roles"},
		ResponseTypesSupported:            []string{"code", "token", "id_token", "code id_token"},
		ResponseModesSupported:            []string{"query", "fragment", "form_post"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token", "client_credentials", "password"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post"},
		ClaimsSupported:                   []string{"sub", "iss", "auth_time", "name", "given_name", "family_name", "nickname", "profile", "picture", "website", "email", "email_verified", "locale", "zoneinfo", "roles"},
	}, nil
}

func (s *oidcService) GetJWKS(ctx context.Context) (*keys.JWKS, error) {
	return keys.GenerateJWKS(s.publicKey, s.kid), nil
}

func (s *oidcService) GetUserInfo(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
