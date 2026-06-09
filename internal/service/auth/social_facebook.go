package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type facebookProvider struct {
	config *oauth2.Config
}

func NewFacebookProvider(clientID, clientSecret, redirectURL string) domain.SocialProviderInterface {
	return &facebookProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     facebook.Endpoint,
			Scopes:       []string{"email", "public_profile"},
		},
	}
}

func (p *facebookProvider) GetProvider() domain.SocialProvider {
	return domain.ProviderFacebook
}

func (p *facebookProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *facebookProvider) ExchangeCode(ctx context.Context, code string) (*domain.SocialUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("facebook exchange failed: %w", err)
	}

	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)")
	if err != nil {
		return nil, fmt.Errorf("facebook userinfo failed: %w", err)
	}
	defer resp.Body.Close()

	var profile struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("facebook decode failed: %w", err)
	}
	if profile.Email == "" {
		return nil, fmt.Errorf("facebook: user denied email permission (required)")
	}

	return &domain.SocialUser{
		ID:            profile.ID,
		Email:         profile.Email,
		Name:          profile.Name,
		AvatarURL:     profile.Picture.Data.URL,
		EmailVerified: true,
	}, nil
}
