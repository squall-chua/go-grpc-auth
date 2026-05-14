package server

import (
	"context"

	"encoding/json"
	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adminGRPCServer struct {
	admin.UnimplementedAdminServiceServer
	service       adminservice.AdminService
	clientService adminservice.OIDCClientService
}

func (s *adminGRPCServer) ListServiceAccounts(ctx context.Context, req *admin.ListServiceAccountsRequest) (*admin.ListServiceAccountsResponse, error) {
	clients, err := s.clientService.ListClients(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	var sas []*admin.ServiceAccount
	for _, c := range clients {
		sas = append(sas, &admin.ServiceAccount{
			ClientId: c.ClientID,
			Name:     c.Name,
			Namespace: c.Namespace,
			Scopes:    c.AllowedScopes,
		})
	}
	return &admin.ListServiceAccountsResponse{ServiceAccounts: sas}, nil
}

func (s *adminGRPCServer) CreateServiceAccount(ctx context.Context, req *admin.CreateServiceAccountRequest) (*admin.ServiceAccount, error) {
	client := &domain.OIDCClient{
		Name:          req.Name,
		Namespace:     req.Namespace,
		AllowedScopes: req.Scopes,
	}
	secret, err := s.clientService.RegisterClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return &admin.ServiceAccount{
		ClientId:     client.ClientID,
		ClientSecret: secret,
		Name:         client.Name,
		Namespace:    client.Namespace,
		Scopes:       client.AllowedScopes,
	}, nil
}

func (s *adminGRPCServer) ListUsers(ctx context.Context, req *admin.ListUsersRequest) (*admin.ListUsersResponse, error) {
	page := 1 // Basic paging for now
	users, total, err := s.service.ListUsers(ctx, req.Namespace, page, int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var protoUsers []*admin.User
	for _, u := range users {
		protoUsers = append(protoUsers, &admin.User{
			Id:        u.ID,
			Email:     u.Email,
			Username:  u.Username,
			Namespace: u.Namespace,
			Status:    u.Status,
			Roles:     u.Roles,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		})
	}

	return &admin.ListUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(total),
	}, nil
}

func (s *adminGRPCServer) GetUser(ctx context.Context, req *admin.GetUserRequest) (*admin.User, error) {
	u, err := s.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &admin.User{
		Id:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Namespace: u.Namespace,
		Status:    u.Status,
		Roles:     u.Roles,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}, nil
}

func (s *adminGRPCServer) UpdateUserStatus(ctx context.Context, req *admin.UpdateUserStatusRequest) (*emptypb.Empty, error) {
	err := s.service.UpdateUserStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) ResetUserPassword(ctx context.Context, req *admin.ResetUserPasswordRequest) (*emptypb.Empty, error) {
	err := s.service.ResetUserPassword(ctx, req.Id, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) GrantRoles(ctx context.Context, req *admin.GrantRolesRequest) (*emptypb.Empty, error) {
	err := s.service.GrantRoles(ctx, req.Id, req.Roles)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) RevokeRoles(ctx context.Context, req *admin.RevokeRolesRequest) (*emptypb.Empty, error) {
	err := s.service.RevokeRoles(ctx, req.Id, req.Roles)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) GrantPermissions(ctx context.Context, req *admin.GrantPermissionsRequest) (*emptypb.Empty, error) {
	err := s.service.GrantPermissions(ctx, req.Id, req.Permissions)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) RevokePermissions(ctx context.Context, req *admin.RevokePermissionsRequest) (*emptypb.Empty, error) {
	err := s.service.RevokePermissions(ctx, req.Id, req.Permissions)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) CreateRole(ctx context.Context, req *admin.CreateRoleRequest) (*admin.Role, error) {
	role := &domain.Role{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Permissions: req.Permissions,
	}
	err := s.service.CreateRole(ctx, role)
	if err != nil {
		return nil, err
	}
	return &admin.Role{
		Id:          role.ID,
		Name:        role.Name,
		Namespace:   role.Namespace,
		Permissions: role.Permissions,
	}, nil
}

func (s *adminGRPCServer) ListRoles(ctx context.Context, req *admin.ListRolesRequest) (*admin.ListRolesResponse, error) {
	roles, err := s.service.ListRoles(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	var protoRoles []*admin.Role
	for _, r := range roles {
		protoRoles = append(protoRoles, &admin.Role{
			Id:          r.ID,
			Name:        r.Name,
			Namespace:   r.Namespace,
			Permissions: r.Permissions,
		})
	}
	return &admin.ListRolesResponse{Roles: protoRoles}, nil
}

func (s *adminGRPCServer) DeleteRole(ctx context.Context, req *admin.DeleteRoleRequest) (*emptypb.Empty, error) {
	err := s.service.DeleteRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

type namespaceGRPCServer struct {
	admin.UnimplementedNamespaceServiceServer
	service adminservice.NamespaceService
}

func (s *namespaceGRPCServer) CreateNamespace(ctx context.Context, req *admin.CreateNamespaceRequest) (*admin.Namespace, error) {
	ns := &domain.Namespace{
		Name: req.Name,
		// Map config if needed...
	}
	newNS, err := s.service.CreateNamespace(ctx, ns)
	if err != nil {
		return nil, err
	}
	return mapNamespace(newNS), nil
}

func (s *namespaceGRPCServer) GetNamespace(ctx context.Context, req *admin.GetNamespaceRequest) (*admin.Namespace, error) {
	ns, err := s.service.GetNamespace(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return mapNamespace(ns), nil
}

func (s *namespaceGRPCServer) ListNamespaces(ctx context.Context, req *admin.ListNamespacesRequest) (*admin.ListNamespacesResponse, error) {
	// Simple page for now
	namespaces, _, err := s.service.ListNamespaces(ctx, 1, int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var protoNamespaces []*admin.Namespace
	for _, ns := range namespaces {
		protoNamespaces = append(protoNamespaces, mapNamespace(ns))
	}

	return &admin.ListNamespacesResponse{
		Namespaces: protoNamespaces,
	}, nil
}

func (s *namespaceGRPCServer) UpdateNamespaceConfig(ctx context.Context, req *admin.UpdateNamespaceConfigRequest) (*emptypb.Empty, error) {
	ns, err := s.service.GetNamespace(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	ns.Config.MFARequired = req.Config.MfaRequired
	ns.Config.AllowedSocialProviders = req.Config.AllowedSocialProviders
	if req.Config.PasswordPolicy != nil {
		ns.Config.PasswordPolicy.MinLength = int(req.Config.PasswordPolicy.MinLength)
		ns.Config.PasswordPolicy.RequireUppercase = req.Config.PasswordPolicy.RequireUppercase
		ns.Config.PasswordPolicy.RequireLowercase = req.Config.PasswordPolicy.RequireLowercase
		ns.Config.PasswordPolicy.RequireNumber = req.Config.PasswordPolicy.RequireNumber
		ns.Config.PasswordPolicy.RequireSpecial = req.Config.PasswordPolicy.RequireSpecial
	}

	err = s.service.UpdateNamespace(ctx, ns)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *namespaceGRPCServer) DeleteNamespace(ctx context.Context, req *admin.DeleteNamespaceRequest) (*emptypb.Empty, error) {
	err := s.service.DeleteNamespace(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func mapNamespace(ns *domain.Namespace) *admin.Namespace {
	return &admin.Namespace{
		Id:   ns.ID,
		Name: ns.Name,
		Config: &admin.NamespaceConfig{
			MfaRequired:            ns.Config.MFARequired,
			AllowedSocialProviders: ns.Config.AllowedSocialProviders,
			PasswordPolicy: &admin.PasswordPolicy{
				MinLength:        int32(ns.Config.PasswordPolicy.MinLength),
				RequireUppercase: ns.Config.PasswordPolicy.RequireUppercase,
				RequireLowercase: ns.Config.PasswordPolicy.RequireLowercase,
				RequireNumber:    ns.Config.PasswordPolicy.RequireNumber,
				RequireSpecial:   ns.Config.PasswordPolicy.RequireSpecial,
			},
		},
	}
}
func (s *adminGRPCServer) ListAuditLogs(ctx context.Context, req *admin.ListAuditLogsRequest) (*admin.ListAuditLogsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	page := int(req.PageToken)
	if page == 0 {
		page = 1
	}

	logs, total, err := s.service.ListAuditLogs(ctx, req.Namespace, page, pageSize)
	if err != nil {
		return nil, err
	}

	var protoLogs []*admin.AuditLog
	for _, l := range logs {
		metadataJSON := ""
		if l.Metadata != nil {
			if b, err := json.Marshal(l.Metadata); err == nil {
				metadataJSON = string(b)
			}
		}

		protoLogs = append(protoLogs, &admin.AuditLog{
			Id:           l.ID,
			Event:        string(l.Event),
			UserId:       l.UserID,
			Namespace:    l.Namespace,
			Ip:           l.IP,
			UserAgent:    l.UserAgent,
			MetadataJson: metadataJSON,
			Timestamp:    timestamppb.New(l.Timestamp),
		})
	}

	var nextPageToken int32
	if int64(page*pageSize) < total {
		nextPageToken = int32(page + 1)
	}

	return &admin.ListAuditLogsResponse{
		Logs:          protoLogs,
		NextPageToken: nextPageToken,
		TotalCount:    total,
	}, nil
}
