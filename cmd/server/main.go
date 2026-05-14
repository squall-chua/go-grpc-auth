package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/squall-chua/go-grpc-auth/internal/service/webhook"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup MongoDB
	db, err := repository.NewMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
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

	// RSA Keys
	privateKey, err := keys.LoadRSAPrivateKey(cfg.RSAPrivateKeyPath)
	if err != nil {
		log.Printf("Warning: RS256 private key not loaded: %v", err)
	}
	publicKey, err := keys.LoadRSAPublicKey(cfg.RSAPublicKeyPath)
	if err != nil {
		log.Printf("Warning: RS256 public key not loaded: %v", err)
	}
	kid := keys.GenerateKID(publicKey)

	// Services
	tokenSvc := tokenservice.NewTokenService(tokenRepo, userRepo, privateKey, kid, cfg.Issuer)
	mfaSvc := authservice.NewMFAService(mfaRepo)
	auditSvc := audit.NewAuditService(auditRepo)

	var rateLimiter ratelimit.RateLimiter
	if cfg.RedisURI != "" {
		rdb := redis.NewClient(&redis.Options{
			Addr: cfg.RedisURI,
		})
		rateLimiter = ratelimit.NewRedisRateLimiter(rdb, 5, 1*time.Minute)
		log.Println("Rate limiting: Redis enabled")
	} else {
		rateLimiter = ratelimit.NewMemoryRateLimiter(5, 1*time.Minute)
		log.Println("Rate limiting: Memory fallback")
	}

	webhookSvc := webhook.NewWebhookService()
	authSvc := authservice.NewAuthService(userRepo, tokenSvc, nsRepo, mfaSvc, auditSvc, rateLimiter, webhookSvc)
	oidcClientSvc := adminservice.NewOIDCClientService(clientRepo)
	adminSvc := adminservice.NewAdminService(userRepo, roleRepo, permRepo, oidcClientSvc, auditSvc)
	nsSvc := adminservice.NewNamespaceService(nsRepo)
	oidcSvc := oidc.NewOIDCService(cfg.Issuer, publicKey, kid, userRepo)

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
		log.Fatalf("Failed to create gateway server: %v", err)
	}

	srv := server.NewServer(cfg.Port, grpcSrv, gatewaySrv)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
