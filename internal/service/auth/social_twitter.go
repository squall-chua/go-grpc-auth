package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"golang.org/x/oauth2"
)

// Twitter/X OAuth 2.0 endpoint. Twitter deprecated OAuth 1.0a for new apps in
// 2023 and now supports OAuth 2.0 for new apps. There is no first-party shim
// in golang.org/x/oauth2, so we define the endpoint inline. Twitter requires
// client credentials in the request body rather than the Basic auth header,
// so we set AuthStyle accordingly.
var twitterEndpoint = oauth2.Endpoint{
	AuthURL:   "https://twitter.com/i/oauth2/authorize",
	TokenURL:  "https://api.twitter.com/2/oauth2/token",
	AuthStyle: oauth2.AuthStyleInParams,
}

type twitterProvider struct {
	config *oauth2.Config
}

func NewTwitterProvider(clientID, clientSecret, redirectURL string) domain.SocialProviderInterface {
	return &twitterProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     twitterEndpoint,
			Scopes:       []string{"tweet.read", "users.read", "users.email"},
		},
	}
}

func (p *twitterProvider) GetProvider() domain.SocialProvider {
	return domain.ProviderTwitter
}

func (p *twitterProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *twitterProvider) ExchangeCode(ctx context.Context, code string) (*domain.SocialUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("twitter exchange failed: %w", err)
	}

	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://api.twitter.com/2/users/me?user.fields=id,name,username,profile_image_url,verified_email")
	if err != nil {
		return nil, fmt.Errorf("twitter userinfo failed: %w", err)
	}
	defer resp.Body.Close()

	var payload struct {
		Data struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Username      string `json:"username"`
			ProfileImage  string `json:"profile_image_url"`
			Email         string `json:"email,omitempty"`
			VerifiedEmail string `json:"verified_email,omitempty"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("twitter decode failed: %w", err)
	}
	if payload.Data.Email == "" {
		return nil, fmt.Errorf("twitter: email not returned (the app must request users.email scope and the user must approve it)")
	}

	return &domain.SocialUser{
		ID:            payload.Data.ID,
		Email:         payload.Data.Email,
		Name:          payload.Data.Name,
		AvatarURL:     payload.Data.ProfileImage,
		EmailVerified: payload.Data.VerifiedEmail != "",
	}, nil
}
