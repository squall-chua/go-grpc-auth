package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Consent struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	ClientID  string        `bson:"client_id"`
	Namespace string        `bson:"namespace"`
	Scopes    []string      `bson:"scopes"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
