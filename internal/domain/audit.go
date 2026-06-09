package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditEvent string

const (
	EventLoginSuccess            AuditEvent = "LOGIN_SUCCESS"
	EventLoginFailed             AuditEvent = "LOGIN_FAILED"
	EventRegisterSuccess         AuditEvent = "REGISTER_SUCCESS"
	EventMFAChallenge            AuditEvent = "MFA_CHALLENGE"
	EventMFAVerified             AuditEvent = "MFA_VERIFIED"
	EventMFAFailed               AuditEvent = "MFA_FAILED"
	EventTokenIssued             AuditEvent = "TOKEN_ISSUED"
	EventTokenRefreshed          AuditEvent = "TOKEN_REFRESHED"
	EventTokenRevoked            AuditEvent = "TOKEN_REVOKED"
	EventRoleAssigned            AuditEvent = "ROLE_ASSIGNED"
	EventRoleRevoked             AuditEvent = "ROLE_REVOKED"
	EventUserBanned              AuditEvent = "USER_BANNED"
	EventUserUnbanned            AuditEvent = "USER_UNBANNED"
	EventNamespaceCreated        AuditEvent = "NAMESPACE_CREATED"
	EventNotificationEmailSent   AuditEvent = "NOTIFICATION_EMAIL_SENT"
	EventNotificationEmailFailed AuditEvent = "NOTIFICATION_EMAIL_FAILED"
	EventNotificationSMSSent     AuditEvent = "NOTIFICATION_SMS_SENT"
	EventNotificationSMSFailed   AuditEvent = "NOTIFICATION_SMS_FAILED"

	EventWeb3NonceIssued        AuditEvent = "WEB3_NONCE_ISSUED"
	EventWeb3SignInSuccess      AuditEvent = "WEB3_SIGNIN_SUCCESS"
	EventWeb3SignInFailed       AuditEvent = "WEB3_SIGNIN_FAILED"
	EventWeb3WalletLinked       AuditEvent = "WEB3_WALLET_LINKED"
	EventWeb3WalletUnlinked     AuditEvent = "WEB3_WALLET_UNLINKED"
	EventWeb3WalletLinkConflict AuditEvent = "WEB3_WALLET_LINK_CONFLICT"
)

type AuditLog struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Event     AuditEvent    `bson:"event" json:"event"`
	UserID    string        `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Namespace string        `bson:"namespace" json:"namespace"`
	IP        string        `bson:"ip" json:"ip"`
	UserAgent string        `bson:"user_agent" json:"user_agent"`
	Metadata  any           `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
}
