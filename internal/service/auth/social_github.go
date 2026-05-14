package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubProvider struct {
	config *oauth2.Config
}

func NewGitHubProvider(clientID, clientSecret, redirectURL string) domain.SocialProviderInterface {
	return &githubProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
			Scopes:       []string{"user:email", "read:user"},
		},
	}
}

func (p *githubProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *githubProvider) GetProvider() domain.SocialProvider {
	return domain.ProviderGitHub
}

func (p *githubProvider) ExchangeCode(ctx context.Context, code string) (*domain.SocialUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("github exchange failed: %w", err)
	}

	client := p.config.Client(ctx, token)
	
	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get github user: %w", err)
	}
	defer resp.Body.Close()

	var profile struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	// GitHub might return empty email if not public, need to fetch emails separately
	if profile.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary {
						profile.Email = e.Email
						break
					}
				}
			}
		}
	}

	return &domain.SocialUser{
		ID:        fmt.Sprintf("%d", profile.ID),
		Email:     profile.Email,
		Name:      profile.Name,
		AvatarURL: profile.AvatarURL,
	}, nil
}
