package domain

import "time"

type OIDCClient struct {
	ClientID     string    `bson:"client_id"`
	ClientSecret string    `bson:"client_secret"` // Hashed
	Name         string    `bson:"name"`
	Namespace    string    `bson:"namespace"`
	RedirectURIs []string  `bson:"redirect_uris"`
	AllowedScopes []string `bson:"allowed_scopes"`
	SkipConsent  bool      `bson:"skip_consent"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}
