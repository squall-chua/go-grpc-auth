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
	ListUsers(ctx context.Context, namespace, query, status string, page, pageSize int) ([]*domain.User, int64, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, id string, status domain.UserStatus) error
	ResetUserPassword(ctx context.Context, id, newPassword string) error
	GrantRoles(ctx context.Context, id string, roles []string) error
	RevokeRoles(ctx context.Context, id string, roles []string) error
	GrantPermissions(ctx context.Context, id string, permissions []string) error
	RevokePermissions(ctx context.Context, id string, permissions []string) error

	// Role Management
	CreateRole(ctx context.Context, role *domain.Role) error
	ListRoles(ctx context.Context, namespace string, page, pageSize int) ([]*domain.Role, int64, error)
	DeleteRole(ctx context.Context, id string) error

	// Permission Management
	CreatePermission(ctx context.Context, perm *domain.Permission) error
	ListPermissions(ctx context.Context, namespace string, page, pageSize int) ([]*domain.Permission, int64, error)
	DeletePermission(ctx context.Context, id string) error

	// Audit Logs
	ListAuditLogs(ctx context.Context, filter audit.AuditListFilter, page, pageSize int) ([]*domain.AuditLog, int64, error)
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

func (s *adminService) ListUsers(ctx context.Context, namespace, query, status string, page, pageSize int) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, namespace, offset, pageSize, repository.UserListFilter{
		Query:  query,
		Status: status,
	})
}

func (s *adminService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return user, nil
}

func (s *adminService) UpdateUserStatus(ctx context.Context, id string, userStatus domain.UserStatus) error {
	return s.userRepo.UpdateStatus(ctx, id, userStatus)
}

func (s *adminService) ResetUserPassword(ctx context.Context, id, newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return status.Error(codes.Internal, "failed to hash password")
	}
	return s.userRepo.UpdatePassword(ctx, id, string(hashed), 0)
}

func (s *adminService) GrantRoles(ctx context.Context, id string, roles []string) error {
	return s.userRepo.AddRoles(ctx, id, roles)
}

func (s *adminService) RevokeRoles(ctx context.Context, id string, roles []string) error {
	return s.userRepo.RemoveRoles(ctx, id, roles)
}

func (s *adminService) GrantPermissions(ctx context.Context, id string, permissions []string) error {
	return s.userRepo.AddPermissions(ctx, id, permissions)
}

func (s *adminService) RevokePermissions(ctx context.Context, id string, permissions []string) error {
	return s.userRepo.RemovePermissions(ctx, id, permissions)
}

func (s *adminService) CreateRole(ctx context.Context, role *domain.Role) error {
	return s.roleRepo.Create(ctx, role)
}

func (s *adminService) ListRoles(ctx context.Context, namespace string, page, pageSize int) ([]*domain.Role, int64, error) {
	offset := (page - 1) * pageSize
	return s.roleRepo.List(ctx, namespace, offset, pageSize)
}

func (s *adminService) DeleteRole(ctx context.Context, id string) error {
	return s.roleRepo.Delete(ctx, id)
}

func (s *adminService) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return s.permRepo.Create(ctx, perm)
}

func (s *adminService) ListPermissions(ctx context.Context, namespace string, page, pageSize int) ([]*domain.Permission, int64, error) {
	offset := (page - 1) * pageSize
	return s.permRepo.List(ctx, namespace, offset, pageSize)
}

func (s *adminService) DeletePermission(ctx context.Context, id string) error {
	return s.permRepo.Delete(ctx, id)
}

func (s *adminService) ListAuditLogs(ctx context.Context, filter audit.AuditListFilter, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	return s.auditService.List(ctx, filter, page, pageSize)
}
