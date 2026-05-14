package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleProvider struct {
	config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string) domain.SocialProviderInterface {
	return &googleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
	}
}

func (p *googleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *googleProvider) GetProvider() domain.SocialProvider {
	return domain.ProviderGoogle
}

func (p *googleProvider) ExchangeCode(ctx context.Context, code string) (*domain.SocialUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange failed: %w", err)
	}

	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %w", err)
	}
	defer resp.Body.Close()

	var profile struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	return &domain.SocialUser{
		ID:        profile.ID,
		Email:     profile.Email,
		Name:      profile.Name,
		AvatarURL: profile.Picture,
	}, nil
}
