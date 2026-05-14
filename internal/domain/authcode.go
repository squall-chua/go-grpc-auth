package domain

import "time"

type AuthCode struct {
	Code                string
	ClientID            string
	UserID              string
	Namespace           string
	RedirectURI         string
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
	ExpiresAt           time.Time
}
