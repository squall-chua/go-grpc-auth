package config

import (
	"flag"
	"os"
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

	flag.Parse()

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
