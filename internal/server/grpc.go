package server

import (
	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	authv1 "github.com/squall-chua/go-grpc-auth/api/v1/auth"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServerConfig struct {
	AuthService       authservice.AuthService
	SocialAuthService authservice.SocialAuthService
	AdminService      adminservice.AdminService
	NamespaceService  adminservice.NamespaceService
	OIDCService       oidc.OIDCService
	OIDCClientService adminservice.OIDCClientService
}

func NewGRPCServer(cfg GRPCServerConfig) *grpc.Server {
	s := grpc.NewServer()

	// Register services
	authv1.RegisterAuthServiceServer(s, &authGRPCServer{
		service:       cfg.AuthService,
		socialService: cfg.SocialAuthService,
	})
	admin.RegisterAdminServiceServer(s, &adminGRPCServer{
		service:       cfg.AdminService,
		clientService: cfg.OIDCClientService,
	})
	admin.RegisterNamespaceServiceServer(s, &namespaceGRPCServer{service: cfg.NamespaceService})
	admin.RegisterOIDCClientServiceServer(s, &oidcClientGRPCServer{service: cfg.OIDCClientService})
	authv1.RegisterOIDCServiceServer(s, &oidcGRPCServer{service: cfg.OIDCService})

	reflection.Register(s)
	return s
}
