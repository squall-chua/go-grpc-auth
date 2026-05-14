package admin

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminService interface {
	ListUsers(ctx context.Context, namespace string, page, pageSize int) ([]*domain.User, int64, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, id, status string) error
	ResetUserPassword(ctx context.Context, id, newPassword string) error
	GrantRoles(ctx context.Context, id string, roles []string) error
	RevokeRoles(ctx context.Context, id string, roles []string) error
	GrantPermissions(ctx context.Context, id string, permissions []string) error
	RevokePermissions(ctx context.Context, id string, permissions []string) error

	// Role Management
	CreateRole(ctx context.Context, role *domain.Role) error
	ListRoles(ctx context.Context, namespace string) ([]*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error

	// Permission Management
	ListPermissions(ctx context.Context, namespace string) ([]*domain.Permission, error)
	DeletePermission(ctx context.Context, id string) error

	// Audit Logs
	ListAuditLogs(ctx context.Context, namespace string, page, pageSize int) ([]*domain.AuditLog, int64, error)
}

type adminService struct {
	userRepo   repository.UserRepository
	roleRepo   repository.RoleRepository
	permRepo   repository.PermissionRepository
	clientRepo OIDCClientService
	auditService audit.AuditService
}

func NewAdminService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	clientRepo OIDCClientService,
	auditService audit.AuditService,
) AdminService {
	return &adminService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		clientRepo:   clientRepo,
		auditService: auditService,
	}
}

func (s *adminService) ListUsers(ctx context.Context, namespace string, page, pageSize int) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, namespace, offset, pageSize)
}

func (s *adminService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return user, nil
}

func (s *adminService) UpdateUserStatus(ctx context.Context, id, userStatus string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	user.Status = userStatus
	return s.userRepo.Update(ctx, user)
}

func (s *adminService) ResetUserPassword(ctx context.Context, id, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return status.Error(codes.Internal, "failed to hash password")
	}

	user.PasswordHash = string(hashed)
	return s.userRepo.Update(ctx, user)
}

func (s *adminService) GrantRoles(ctx context.Context, id string, roles []string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	// Add roles if not already present
	roleMap := make(map[string]bool)
	for _, r := range user.Roles {
		roleMap[r] = true
	}
	for _, r := range roles {
		if !roleMap[r] {
			user.Roles = append(user.Roles, r)
		}
	}

	return s.userRepo.Update(ctx, user)
}

func (s *adminService) RevokeRoles(ctx context.Context, id string, roles []string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	revokeMap := make(map[string]bool)
	for _, r := range roles {
		revokeMap[r] = true
	}

	var newRoles []string
	for _, r := range user.Roles {
		if !revokeMap[r] {
			newRoles = append(newRoles, r)
		}
	}
	user.Roles = newRoles

	return s.userRepo.Update(ctx, user)
}

func (s *adminService) GrantPermissions(ctx context.Context, id string, permissions []string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	permMap := make(map[string]bool)
	for _, p := range user.Permissions {
		permMap[p] = true
	}
	for _, p := range permissions {
		if !permMap[p] {
			user.Permissions = append(user.Permissions, p)
		}
	}

	return s.userRepo.Update(ctx, user)
}

func (s *adminService) RevokePermissions(ctx context.Context, id string, permissions []string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	revokeMap := make(map[string]bool)
	for _, p := range permissions {
		revokeMap[p] = true
	}

	var newPerms []string
	for _, p := range user.Permissions {
		if !revokeMap[p] {
			newPerms = append(newPerms, p)
		}
	}
	user.Permissions = newPerms

	return s.userRepo.Update(ctx, user)
}

func (s *adminService) CreateRole(ctx context.Context, role *domain.Role) error {
	return s.roleRepo.Create(ctx, role)
}

func (s *adminService) ListRoles(ctx context.Context, namespace string) ([]*domain.Role, error) {
	return s.roleRepo.List(ctx, namespace)
}

func (s *adminService) DeleteRole(ctx context.Context, id string) error {
	return s.roleRepo.Delete(ctx, id)
}

func (s *adminService) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return s.permRepo.Create(ctx, perm)
}

func (s *adminService) ListPermissions(ctx context.Context, namespace string) ([]*domain.Permission, error) {
	return s.permRepo.List(ctx, namespace)
}

func (s *adminService) DeletePermission(ctx context.Context, id string) error {
	return s.permRepo.Delete(ctx, id)
}

func (s *adminService) ListAuditLogs(ctx context.Context, namespace string, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.auditService.List(ctx, namespace, offset, pageSize)
}
