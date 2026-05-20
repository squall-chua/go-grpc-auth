package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                 string
	MongoURI             string
	RedisURI             string
	RSAPrivateKeyPath    string
	RSAPublicKeyPath     string
	Issuer               string
	AppName              string
	GoogleClientID       string
	GoogleClientSecret   string
	GoogleRedirectURL    string
	GitHubClientID       string
	GitHubClientSecret   string
	GitHubRedirectURL    string
	RateLimitRequests    int
	RateLimitWindow      time.Duration
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	// Notification defaults
	DefaultEmailProvider string
	DefaultSMSProvider   string

	// SMTP — enabled when SMTPHost != ""
	SMTPHost        string
	SMTPPort        int
	SMTPUsername    string
	SMTPPassword    string
	SMTPFromAddress string
	SMTPFromName    string
	SMTPUseTLS      bool

	// SES — enabled when SESRegion != ""
	SESRegion         string
	SESFromAddress    string
	SESFromName       string
	SESAccessKeyID    string
	SESSecretAccessKey string

	// SNS — enabled when SNSRegion != ""
	SNSRegion         string
	SNSSenderID       string
	SNSAccessKeyID    string
	SNSSecretAccessKey string
}

func Load() Config {
	var cfg Config
	flag.StringVar(&cfg.Port, "port", getEnv("PORT", "8080"), "Server port (multiplexed gRPC + HTTP)")
	flag.StringVar(&cfg.MongoURI, "mongo-uri", getEnv("MONGO_URI", "mongodb://localhost:27017/auth_db"), "MongoDB connection string")
	flag.StringVar(&cfg.RedisURI, "redis-uri", getEnv("REDIS_URI", "localhost:6379"), "Redis URI (optional)")
	flag.StringVar(&cfg.RSAPrivateKeyPath, "rsa-private-key", getEnv("RSA_PRIVATE_KEY", "keys/id_rs256"), "RSA private key path")
	flag.StringVar(&cfg.RSAPublicKeyPath, "rsa-public-key", getEnv("RSA_PUBLIC_KEY", "keys/id_rs256.pub"), "RSA public key path")
	flag.StringVar(&cfg.Issuer, "issuer", getEnv("ISSUER", "https://auth.example.com"), "Token issuer URL")
	flag.StringVar(&cfg.AppName, "app-name", getEnv("APP_NAME", "Go"), "Application name displayed in the UI")

	flag.StringVar(&cfg.GoogleClientID, "google-client-id", getEnv("GOOGLE_CLIENT_ID", ""), "Google Client ID")
	flag.StringVar(&cfg.GoogleClientSecret, "google-client-secret", getEnv("GOOGLE_CLIENT_SECRET", ""), "Google Client Secret")
	flag.StringVar(&cfg.GoogleRedirectURL, "google-redirect-url", getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/social/google/callback"), "Google Redirect URL")
	flag.StringVar(&cfg.GitHubClientID, "github-client-id", getEnv("GITHUB_CLIENT_ID", ""), "GitHub Client ID")
	flag.StringVar(&cfg.GitHubClientSecret, "github-client-secret", getEnv("GITHUB_CLIENT_SECRET", ""), "GitHub Client Secret")
	flag.StringVar(&cfg.GitHubRedirectURL, "github-redirect-url", getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/social/github/callback"), "GitHub Redirect URL")

	flag.IntVar(&cfg.RateLimitRequests, "rate-limit-requests", getEnvInt("RATE_LIMIT_REQUESTS", 5), "Max requests per window")
	flag.DurationVar(&cfg.RateLimitWindow, "rate-limit-window", getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute), "Rate limit window duration")
	flag.DurationVar(&cfg.AccessTokenDuration, "access-token-duration", getEnvDuration("ACCESS_TOKEN_DURATION", 15*time.Minute), "Access token duration")
	flag.DurationVar(&cfg.RefreshTokenDuration, "refresh-token-duration", getEnvDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour), "Refresh token duration")

	// Notification
	flag.StringVar(&cfg.DefaultEmailProvider, "default-email-provider", getEnv("DEFAULT_EMAIL_PROVIDER", "log"), "Default email provider name (smtp|ses|log)")
	flag.StringVar(&cfg.DefaultSMSProvider, "default-sms-provider", getEnv("DEFAULT_SMS_PROVIDER", "log"), "Default SMS provider name (sns|log)")
	flag.StringVar(&cfg.SMTPHost, "smtp-host", getEnv("SMTP_HOST", ""), "SMTP host (enables smtp provider)")
	flag.IntVar(&cfg.SMTPPort, "smtp-port", getEnvInt("SMTP_PORT", 587), "SMTP port")
	flag.StringVar(&cfg.SMTPUsername, "smtp-username", getEnv("SMTP_USERNAME", ""), "SMTP username")
	flag.StringVar(&cfg.SMTPPassword, "smtp-password", getEnv("SMTP_PASSWORD", ""), "SMTP password")
	flag.StringVar(&cfg.SMTPFromAddress, "smtp-from-address", getEnv("SMTP_FROM_ADDRESS", ""), "SMTP From address")
	flag.StringVar(&cfg.SMTPFromName, "smtp-from-name", getEnv("SMTP_FROM_NAME", ""), "SMTP From name")
	flag.BoolVar(&cfg.SMTPUseTLS, "smtp-tls", getEnvBool("SMTP_TLS", true), "Use TLS for SMTP connections")
	flag.StringVar(&cfg.SESRegion, "ses-region", getEnv("SES_REGION", ""), "AWS SES region (enables ses provider)")
	flag.StringVar(&cfg.SESFromAddress, "ses-from-address", getEnv("SES_FROM_ADDRESS", ""), "SES From address")
	flag.StringVar(&cfg.SESFromName, "ses-from-name", getEnv("SES_FROM_NAME", ""), "SES From name")
	flag.StringVar(&cfg.SESAccessKeyID, "ses-access-key-id", getEnv("SES_ACCESS_KEY_ID", ""), "SES AWS access key ID (falls back to default credential chain)")
	flag.StringVar(&cfg.SESSecretAccessKey, "ses-secret-access-key", getEnv("SES_SECRET_ACCESS_KEY", ""), "SES AWS secret access key")
	flag.StringVar(&cfg.SNSRegion, "sns-region", getEnv("SNS_REGION", ""), "AWS SNS region (enables sns provider)")
	flag.StringVar(&cfg.SNSSenderID, "sns-sender-id", getEnv("SNS_SENDER_ID", ""), "SNS Sender ID (optional)")
	flag.StringVar(&cfg.SNSAccessKeyID, "sns-access-key-id", getEnv("SNS_ACCESS_KEY_ID", ""), "SNS AWS access key ID (falls back to default credential chain)")
	flag.StringVar(&cfg.SNSSecretAccessKey, "sns-secret-access-key", getEnv("SNS_SECRET_ACCESS_KEY", ""), "SNS AWS secret access key")

	flag.Parse()
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
