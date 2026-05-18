package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

type User struct {
	ID               bson.ObjectID    `bson:"_id,omitempty"`
	Email            string           `bson:"email"`
	Username         string           `bson:"username"`
	PasswordHash     string           `bson:"password_hash"`
	Status           UserStatus       `bson:"status"`
	Roles            []string         `bson:"roles"`
	Permissions      []string         `bson:"permissions"`
	PasswordHistory  []string         `bson:"password_history"`
	Namespace        string           `bson:"namespace"`
	SocialIdentities []SocialIdentity `bson:"social_identities"`
	CreatedAt        time.Time        `bson:"created_at"`
	UpdatedAt        time.Time        `bson:"updated_at"`
}
