package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthCode struct {
	ID                  bson.ObjectID `bson:"_id,omitempty"`
	Code                string        `bson:"code"`
	ClientID            string        `bson:"client_id"`
	UserID              string        `bson:"user_id"`
	Namespace           string        `bson:"namespace"`
	RedirectURI         string        `bson:"redirect_uri"`
	Scopes              []string      `bson:"scopes"`
	CodeChallenge       string        `bson:"code_challenge"`
	CodeChallengeMethod string        `bson:"code_challenge_method"`
	Nonce               string        `bson:"nonce"`
	ExpiresAt           time.Time     `bson:"expires_at"`
}
