package domain

import "time"

type MFAMethod string

const (
	MFAMethodTOTP  MFAMethod = "totp"
	MFAMethodEmail MFAMethod = "email"
	MFAMethodSMS   MFAMethod = "sms"
)

type MFASecret struct {
	UserID    string    `bson:"user_id"`
	Method    MFAMethod `bson:"method"`
	Secret    string    `bson:"secret"`
	Confirmed bool      `bson:"confirmed"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type MFAToken struct {
	Token     string    `bson:"token"`
	UserID    string    `bson:"user_id"`
	Namespace string    `bson:"namespace"`
	Method    MFAMethod `bson:"method"`
	ExpiresAt time.Time `bson:"expires_at"`
}
