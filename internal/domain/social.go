package domain

import "context"

type SocialProvider string

const (
	ProviderGoogle SocialProvider = "google"
	ProviderGitHub SocialProvider = "github"
)

type SocialIdentity struct {
	Provider   SocialProvider `bson:"provider"`
	ExternalID string         `bson:"external_id"`
	Email      string         `bson:"email"`
	Name       string         `bson:"name"`
	AvatarURL  string         `bson:"avatar_url"`
}

type SocialUser struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

type SocialProviderInterface interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*SocialUser, error)
	GetProvider() SocialProvider
}
