package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OIDCClient struct {
	ID                     bson.ObjectID `bson:"_id,omitempty"`
	ClientID               string        `bson:"client_id"`
	ClientSecret           string        `bson:"client_secret"` // Hashed
	Name                   string        `bson:"name"`
	Namespace              string        `bson:"namespace"`
	RedirectURIs           []string      `bson:"redirect_uris"`
	PostLogoutRedirectURIs []string      `bson:"post_logout_redirect_uris"`
	AllowedScopes          []string      `bson:"allowed_scopes"`
	SkipConsent            bool          `bson:"skip_consent"`
	CreatedAt              time.Time     `bson:"created_at"`
	UpdatedAt              time.Time     `bson:"updated_at"`
}
