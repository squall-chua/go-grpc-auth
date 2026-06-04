package server

import (
	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	authv1 "github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServerConfig struct {
	AuthService       authservice.AuthService
	SocialAuthService authservice.SocialAuthService
	Web3AuthService   authservice.Web3AuthService
	Web3Issuer        string
	UserRepo          repository.UserRepository
	AdminService      adminservice.AdminService
	NamespaceService  adminservice.NamespaceService
	NotifRegistry     *notification.Registry
	OIDCService       oidc.OIDCService
	OIDCClientService adminservice.OIDCClientService
}

func NewGRPCServer(cfg GRPCServerConfig) *grpc.Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(AuthUnaryInterceptor(cfg.AuthService)),
	)

	// Register services
	authv1.RegisterAuthServiceServer(s, &authGRPCServer{
		service:       cfg.AuthService,
		socialService: cfg.SocialAuthService,
	})
	authv1.RegisterWeb3AuthServiceServer(s, &web3GRPCServer{
		service: cfg.Web3AuthService,
		issuer:  cfg.Web3Issuer,
		users:   cfg.UserRepo,
	})
	admin.RegisterAdminServiceServer(s, &adminGRPCServer{
		service:       cfg.AdminService,
		clientService: cfg.OIDCClientService,
		notifRegistry: cfg.NotifRegistry,
	})
	admin.RegisterNamespaceServiceServer(s, &namespaceGRPCServer{service: cfg.NamespaceService})
	admin.RegisterOIDCClientServiceServer(s, &oidcClientGRPCServer{service: cfg.OIDCClientService})
	authv1.RegisterOIDCServiceServer(s, &oidcGRPCServer{service: cfg.OIDCService})

	reflection.Register(s)
	return s
}
