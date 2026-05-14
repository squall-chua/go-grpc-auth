package server

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
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
	client, err := s.service.GetClient(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}

	client.Name = req.Name
	client.RedirectURIs = req.RedirectUris
	client.AllowedScopes = req.AllowedScopes
	client.SkipConsent = req.SkipConsent

	if err := s.service.UpdateClient(ctx, client); err != nil {
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
	clients, err := s.service.ListClients(ctx, req.Namespace)
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
	return &admin.ListClientsResponse{Clients: protoClients}, nil
}
