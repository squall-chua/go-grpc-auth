package domain

import "go.mongodb.org/mongo-driver/v2/bson"

type Namespace struct {
	ID     bson.ObjectID   `bson:"_id,omitempty"`
	Name   string          `bson:"name"`
	Config NamespaceConfig `bson:"config"`
}

type NamespaceConfig struct {
	MFARequired            bool           `bson:"mfa_required"`
	AllowedSocialProviders []string       `bson:"allowed_social_providers"`
	PasswordPolicy         PasswordPolicy `bson:"password_policy"`
	IPAllowlist            []string       `bson:"ip_allowlist"`
	IPDenylist             []string       `bson:"ip_denylist"`
	WebhookURL             string         `bson:"webhook_url"`
	WebhookSecret          string         `bson:"webhook_secret"`
}

type PasswordPolicy struct {
	MinLength        int  `bson:"min_length"`
	RequireUppercase bool `bson:"require_uppercase"`
	RequireLowercase bool `bson:"require_lowercase"`
	RequireNumber    bool `bson:"require_number"`
	RequireSpecial   bool `bson:"require_special"`
	PasswordHistory  int  `bson:"password_history"`
}
