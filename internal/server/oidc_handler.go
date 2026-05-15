package server

import (
	"context"
	"encoding/json"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type oidcGRPCServer struct {
	auth.UnimplementedOIDCServiceServer
	service oidc.OIDCService
}

func (s *oidcGRPCServer) GetDiscovery(ctx context.Context, req *emptypb.Empty) (*structpb.Struct, error) {
	discovery, err := s.service.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to structpb.Struct for dynamic response
	data, err := json.Marshal(discovery)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return structpb.NewStruct(m)
}

func (s *oidcGRPCServer) GetJWKS(ctx context.Context, req *emptypb.Empty) (*structpb.Struct, error) {
	jwks, err := s.service.GetJWKS(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(jwks)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return structpb.NewStruct(m)
}

func (s *oidcGRPCServer) GetUserInfo(ctx context.Context, req *auth.GetUserInfoRequest) (*auth.UserInfo, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	u, err := s.service.GetUserInfo(ctx, p.UserId)
	if err != nil {
		return nil, err
	}

	return &auth.UserInfo{
		Sub:       u.ID.Hex(),
		Name:      u.Username,
		Email:     u.Email,
		Namespace: u.Namespace,
		Roles:     u.Roles,
	}, nil
}

func (s *oidcGRPCServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return s.service.Authorize(ctx, req)
}

func (s *oidcGRPCServer) Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	return s.service.Token(ctx, req)
}

func (s *oidcGRPCServer) GetConsentRequest(ctx context.Context, req *auth.GetConsentRequestRequest) (*auth.ConsentDetails, error) {
	return s.service.GetConsentRequest(ctx, req)
}

func (s *oidcGRPCServer) AcceptConsent(ctx context.Context, req *auth.AcceptConsentRequest) (*emptypb.Empty, error) {
	return s.service.AcceptConsent(ctx, req)
}

func (s *oidcGRPCServer) RejectConsent(ctx context.Context, req *auth.RejectConsentRequest) (*emptypb.Empty, error) {
	return s.service.RejectConsent(ctx, req)
}

func (s *oidcGRPCServer) Logout(ctx context.Context, req *auth.OIDCLogoutRequest) (*auth.LogoutResponse, error) {
	return s.service.Logout(ctx, req)
}
