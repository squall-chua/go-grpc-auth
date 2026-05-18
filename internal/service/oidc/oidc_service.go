package oidc

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/keys"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
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
	Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)
	Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error)
	GetConsentRequest(ctx context.Context, req *auth.GetConsentRequestRequest) (*auth.ConsentDetails, error)
	AcceptConsent(ctx context.Context, req *auth.AcceptConsentRequest) (*emptypb.Empty, error)
	RejectConsent(ctx context.Context, req *auth.RejectConsentRequest) (*emptypb.Empty, error)
	Logout(ctx context.Context, req *auth.OIDCLogoutRequest) (*auth.LogoutResponse, error)
}

type oidcService struct {
	issuer       string
	publicKey    *rsa.PublicKey
	kid          string
	userRepo     repository.UserRepository
	clientRepo   repository.ClientRepository
	authCodeRepo repository.AuthCodeRepository
	tokenService token.TokenService
	sessionRepo  repository.SessionRepository
	consentRepo  repository.ConsentRepository
}

func NewOIDCService(
	issuer string,
	publicKey *rsa.PublicKey,
	kid string,
	userRepo repository.UserRepository,
	clientRepo repository.ClientRepository,
	authCodeRepo repository.AuthCodeRepository,
	tokenService token.TokenService,
	sessionRepo repository.SessionRepository,
	consentRepo repository.ConsentRepository,
) OIDCService {
	return &oidcService{
		issuer:       issuer,
		publicKey:    publicKey,
		kid:          kid,
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		tokenService: tokenService,
		sessionRepo:  sessionRepo,
		consentRepo:  consentRepo,
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

func (s *oidcService) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	// 1. Validate Client
	client, err := s.clientRepo.GetByID(ctx, req.ClientId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "client not found")
	}

	// 2. Validate Redirect URI
	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == req.RedirectUri {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return nil, status.Error(codes.InvalidArgument, "invalid redirect_uri")
	}

	// 3. Check Session (SSO)
	principal := util.GetPrincipal(ctx)
	if principal == nil {
		// No session -> Redirect to login page
		loginURL := fmt.Sprintf("/ui/login?return_to=%s", url.QueryEscape(s.issuer+"/oauth2/authorize?"+url.Values{
			"response_type": {req.ResponseType},
			"client_id":     {req.ClientId},
			"redirect_uri":  {req.RedirectUri},
			"scope":         {req.Scope},
			"state":         {req.State},
			"nonce":         {req.Nonce},
		}.Encode()))
		return &auth.AuthorizeResponse{RedirectUri: loginURL}, nil
	}

	// 4. Check Consent
	if !client.SkipConsent {
		consent, err := s.consentRepo.Get(ctx, principal.UserId, req.ClientId)
		if err != nil || consent == nil {
			// Redirect to consent page
			consentURL := fmt.Sprintf("/ui/consent?client_id=%s&scope=%s&return_to=%s",
				req.ClientId,
				url.QueryEscape(req.Scope),
				url.QueryEscape(s.issuer+"/oauth2/authorize?"+url.Values{
					"response_type": {req.ResponseType},
					"client_id":     {req.ClientId},
					"redirect_uri":  {req.RedirectUri},
					"scope":         {req.Scope},
					"state":         {req.State},
					"nonce":         {req.Nonce},
				}.Encode()))
			return &auth.AuthorizeResponse{RedirectUri: consentURL}, nil
		}
	}

	// 5. Generate Auth Code
	code := util.RandomString(16)
	authCode := &domain.AuthCode{
		Code:                code,
		ClientID:            req.ClientId,
		UserID:              principal.UserId,
		Namespace:           principal.Namespace,
		RedirectURI:         req.RedirectUri,
		Scopes:              strings.Split(req.Scope, " "),
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Nonce:               req.Nonce,
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}

	if err := s.authCodeRepo.Save(ctx, authCode); err != nil {
		return nil, status.Error(codes.Internal, "failed to save auth code")
	}

	// 6. Redirect back with code
	u, err := url.Parse(req.RedirectUri)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid redirect_uri")
	}

	q := u.Query()
	q.Set("code", code)
	if req.State != "" {
		q.Set("state", req.State)
	}
	u.RawQuery = q.Encode()

	return &auth.AuthorizeResponse{RedirectUri: u.String()}, nil
}

func (s *oidcService) Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	switch req.GrantType {
	case "authorization_code":
		return s.handleAuthorizationCode(ctx, req)
	case "client_credentials":
		return s.handleClientCredentials(ctx, req)
	case "refresh_token":
		return s.handleRefreshToken(ctx, req)
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported grant_type")
	}
}

func (s *oidcService) handleAuthorizationCode(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	// 1. Get Auth Code
	ac, err := s.authCodeRepo.Get(ctx, req.Code)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired code")
	}

	// 2. Validate Client Credentials (if provided)
	client, err := s.clientRepo.GetByID(ctx, ac.ClientID)
	if err != nil {
		return nil, status.Error(codes.Internal, "client not found")
	}

	if client.ClientSecret != "" {
		if req.ClientSecret == "" {
			return nil, status.Error(codes.Unauthenticated, "client_secret is required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecret), []byte(req.ClientSecret)); err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid client secret")
		}
	}

	// 3. Revoke code (single use)
	_ = s.authCodeRepo.Delete(ctx, req.Code)

	// 4. Get User
	user, err := s.userRepo.GetByID(ctx, ac.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	// 5. Generate Tokens with scopes from authorization code
	pair, err := s.tokenService.GenerateTokenPair(ctx, user, ac.ClientID, ac.Scopes)
	if err != nil {
		return nil, err
	}

	return &auth.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(pair.ExpiresIn),
	}, nil
}

func (s *oidcService) handleClientCredentials(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	client, err := s.clientRepo.GetByID(ctx, req.ClientId)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid client credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecret), []byte(req.ClientSecret)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid client credentials")
	}

	pair, err := s.tokenService.GenerateClientToken(ctx, client, client.AllowedScopes)
	if err != nil {
		return nil, err
	}

	return &auth.TokenResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int32(pair.ExpiresIn),
	}, nil
}

func (s *oidcService) handleRefreshToken(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	pair, err := s.tokenService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &auth.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(pair.ExpiresIn),
	}, nil
}

func (s *oidcService) GetConsentRequest(ctx context.Context, req *auth.GetConsentRequestRequest) (*auth.ConsentDetails, error) {
	client, err := s.clientRepo.GetByID(ctx, req.ClientId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "client not found")
	}

	return &auth.ConsentDetails{
		ClientName: client.Name,
		Scopes:     req.Scopes,
	}, nil
}

func (s *oidcService) AcceptConsent(ctx context.Context, req *auth.AcceptConsentRequest) (*emptypb.Empty, error) {
	principal := util.GetPrincipal(ctx)
	if principal == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	consent := &domain.Consent{
		UserID:    principal.UserId,
		ClientID:  req.ClientId,
		Namespace: principal.Namespace,
		Scopes:    req.Scopes,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.consentRepo.Save(ctx, consent); err != nil {
		return nil, status.Error(codes.Internal, "failed to save consent")
	}

	// After consent, we should probably redirect back to Authorize
	// The frontend will handle the redirect to the return_to URL
	return &emptypb.Empty{}, nil
}

func (s *oidcService) RejectConsent(ctx context.Context, req *auth.RejectConsentRequest) (*emptypb.Empty, error) {
	// If user rejects, we redirect back to redirect_uri with error=access_denied
	// This requires knowing the original redirect_uri, which should be in return_to
	return &emptypb.Empty{}, nil
}

func (s *oidcService) Logout(ctx context.Context, req *auth.OIDCLogoutRequest) (*auth.LogoutResponse, error) {
	principal := util.GetPrincipal(ctx)
	if principal != nil {
		// Revoke all tokens for this user
		_ = s.tokenService.RevokeAllForUser(ctx, principal.UserId)
		// Delete SSO sessions
		_ = s.sessionRepo.DeleteByUserID(ctx, principal.UserId)
	}

	return &auth.LogoutResponse{RedirectUri: req.PostLogoutRedirectUri}, nil
}
