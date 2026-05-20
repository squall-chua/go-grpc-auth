package domain

import "go.mongodb.org/mongo-driver/v2/bson"

type Namespace struct {
	ID     bson.ObjectID   `bson:"_id,omitempty"`
	Name   string          `bson:"name"`
	Config NamespaceConfig `bson:"config"`
}

type NamespaceConfig struct {
	MFARequired            bool                        `bson:"mfa_required"`
	AllowedSocialProviders []string                    `bson:"allowed_social_providers"`
	PasswordPolicy         PasswordPolicy              `bson:"password_policy"`
	IPAllowList            []string                    `bson:"ip_allow_list"`
	IPDenyList             []string                    `bson:"ip_deny_list"`
	WebhookURL             string                      `bson:"webhook_url"`
	WebhookSecret          string                      `bson:"webhook_secret"`
	Notification           NamespaceNotificationConfig `bson:"notification,omitempty"`
}

type NamespaceNotificationConfig struct {
	EmailProvider  string                           `bson:"email_provider,omitempty"`
	SMSProvider    string                           `bson:"sms_provider,omitempty"`
	EmailTemplates map[string]EmailTemplateOverride `bson:"email_templates,omitempty"`
	SMSTemplates   map[string]SMSTemplateOverride   `bson:"sms_templates,omitempty"`
}

type EmailTemplateOverride struct {
	Subject  string `bson:"subject,omitempty"`
	HTMLBody string `bson:"html_body,omitempty"`
	TextBody string `bson:"text_body,omitempty"`
}

type SMSTemplateOverride struct {
	Body string `bson:"body,omitempty"`
}

type PasswordPolicy struct {
	MinLength        int  `bson:"min_length"`
	RequireUppercase bool `bson:"require_uppercase"`
	RequireLowercase bool `bson:"require_lowercase"`
	RequireNumber    bool `bson:"require_number"`
	RequireSpecial   bool `bson:"require_special"`
	PasswordHistory  int  `bson:"password_history"`
}
