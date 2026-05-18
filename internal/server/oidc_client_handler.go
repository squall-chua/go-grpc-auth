package server

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type oidcClientGRPCServer struct {
	admin.UnimplementedOIDCClientServiceServer
	service adminservice.OIDCClientService
}

func (s *oidcClientGRPCServer) RegisterClient(ctx context.Context, req *admin.RegisterClientRequest) (*admin.OIDCClient, error) {
	client := &domain.OIDCClient{
		Name:          req.Name,
		RedirectURIs:  req.RedirectUris,
		AllowedScopes: req.AllowedScopes,
		SkipConsent:   req.SkipConsent,
		Namespace:     req.Namespace,
	}
	secret, err := s.service.RegisterClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return &admin.OIDCClient{
		ClientId:      client.ClientID,
		ClientSecret:  secret,
		Name:          client.Name,
		RedirectUris:  client.RedirectURIs,
		AllowedScopes: client.AllowedScopes,
		SkipConsent:   client.SkipConsent,
		Namespace:     client.Namespace,
	}, nil
}

func (s *oidcClientGRPCServer) GetClient(ctx context.Context, req *admin.GetClientRequest) (*admin.OIDCClient, error) {
	client, err := s.service.GetClient(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}
	return &admin.OIDCClient{
		ClientId:      client.ClientID,
		Name:          client.Name,
		RedirectUris:  client.RedirectURIs,
		AllowedScopes: client.AllowedScopes,
		SkipConsent:   client.SkipConsent,
		Namespace:     client.Namespace,
	}, nil
}

func (s *oidcClientGRPCServer) UpdateClient(ctx context.Context, req *admin.UpdateClientRequest) (*admin.OIDCClient, error) {
	client, err := s.service.UpdateClient(ctx, req.ClientId, repository.ClientUpdateFields{
		Name:          req.Name,
		RedirectURIs:  req.RedirectUris,
		AllowedScopes: req.AllowedScopes,
		SkipConsent:   req.SkipConsent,
	})
	if err != nil {
		return nil, err
	}

	return &admin.OIDCClient{
		ClientId:      client.ClientID,
		Name:          client.Name,
		RedirectUris:  client.RedirectURIs,
		AllowedScopes: client.AllowedScopes,
		SkipConsent:   client.SkipConsent,
		Namespace:     client.Namespace,
	}, nil
}

func (s *oidcClientGRPCServer) DeleteClient(ctx context.Context, req *admin.DeleteClientRequest) (*emptypb.Empty, error) {
	if err := s.service.DeleteClient(ctx, req.ClientId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *oidcClientGRPCServer) ListClients(ctx context.Context, req *admin.ListClientsRequest) (*admin.ListClientsResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	clients, total, err := s.service.ListClients(ctx, req.Namespace, req.Query, page, pageSize)
	if err != nil {
		return nil, err
	}
	var protoClients []*admin.OIDCClient
	for _, c := range clients {
		protoClients = append(protoClients, &admin.OIDCClient{
			ClientId:      c.ClientID,
			Name:          c.Name,
			RedirectUris:  c.RedirectURIs,
			AllowedScopes: c.AllowedScopes,
			SkipConsent:   c.SkipConsent,
			Namespace:     c.Namespace,
		})
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return &admin.ListClientsResponse{
		Clients:    protoClients,
		TotalPages: totalPages,
		TotalCount: int32(total),
	}, nil
}

func (s *oidcClientGRPCServer) RotateClientSecret(ctx context.Context, req *admin.RotateClientSecretRequest) (*admin.RotateClientSecretResponse, error) {
	secret, err := s.service.RotateSecret(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}
	return &admin.RotateClientSecretResponse{
		ClientSecret: secret,
	}, nil
}
