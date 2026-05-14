package server

import (
	"context"
	"encoding/json"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/service/oidc"
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
	// TODO: Extract userID from context (set by interceptor)
	userID := "mock-user-id"

	u, err := s.service.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &auth.UserInfo{
		Sub:       u.ID,
		Name:      u.Username,
		Email:     u.Email,
		Namespace: u.Namespace,
		Roles:     u.Roles,
	}, nil
}
