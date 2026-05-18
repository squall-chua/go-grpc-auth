package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	MongoURI          string
	RedisURI          string
	RSAPrivateKeyPath string
	RSAPublicKeyPath  string
	Issuer            string
	// Social Providers
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration
	// Token Durations
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	// MFA Delivery
	MFAEmailEnabled bool
	MFASMSEnabled   bool
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", getEnv("PORT", "8080"), "Server port (multiplexed gRPC + HTTP)")
	flag.StringVar(&cfg.MongoURI, "mongo-uri", getEnv("MONGO_URI", "mongodb://localhost:27017/auth_db"), "MongoDB connection string")
	flag.StringVar(&cfg.RedisURI, "redis-uri", getEnv("REDIS_URI", "localhost:6379"), "Redis URI (optional)")
	flag.StringVar(&cfg.RSAPrivateKeyPath, "rsa-private-key", getEnv("RSA_PRIVATE_KEY", "keys/id_rs256"), "RSA private key path")
	flag.StringVar(&cfg.RSAPublicKeyPath, "rsa-public-key", getEnv("RSA_PUBLIC_KEY", "keys/id_rs256.pub"), "RSA public key path")
	flag.StringVar(&cfg.Issuer, "issuer", getEnv("ISSUER", "https://auth.example.com"), "Token issuer URL")

	// Social
	flag.StringVar(&cfg.GoogleClientID, "google-client-id", getEnv("GOOGLE_CLIENT_ID", ""), "Google Client ID")
	flag.StringVar(&cfg.GoogleClientSecret, "google-client-secret", getEnv("GOOGLE_CLIENT_SECRET", ""), "Google Client Secret")
	flag.StringVar(&cfg.GoogleRedirectURL, "google-redirect-url", getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/social/google/callback"), "Google Redirect URL")
	flag.StringVar(&cfg.GitHubClientID, "github-client-id", getEnv("GITHUB_CLIENT_ID", ""), "GitHub Client ID")
	flag.StringVar(&cfg.GitHubClientSecret, "github-client-secret", getEnv("GITHUB_CLIENT_SECRET", ""), "GitHub Client Secret")
	flag.StringVar(&cfg.GitHubRedirectURL, "github-redirect-url", getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/social/github/callback"), "GitHub Redirect URL")

	// Rate Limiting
	flag.IntVar(&cfg.RateLimitRequests, "rate-limit-requests", getEnvInt("RATE_LIMIT_REQUESTS", 5), "Max requests per window")
	flag.DurationVar(&cfg.RateLimitWindow, "rate-limit-window", getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute), "Rate limit window duration")

	// Token Durations
	flag.DurationVar(&cfg.AccessTokenDuration, "access-token-duration", getEnvDuration("ACCESS_TOKEN_DURATION", 15*time.Minute), "Access token duration")
	flag.DurationVar(&cfg.RefreshTokenDuration, "refresh-token-duration", getEnvDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour), "Refresh token duration")

	// MFA Delivery
	flag.BoolVar(&cfg.MFAEmailEnabled, "mfa-email-enabled", getEnvBool("MFA_EMAIL_ENABLED", false), "Enable email OTP delivery")
	flag.BoolVar(&cfg.MFASMSEnabled, "mfa-sms-enabled", getEnvBool("MFA_SMS_ENABLED", false), "Enable SMS OTP delivery")

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
