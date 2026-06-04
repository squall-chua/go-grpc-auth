package domain

import "context"

type SocialProvider string

const (
	ProviderGoogle   SocialProvider = "google"
	ProviderGitHub   SocialProvider = "github"
	ProviderFacebook SocialProvider = "facebook"
	ProviderTwitter  SocialProvider = "twitter"
	ProviderApple    SocialProvider = "apple"
)

type SocialIdentity struct {
	Provider      SocialProvider `bson:"provider"`
	ExternalID    string         `bson:"external_id"`
	Email         string         `bson:"email"`
	Name          string         `bson:"name"`
	AvatarURL     string         `bson:"avatar_url"`
	EmailVerified bool           `bson:"email_verified"`
}

type SocialUser struct {
	ID            string
	Email         string
	Name          string
	AvatarURL     string
	EmailVerified bool
}

type SocialProviderInterface interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*SocialUser, error)
	GetProvider() SocialProvider
}
