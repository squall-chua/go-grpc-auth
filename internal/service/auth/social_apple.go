package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
)

const (
	appleAuthURL  = "https://appleid.apple.com/auth/authorize"
	appleTokenURL = "https://appleid.apple.com/auth/token"
)

type appleProvider struct {
	signer      *appleSecretSigner
	clientID    string
	redirectURL string
	http        *http.Client
}

func NewAppleProvider(teamID, keyID, clientID, p8Path, redirectURL string) (domain.SocialProviderInterface, error) {
	signer, err := newAppleSecretSigner(teamID, keyID, clientID, p8Path)
	if err != nil {
		return nil, err
	}
	return &appleProvider{
		signer:      signer,
		clientID:    clientID,
		redirectURL: redirectURL,
		http:        http.DefaultClient,
	}, nil
}

func (p *appleProvider) GetProvider() domain.SocialProvider {
	return domain.ProviderApple
}

func (p *appleProvider) GetAuthURL(state string) string {
	v := url.Values{}
	v.Set("client_id", p.clientID)
	v.Set("redirect_uri", p.redirectURL)
	v.Set("response_type", "code")
	v.Set("scope", "name email")
	v.Set("response_mode", "form_post")
	v.Set("state", state)
	return appleAuthURL + "?" + v.Encode()
}

func (p *appleProvider) ExchangeCode(ctx context.Context, code string) (*domain.SocialUser, error) {
	clientSecret, err := p.signer.ClientSecret()
	if err != nil {
		return nil, fmt.Errorf("apple: build client secret: %w", err)
	}

	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", p.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		appleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("apple: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apple: token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apple: token endpoint returned %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("apple: decode token response: %w", err)
	}
	if tokenResp.IDToken == "" {
		return nil, errors.New("apple: id_token missing from token response")
	}

	claims, err := decodeAppleIDToken(tokenResp.IDToken, p.clientID)
	if err != nil {
		return nil, err
	}

	return &domain.SocialUser{
		ID:        claims.Sub,
		Email:     claims.Email,
		Name:      claims.Name,
		AvatarURL: "",
		// Apple marks a relay email as `is_private_email=true`. We treat
		// relay emails as unverified for the purposes of account linkage
		// (see social_service.go and the EmailVerified flag).
		EmailVerified: claims.EmailVerified && !claims.IsPrivateEmail,
	}, nil
}

type appleIDClaims struct {
	jwt.RegisteredClaims
	Sub            string `json:"sub"`
	Email          string `json:"email"`
	EmailVerified  bool   `json:"email_verified"`
	IsPrivateEmail bool   `json:"is_private_email"`
	Name           string `json:"name,omitempty"`
}

// decodeAppleIDToken parses the Apple id_token, validates iss and aud, but
// does not perform RS256 signature verification in v1 (see ADR
// 0001-apple-client-secret-jwt.md). The token is received directly from
// Apple's HTTPS token endpoint, so the channel is trusted; signature
// verification is a follow-up tracked separately.
func decodeAppleIDToken(raw, expectedClientID string) (*appleIDClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claims := &appleIDClaims{}
	_, _, err := parser.ParseUnverified(raw, claims)
	if err != nil {
		return nil, fmt.Errorf("apple: parse id_token: %w", err)
	}

	// The iss and aud claims are inside the JWT but not in our struct. Use a
	// second pass with MapClaims to extract them without redefining the
	// struct.
	parserMap := jwt.NewParser(jwt.WithoutClaimsValidation())
	mapClaims := jwt.MapClaims{}
	if _, _, err := parserMap.ParseUnverified(raw, mapClaims); err != nil {
		return nil, fmt.Errorf("apple: parse id_token map claims: %w", err)
	}
	iss, _ := mapClaims["iss"].(string)
	audRaw, _ := mapClaims["aud"].(string)
	if iss != appleIssuer {
		return nil, fmt.Errorf("apple: id_token iss mismatch: %q", iss)
	}
	if audRaw != expectedClientID {
		return nil, fmt.Errorf("apple: id_token aud mismatch: %q", audRaw)
	}
	return claims, nil
}
