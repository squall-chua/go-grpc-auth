package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Token struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	TokenHash string        `bson:"token_hash"`
	UserID    string        `bson:"user_id"`
	Namespace string        `bson:"namespace"`
	Type      TokenType     `bson:"type"`
	Audience  string        `bson:"audience"`
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
	MFAMethods   []string
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
