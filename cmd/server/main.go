package main

import (
	"context"
	"fmt"
	"net"
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
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/templates"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification/wiring"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
	"github.com/squall-chua/go-grpc-auth/internal/service/ratelimit"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/service/webhook"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufconnSize = 1024 * 1024

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
	auditSvc := audit.NewAuditService(auditRepo)

	notifRegistry, _, err := wiring.BuildRegistry(ctx, wiring.BuildConfig{
		DefaultEmailProvider: cfg.DefaultEmailProvider,
		DefaultSMSProvider:   cfg.DefaultSMSProvider,
		SMTPHost:             cfg.SMTPHost,
		SMTPPort:             cfg.SMTPPort,
		SMTPUsername:         cfg.SMTPUsername,
		SMTPPassword:         cfg.SMTPPassword,
		SMTPFromAddress:      cfg.SMTPFromAddress,
		SMTPFromName:         cfg.SMTPFromName,
		SMTPUseTLS:           cfg.SMTPUseTLS,
		SESRegion:            cfg.SESRegion,
		SESFromAddress:       cfg.SESFromAddress,
		SESFromName:          cfg.SESFromName,
		SESAccessKeyID:       cfg.SESAccessKeyID,
		SESSecretAccessKey:   cfg.SESSecretAccessKey,
		SNSRegion:            cfg.SNSRegion,
		SNSSenderID:          cfg.SNSSenderID,
		SNSAccessKeyID:       cfg.SNSAccessKeyID,
		SNSSecretAccessKey:   cfg.SNSSecretAccessKey,
	})
	if err != nil {
		zap.L().Fatal("failed to build notification registry", zap.Error(err))
	}
	notifTemplates := notification.NewTemplateRegistry()
	notifTemplates.RegisterEmail(templates.MFAEmailOTP)
	notifTemplates.RegisterSMS(templates.MFASMSOTP)

	notifier := notification.NewService(
		notifRegistry,
		notifTemplates,
		namespaceResolverAdapter{repo: nsRepo},
		auditEmitterAdapter{svc: auditSvc},
	)

	mfaSvc := authservice.NewMFAService(mfaRepo, notifier, cfg.AppName)

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
	if cfg.FacebookClientID != "" {
		socialProviders = append(socialProviders, authservice.NewFacebookProvider(cfg.FacebookClientID, cfg.FacebookClientSecret, cfg.FacebookRedirectURL))
	}
	if cfg.TwitterClientID != "" {
		socialProviders = append(socialProviders, authservice.NewTwitterProvider(cfg.TwitterClientID, cfg.TwitterClientSecret, cfg.TwitterRedirectURL))
	}
	if cfg.AppleClientID != "" {
		// Apple requires a .p8 key path; treat a missing or unreadable file
		// as a startup failure (loud) so operators do not silently run with
		// a half-configured Apple provider.
		appleProvider, err := authservice.NewAppleProvider(cfg.AppleTeamID, cfg.AppleKeyID, cfg.AppleClientID, cfg.ApplePrivateKeyPath, cfg.AppleRedirectURL)
		if err != nil {
			logger.Fatal("init apple provider", zap.Error(err))
		}
		socialProviders = append(socialProviders, appleProvider)
	}
	socialAuthSvc := authservice.NewSocialAuthService(userRepo, tokenSvc, socialProviders)

	// Servers
	grpcSrv := server.NewGRPCServer(server.GRPCServerConfig{
		AuthService:       authSvc,
		SocialAuthService: socialAuthSvc,
		AdminService:      adminSvc,
		NamespaceService:  nsSvc,
		NotifRegistry:     notifRegistry,
		OIDCService:       oidcSvc,
		OIDCClientService: oidcClientSvc,
	})

	// In-memory listener used by grpc-gateway to dial the gRPC server,
	// avoiding a TCP loopback round-trip on every REST request.
	bufLis := bufconn.Listen(bufconnSize)
	gwConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return bufLis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("Failed to create gateway gRPC client", zap.Error(err))
	}
	defer gwConn.Close()

	gatewaySrv, err := server.NewGatewayServer(ctx, gwConn, server.UIConfig{
		ApiBase: fmt.Sprintf("http://localhost:%s", cfg.Port),
		AppName: cfg.AppName,
	})
	if err != nil {
		logger.Fatal("Failed to create gateway server", zap.Error(err))
	}

	srv := server.NewServer(cfg.Port, bufLis, grpcSrv, gatewaySrv)

	logger.Info("Starting multiplexed server", zap.String("port", cfg.Port))
	if err := srv.Start(ctx); err != nil {
		logger.Fatal("Server stopped with error", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

// namespaceResolverAdapter adapts repository.NamespaceRepository to
// notification.NamespaceResolver.
type namespaceResolverAdapter struct {
	repo repository.NamespaceRepository
}

func (a namespaceResolverAdapter) NotificationConfig(ctx context.Context, namespace string) (notification.NamespaceNotificationView, error) {
	ns, err := a.repo.GetByName(ctx, namespace)
	if err != nil {
		return notification.NamespaceNotificationView{}, err
	}
	v := notification.NamespaceNotificationView{
		EmailProvider: ns.Config.Notification.EmailProvider,
		SMSProvider:   ns.Config.Notification.SMSProvider,
	}
	if len(ns.Config.Notification.EmailTemplates) > 0 {
		v.EmailTemplates = make(map[string]notification.EmailTemplateOverride, len(ns.Config.Notification.EmailTemplates))
		for k, o := range ns.Config.Notification.EmailTemplates {
			v.EmailTemplates[k] = notification.EmailTemplateOverride{
				Subject:  o.Subject,
				HTMLBody: o.HTMLBody,
				TextBody: o.TextBody,
			}
		}
	}
	if len(ns.Config.Notification.SMSTemplates) > 0 {
		v.SMSTemplates = make(map[string]notification.SMSTemplateOverride, len(ns.Config.Notification.SMSTemplates))
		for k, o := range ns.Config.Notification.SMSTemplates {
			v.SMSTemplates[k] = notification.SMSTemplateOverride{Body: o.Body}
		}
	}
	return v, nil
}

// auditEmitterAdapter adapts audit.AuditService to notification.AuditEmitter.
type auditEmitterAdapter struct {
	svc audit.AuditService
}

func (a auditEmitterAdapter) LogNotification(ctx context.Context, event, userID, namespace string, metadata any) {
	a.svc.Log(ctx, domain.AuditEvent(event), userID, namespace, "", "", metadata)
}
