package authclient

import (
	"context"
	"sync"
	"time"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/pkg/interceptor/client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	Auth       auth.AuthServiceClient
	Admin      admin.AdminServiceClient
	OIDC       auth.OIDCServiceClient
	OIDCClient admin.OIDCClientServiceClient
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
		Auth:       auth.NewAuthServiceClient(conn),
		Admin:      admin.NewAdminServiceClient(conn),
		OIDC:       auth.NewOIDCServiceClient(conn),
		OIDCClient: admin.NewOIDCClientServiceClient(conn),
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

func (c *Client) Logout(ctx context.Context) error {
	_, err := c.Auth.Logout(ctx, &emptypb.Empty{})
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

func (c *Client) ListUsers(ctx context.Context, namespace string, pageSize, page int32) (*admin.ListUsersResponse, error) {
	return c.Admin.ListUsers(ctx, &admin.ListUsersRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		Page:      page,
	})
}

func (c *Client) GetUser(ctx context.Context, id string) (*admin.User, error) {
	return c.Admin.GetUser(ctx, &admin.GetUserRequest{Id: id})
}

func (c *Client) UpdateUserStatus(ctx context.Context, id string, status admin.UserStatus) error {
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

func (c *Client) ListAuditLogs(ctx context.Context, namespace string, pageSize, page int32) (*admin.ListAuditLogsResponse, error) {
	return c.Admin.ListAuditLogs(ctx, &admin.ListAuditLogsRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		Page:      page,
	})
}

func (c *Client) CreateRole(ctx context.Context, name, namespace string, permissions []string) (*admin.Role, error) {
	return c.Admin.CreateRole(ctx, &admin.CreateRoleRequest{
		Name:        name,
		Namespace:   namespace,
		Permissions: permissions,
	})
}

func (c *Client) ListRoles(ctx context.Context, query string, pageSize, page int32) (*admin.ListRolesResponse, error) {
	return c.Admin.ListRoles(ctx, &admin.ListRolesRequest{
		Query:    query,
		PageSize: pageSize,
		Page:     page,
	})
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	_, err := c.Admin.DeleteRole(ctx, &admin.DeleteRoleRequest{Id: id})
	return err
}

func (c *Client) RegisterOIDCClient(ctx context.Context, name, namespace string, redirectUris, allowedScopes []string, skipConsent bool) (*admin.OIDCClient, error) {
	return c.OIDCClient.RegisterClient(ctx, &admin.RegisterClientRequest{
		Name:          name,
		Namespace:     namespace,
		RedirectUris:  redirectUris,
		AllowedScopes: allowedScopes,
		SkipConsent:   skipConsent,
	})
}

func (c *Client) GetOIDCClient(ctx context.Context, clientID string) (*admin.OIDCClient, error) {
	return c.OIDCClient.GetClient(ctx, &admin.GetClientRequest{ClientId: clientID})
}

func (c *Client) UpdateOIDCClient(ctx context.Context, req *admin.UpdateClientRequest) (*admin.OIDCClient, error) {
	return c.OIDCClient.UpdateClient(ctx, req)
}

func (c *Client) DeleteOIDCClient(ctx context.Context, clientID string) error {
	_, err := c.OIDCClient.DeleteClient(ctx, &admin.DeleteClientRequest{ClientId: clientID})
	return err
}

func (c *Client) ListOIDCClients(ctx context.Context, namespace string, pageSize, page int32) (*admin.ListClientsResponse, error) {
	return c.OIDCClient.ListClients(ctx, &admin.ListClientsRequest{
		Namespace: namespace,
		PageSize:  pageSize,
		Page:      page,
	})
}

func (c *Client) RotateOIDCClientSecret(ctx context.Context, clientID string) (string, error) {
	resp, err := c.OIDCClient.RotateClientSecret(ctx, &admin.RotateClientSecretRequest{ClientId: clientID})
	if err != nil {
		return "", err
	}
	return resp.ClientSecret, nil
}

// OIDC Methods

func (c *Client) GetUserInfo(ctx context.Context) (*auth.UserInfo, error) {
	return c.OIDC.GetUserInfo(ctx, &auth.GetUserInfoRequest{})
}

// ClientCredentialsProvider implements client.TokenProvider for M2M communication
type ClientCredentialsProvider struct {
	oidcClient   auth.OIDCServiceClient
	clientID     string
	clientSecret string

	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

func NewClientCredentialsProvider(oidcClient auth.OIDCServiceClient, clientID, clientSecret string) *ClientCredentialsProvider {
	return &ClientCredentialsProvider{
		oidcClient:   oidcClient,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (p *ClientCredentialsProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.RLock()
	if p.token != "" && time.Now().UTC().Before(p.expiresAt.Add(-1*time.Minute)) {
		defer p.mu.RUnlock()
		return p.token, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check
	if p.token != "" && time.Now().UTC().Before(p.expiresAt.Add(-1*time.Minute)) {
		return p.token, nil
	}

	resp, err := p.oidcClient.Token(ctx, &auth.TokenRequest{
		GrantType:    "client_credentials",
		ClientId:     p.clientID,
		ClientSecret: p.clientSecret,
	})
	if err != nil {
		return "", err
	}

	p.token = resp.AccessToken
	p.expiresAt = time.Now().UTC().Add(time.Duration(resp.ExpiresIn) * time.Second)

	return p.token, nil
}
