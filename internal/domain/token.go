package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Token struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	TokenHash string        `bson:"token_hash"`
	UserID    string        `bson:"user_id"`
	Namespace string        `bson:"namespace"`
	Type      string        `bson:"type"` // access, refresh
	Scopes    []string      `bson:"scopes"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	MFARequired  bool
	MFAToken     string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	MFARequired  bool   `json:"mfa_required,omitempty"`
	MFAToken     string `json:"mfa_token,omitempty"`
}

type Principal struct {
	UserID      string
	Namespace   string
	Roles       []string
	Permissions []string
	ExpiresAt   int64
}
