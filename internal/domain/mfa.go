package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MFAMethod string

const (
	MFAMethodTOTP  MFAMethod = "totp"
	MFAMethodEmail MFAMethod = "email"
	MFAMethodSMS   MFAMethod = "sms"
)

type MFASecret struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	Method    MFAMethod     `bson:"method"`
	Secret    string        `bson:"secret"`
	Confirmed bool          `bson:"confirmed"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type MFAToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Token     string        `bson:"token"`
	UserID    string        `bson:"user_id"`
	Namespace string        `bson:"namespace"`
	Method    MFAMethod     `bson:"method"`
	ExpiresAt time.Time     `bson:"expires_at"`
}
