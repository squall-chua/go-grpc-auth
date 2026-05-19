package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/squall-chua/go-grpc-auth/internal/config"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/keys"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/server"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
	"github.com/squall-chua/go-grpc-auth/internal/service/ratelimit"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/service/webhook"
	"go.uber.org/zap"
)

func main() {
	// Initialize Zap Logger
	logger, _ := zap.NewProduction()
	if os.Getenv("ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cfg := config.Load()

	// Setup context with signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Setup MongoDB
	db, err := repository.NewMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Close(ctx)

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	tokenRepo := repository.NewTokenRepository(db.DB)
	nsRepo := repository.NewNamespaceRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)
	permRepo := repository.NewPermissionRepository(db.DB)
	clientRepo := repository.NewClientRepository(db.DB)
	mfaRepo := repository.NewMFARepository(db.DB)
	auditRepo := repository.NewAuditRepository(db.DB)
	authCodeRepo := repository.NewAuthCodeRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	consentRepo := repository.NewConsentRepository(db.DB)

	// RSA Keys
	privateKey, err := keys.LoadRSAPrivateKey(cfg.RSAPrivateKeyPath)
	if err != nil {
		logger.Warn("RS256 private key not loaded", zap.Error(err))
	}
	publicKey, err := keys.LoadRSAPublicKey(cfg.RSAPublicKeyPath)
	if err != nil {
		logger.Warn("RS256 public key not loaded", zap.Error(err))
	}
	kid := keys.GenerateKID(publicKey)

	// Services
	tokenSvc := tokenservice.NewTokenService(tokenRepo, userRepo, clientRepo, roleRepo, privateKey, kid, cfg.Issuer, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	mfaSvc := authservice.NewMFAService(mfaRepo, authservice.MFAConfig{
		EmailEnabled: cfg.MFAEmailEnabled,
		SMSEnabled:   cfg.MFASMSEnabled,
	})
	auditSvc := audit.NewAuditService(auditRepo)

	var rateLimiter ratelimit.RateLimiter
	if cfg.RedisURI != "" {
		rdb := redis.NewClient(&redis.Options{
			Addr: cfg.RedisURI,
		})
		rateLimiter = ratelimit.NewRedisRateLimiter(rdb, cfg.RateLimitRequests, cfg.RateLimitWindow)
		logger.Info("Rate limiting: Redis enabled", zap.String("addr", cfg.RedisURI))
	} else {
		rateLimiter = ratelimit.NewMemoryRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
		logger.Info("Rate limiting: Memory fallback")
	}

	webhookSvc := webhook.NewWebhookService(nsRepo)
	authSvc := authservice.NewAuthService(cfg.Issuer, userRepo, tokenSvc, nsRepo, mfaSvc, auditSvc, rateLimiter, webhookSvc)
	oidcClientSvc := adminservice.NewOIDCClientService(clientRepo)
	adminSvc := adminservice.NewAdminService(userRepo, roleRepo, permRepo, oidcClientSvc, auditSvc)
	nsSvc := adminservice.NewNamespaceService(nsRepo)
	oidcSvc := oidc.NewOIDCService(cfg.Issuer, publicKey, kid, userRepo, clientRepo, authCodeRepo, tokenSvc, sessionRepo, consentRepo)

	// Social Providers
	var socialProviders []domain.SocialProviderInterface
	if cfg.GoogleClientID != "" {
		socialProviders = append(socialProviders, authservice.NewGoogleProvider(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL))
	}
	if cfg.GitHubClientID != "" {
		socialProviders = append(socialProviders, authservice.NewGitHubProvider(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubRedirectURL))
	}
	socialAuthSvc := authservice.NewSocialAuthService(userRepo, tokenSvc, socialProviders)

	// Servers
	grpcSrv := server.NewGRPCServer(server.GRPCServerConfig{
		AuthService:       authSvc,
		SocialAuthService: socialAuthSvc,
		AdminService:      adminSvc,
		NamespaceService:  nsSvc,
		OIDCService:       oidcSvc,
		OIDCClientService: oidcClientSvc,
	})

	gatewaySrv, err := server.NewGatewayServer(ctx, cfg.Port)
	if err != nil {
		logger.Fatal("Failed to create gateway server", zap.Error(err))
	}

	srv := server.NewServer(cfg.Port, grpcSrv, gatewaySrv)

	logger.Info("Starting multiplexed server", zap.String("port", cfg.Port))
	if err := srv.Start(ctx); err != nil {
		logger.Fatal("Server stopped with error", zap.Error(err))
	}
	
	logger.Info("Server exited gracefully")
}
