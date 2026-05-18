package server

import (
	"context"
	"strings"

	"encoding/json"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	adminservice "github.com/squall-chua/go-grpc-auth/internal/service/admin"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adminGRPCServer struct {
	admin.UnimplementedAdminServiceServer
	service       adminservice.AdminService
	clientService adminservice.OIDCClientService
}

func (s *adminGRPCServer) ListUsers(ctx context.Context, req *admin.ListUsersRequest) (*admin.ListUsersResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	users, total, err := s.service.ListUsers(ctx, req.Namespace, req.Query, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	var protoUsers []*admin.User
	for _, u := range users {
		protoUsers = append(protoUsers, &admin.User{
			Id:          u.ID.Hex(),
			Email:       u.Email,
			Username:    u.Username,
			Namespace:   u.Namespace,
			Status:      mapUserStatus(u.Status),
			Roles:       u.Roles,
			Permissions: u.Permissions,
			CreatedAt:   timestamppb.New(u.CreatedAt),
			UpdatedAt:   timestamppb.New(u.UpdatedAt),
		})
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &admin.ListUsersResponse{
		Users:      protoUsers,
		TotalPages: totalPages,
		TotalCount: int32(total),
	}, nil
}

func (s *adminGRPCServer) GetUser(ctx context.Context, req *admin.GetUserRequest) (*admin.User, error) {
	u, err := s.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &admin.User{
		Id:          u.ID.Hex(),
		Email:       u.Email,
		Username:    u.Username,
		Namespace:   u.Namespace,
		Status:      mapUserStatus(u.Status),
		Roles:       u.Roles,
		Permissions: u.Permissions,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}, nil
}

func (s *adminGRPCServer) UpdateUserStatus(ctx context.Context, req *admin.UpdateUserStatusRequest) (*emptypb.Empty, error) {
	err := s.service.UpdateUserStatus(ctx, req.Id, mapProtoUserStatus(req.Status))
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
		Id:          role.ID.Hex(),
		Name:        role.Name,
		Namespace:   role.Namespace,
		Permissions: role.Permissions,
	}, nil
}

func (s *adminGRPCServer) ListRoles(ctx context.Context, req *admin.ListRolesRequest) (*admin.ListRolesResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	roles, total, err := s.service.ListRoles(ctx, req.Namespace, page, pageSize)
	if err != nil {
		return nil, err
	}
	var protoRoles []*admin.Role
	for _, r := range roles {
		protoRoles = append(protoRoles, &admin.Role{
			Id:          r.ID.Hex(),
			Name:        r.Name,
			Namespace:   r.Namespace,
			Permissions: r.Permissions,
		})
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return &admin.ListRolesResponse{
		Roles:      protoRoles,
		TotalPages: totalPages,
		TotalCount: int32(total),
	}, nil
}

func (s *adminGRPCServer) DeleteRole(ctx context.Context, req *admin.DeleteRoleRequest) (*emptypb.Empty, error) {
	err := s.service.DeleteRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *adminGRPCServer) CreatePermission(ctx context.Context, req *admin.CreatePermissionRequest) (*admin.Permission, error) {
	perm := &domain.Permission{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Description: req.Description,
	}
	err := s.service.CreatePermission(ctx, perm)
	if err != nil {
		return nil, err
	}
	return &admin.Permission{
		Id:          perm.ID.Hex(),
		Name:        perm.Name,
		Namespace:   perm.Namespace,
		Description: perm.Description,
	}, nil
}

func (s *adminGRPCServer) ListPermissions(ctx context.Context, req *admin.ListPermissionsRequest) (*admin.ListPermissionsResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	perms, total, err := s.service.ListPermissions(ctx, req.Namespace, page, pageSize)
	if err != nil {
		return nil, err
	}
	var protoPerms []*admin.Permission
	for _, p := range perms {
		protoPerms = append(protoPerms, &admin.Permission{
			Id:          p.ID.Hex(),
			Name:        p.Name,
			Namespace:   p.Namespace,
			Description: p.Description,
		})
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return &admin.ListPermissionsResponse{
		Permissions: protoPerms,
		TotalPages:  totalPages,
		TotalCount:  int32(total),
	}, nil
}

func (s *adminGRPCServer) DeletePermission(ctx context.Context, req *admin.DeletePermissionRequest) (*emptypb.Empty, error) {
	err := s.service.DeletePermission(ctx, req.Id)
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
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	namespaces, total, err := s.service.ListNamespaces(ctx, req.Query, page, pageSize)
	if err != nil {
		return nil, err
	}

	var protoNamespaces []*admin.Namespace
	for _, ns := range namespaces {
		protoNamespaces = append(protoNamespaces, mapNamespace(ns))
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &admin.ListNamespacesResponse{
		Namespaces: protoNamespaces,
		TotalPages: totalPages,
		TotalCount: int32(total),
	}, nil
}

func (s *namespaceGRPCServer) UpdateNamespaceConfig(ctx context.Context, req *admin.UpdateNamespaceConfigRequest) (*emptypb.Empty, error) {
	ns, err := s.service.GetNamespace(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	ns.Config.MFARequired = req.Config.GetMfaRequired()
	ns.Config.AllowedSocialProviders = req.Config.GetAllowedSocialProviders()
	pp := req.Config.GetPasswordPolicy()
	ns.Config.PasswordPolicy.MinLength = int(pp.GetMinLength())
	ns.Config.PasswordPolicy.RequireUppercase = pp.GetRequireUppercase()
	ns.Config.PasswordPolicy.RequireLowercase = pp.GetRequireLowercase()
	ns.Config.PasswordPolicy.RequireNumber = pp.GetRequireNumber()
	ns.Config.PasswordPolicy.RequireSpecial = pp.GetRequireSpecial()
	ns.Config.PasswordPolicy.PasswordHistory = int(pp.GetPasswordHistory())

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
		Id:   ns.ID.Hex(),
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
				PasswordHistory:  int32(ns.Config.PasswordPolicy.PasswordHistory),
			},
		},
	}
}
func (s *adminGRPCServer) ListAuditLogs(ctx context.Context, req *admin.ListAuditLogsRequest) (*admin.ListAuditLogsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	page := int(req.Page)
	if page == 0 {
		page = 1
	}

	filter := audit.AuditListFilter{
		Namespace: req.Namespace,
		Event:     req.Event,
		UserID:    req.UserId,
	}
	if req.From != nil {
		filter.From = req.From.AsTime()
	}
	if req.To != nil {
		filter.To = req.To.AsTime()
	}

	logs, total, err := s.service.ListAuditLogs(ctx, filter, page, pageSize)
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
			Id:           l.ID.Hex(),
			Event:        string(l.Event),
			UserId:       l.UserID,
			Namespace:    l.Namespace,
			Ip:           l.IP,
			UserAgent:    l.UserAgent,
			MetadataJson: metadataJSON,
			Timestamp:    timestamppb.New(l.Timestamp),
		})
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &admin.ListAuditLogsResponse{
		Logs:       protoLogs,
		TotalPages: totalPages,
		TotalCount: int32(total),
	}, nil
}

func mapUserStatus(status domain.UserStatus) admin.UserStatus {
	key := "USER_STATUS_" + strings.ToUpper(string(status))
	if val, ok := admin.UserStatus_value[key]; ok {
		return admin.UserStatus(val)
	}
	return admin.UserStatus_USER_STATUS_UNSPECIFIED
}

func mapProtoUserStatus(status admin.UserStatus) domain.UserStatus {
	name := admin.UserStatus_name[int32(status)]
	return domain.UserStatus(strings.ToLower(strings.TrimPrefix(name, "USER_STATUS_")))
}
