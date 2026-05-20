package templates

import "github.com/squall-chua/go-grpc-auth/internal/service/notification"

// MFAEmailOTP is the template for emailed MFA verification codes.
// Data: {Code, TTLMinutes, AppName}.
var MFAEmailOTP = notification.MustEmailTemplate(
	"mfa_email_otp",
	`{{.AppName}} verification code`,
	`<p>Your verification code is <b>{{.Code}}</b>.</p><p>It expires in {{.TTLMinutes}} minutes. If you did not request this, you can safely ignore the message.</p>`,
	`Your verification code is {{.Code}}. It expires in {{.TTLMinutes}} minutes. If you did not request this, you can safely ignore the message.`,
)

// MFASMSOTP is the template for SMS MFA verification codes.
// Data: {Code, TTLMinutes}.
var MFASMSOTP = notification.MustSMSTemplate(
	"mfa_sms_otp",
	`Your code: {{.Code}} (expires in {{.TTLMinutes}}m)`,
)
