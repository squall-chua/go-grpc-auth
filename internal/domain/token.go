package domain

import "time"

type Token struct {
	ID        string
	TokenHash string
	UserID    string
	Namespace string
	Type      string // access, refresh
	Scopes    []string
	ExpiresAt time.Time
	CreatedAt time.Time
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
