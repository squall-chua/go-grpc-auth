package authclient

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/pkg/interceptor/client"
	"google.golang.org/grpc"
)

type Client struct {
	Auth  auth.AuthServiceClient
	Admin admin.AdminServiceClient
}

// NewClient creates a new auth client. If provider is nil, no auth interceptor is added.
func NewClient(target string, provider client.TokenProvider, opts ...grpc.DialOption) (*Client, error) {
	if provider != nil {
		opts = append(opts, grpc.WithChainUnaryInterceptor(client.UnaryAuthInterceptor(provider)))
		opts = append(opts, grpc.WithChainStreamInterceptor(client.StreamAuthInterceptor(provider)))
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Auth:  auth.NewAuthServiceClient(conn),
		Admin: admin.NewAdminServiceClient(conn),
	}, nil
}

// AuthService Methods

func (c *Client) Register(ctx context.Context, email, username, password, namespace string) (*auth.TokenPair, error) {
	return c.Auth.Register(ctx, &auth.RegisterRequest{
		Email:     email,
		Username:  username,
		Password:  password,
		Namespace: namespace,
	})
}

func (c *Client) Login(ctx context.Context, login, password, namespace string) (*auth.TokenPair, error) {
	return c.Auth.Login(ctx, &auth.LoginRequest{
		Login:     login,
		Password:  password,
		Namespace: namespace,
	})
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	return c.Auth.RefreshToken(ctx, &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
}

func (c *Client) Logout(ctx context.Context, accessToken, refreshToken string) error {
	_, err := c.Auth.Logout(ctx, &auth.LogoutRequest{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	return err
}

func (c *Client) ChangePassword(ctx context.Context, current, new string) error {
	_, err := c.Auth.ChangePassword(ctx, &auth.ChangePasswordRequest{
		CurrentPassword: current,
		NewPassword:     new,
	})
	return err
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*auth.Principal, error) {
	return c.Auth.ValidateToken(ctx, &auth.ValidateTokenRequest{
		Token: token,
	})
}

func (c *Client) InitiateMFA(ctx context.Context, mfaToken, method string) (*auth.InitiateMFAResponse, error) {
	return c.Auth.InitiateMFA(ctx, &auth.InitiateMFARequest{
		MfaToken: mfaToken,
		Method:   method,
	})
}

func (c *Client) VerifyMFA(ctx context.Context, mfaToken, code string) (*auth.TokenPair, error) {
	return c.Auth.VerifyMFA(ctx, &auth.VerifyMFARequest{
		MfaToken: mfaToken,
		Code:     code,
	})
}

// AdminService Methods

func (c *Client) ListUsers(ctx context.Context, namespace string, pageSize int32, pageToken string) (*admin.ListUsersResponse, error) {
	return c.Admin.ListUsers(ctx, &admin.ListUsersRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
}

func (c *Client) GetUser(ctx context.Context, id string) (*admin.User, error) {
	return c.Admin.GetUser(ctx, &admin.GetUserRequest{Id: id})
}

func (c *Client) UpdateUserStatus(ctx context.Context, id, status string) error {
	_, err := c.Admin.UpdateUserStatus(ctx, &admin.UpdateUserStatusRequest{
		Id:     id,
		Status: status,
	})
	return err
}

func (c *Client) ResetUserPassword(ctx context.Context, id, newPassword string) error {
	_, err := c.Admin.ResetUserPassword(ctx, &admin.ResetUserPasswordRequest{
		Id:          id,
		NewPassword: newPassword,
	})
	return err
}

func (c *Client) GrantRoles(ctx context.Context, id string, roles []string) error {
	_, err := c.Admin.GrantRoles(ctx, &admin.GrantRolesRequest{
		Id:    id,
		Roles: roles,
	})
	return err
}

func (c *Client) RevokeRoles(ctx context.Context, id string, roles []string) error {
	_, err := c.Admin.RevokeRoles(ctx, &admin.RevokeRolesRequest{
		Id:    id,
		Roles: roles,
	})
	return err
}

func (c *Client) GrantPermissions(ctx context.Context, id string, permissions []string) error {
	_, err := c.Admin.GrantPermissions(ctx, &admin.GrantPermissionsRequest{
		Id:          id,
		Permissions: permissions,
	})
	return err
}

func (c *Client) RevokePermissions(ctx context.Context, id string, permissions []string) error {
	_, err := c.Admin.RevokePermissions(ctx, &admin.RevokePermissionsRequest{
		Id:          id,
		Permissions: permissions,
	})
	return err
}

func (c *Client) ListAuditLogs(ctx context.Context, namespace string, pageSize, pageToken int32) (*admin.ListAuditLogsResponse, error) {
	return c.Admin.ListAuditLogs(ctx, &admin.ListAuditLogsRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
}

func (c *Client) CreateRole(ctx context.Context, name, namespace string, permissions []string) (*admin.Role, error) {
	return c.Admin.CreateRole(ctx, &admin.CreateRoleRequest{
		Name:        name,
		Namespace:   namespace,
		Permissions: permissions,
	})
}

func (c *Client) ListRoles(ctx context.Context, namespace string) (*admin.ListRolesResponse, error) {
	return c.Admin.ListRoles(ctx, &admin.ListRolesRequest{Namespace: namespace})
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	_, err := c.Admin.DeleteRole(ctx, &admin.DeleteRoleRequest{Id: id})
	return err
}

func (c *Client) CreateServiceAccount(ctx context.Context, name, namespace string, scopes []string) (*admin.ServiceAccount, error) {
	return c.Admin.CreateServiceAccount(ctx, &admin.CreateServiceAccountRequest{
		Name:      name,
		Namespace: namespace,
		Scopes:    scopes,
	})
}

func (c *Client) ListServiceAccounts(ctx context.Context, namespace string) (*admin.ListServiceAccountsResponse, error) {
	return c.Admin.ListServiceAccounts(ctx, &admin.ListServiceAccountsRequest{Namespace: namespace})
}
