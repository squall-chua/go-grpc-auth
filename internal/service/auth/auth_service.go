package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/pkg/ctxutil"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	"github.com/squall-chua/go-grpc-auth/internal/service/ratelimit"
	"github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/service/webhook"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(ctx context.Context, email, username, password, namespace string) (*domain.TokenPair, error)
	Login(ctx context.Context, login, password, namespace string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ValidateToken(ctx context.Context, token string) (*domain.Principal, error)
	
	InitiateMFA(ctx context.Context, mfaToken, method string) (secret, qrCodeURL string, err error)
	VerifyMFA(ctx context.Context, mfaToken, code string) (*domain.TokenPair, error)
}

type authService struct {
	userRepo     repository.UserRepository
	tokenService token.TokenService
	nsRepo       repository.NamespaceRepository
	mfaService   MFAService
	auditService audit.AuditService
	rateLimiter  ratelimit.RateLimiter
	webhookSvc   webhook.WebhookService
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenService token.TokenService,
	nsRepo repository.NamespaceRepository,
	mfaService MFAService,
	auditService audit.AuditService,
	rateLimiter ratelimit.RateLimiter,
	webhookSvc webhook.WebhookService,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
		nsRepo:       nsRepo,
		mfaService:   mfaService,
		auditService: auditService,
		rateLimiter:  rateLimiter,
		webhookSvc:   webhookSvc,
	}
}

func (s *authService) Register(ctx context.Context, email, username, password, namespace string) (*domain.TokenPair, error) {
	ip := ctxutil.GetClientIP(ctx)
	if ok, _ := s.rateLimiter.Allow(ctx, "reg:"+ip); !ok {
		return nil, status.Error(codes.ResourceExhausted, "too many registration attempts")
	}

	if namespace == "" {
		namespace = "default"
	}

	ns, err := s.nsRepo.GetByName(ctx, namespace)
	if err == nil {
		if err := ValidateIP(ip, ns.Config.IPAllowlist, ns.Config.IPDenylist); err != nil {
			return nil, err
		}
		if err := ValidatePassword(password, ns.Config.PasswordPolicy); err != nil {
			return nil, err
		}
	}

	// Check if user already exists
	if email != "" {
		_, err := s.userRepo.GetByEmail(ctx, namespace, email)
		if err == nil {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
	}
	if username != "" {
		_, err := s.userRepo.GetByUsername(ctx, namespace, username)
		if err == nil {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		Namespace:    namespace,
		Status:       "active",
		Roles:        []string{"user"},
		PasswordHistory: []string{string(hash)},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	s.auditService.Log(ctx, domain.EventRegisterSuccess, user.ID, namespace, ctxutil.GetClientIP(ctx), ctxutil.GetUserAgent(ctx), nil)
	s.webhookSvc.Notify(ctx, namespace, domain.EventRegisterSuccess, map[string]string{"user_id": user.ID, "username": user.Username})

	// Check if MFA is required for this namespace
	ns, err = s.nsRepo.GetByName(ctx, namespace)
	if err == nil && ns.Config.MFARequired {
		mfaToken, err := s.mfaService.CreateMFAToken(ctx, user.ID, namespace, domain.MFAMethodTOTP)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create MFA token")
		}
		s.auditService.Log(ctx, domain.EventMFAChallenge, user.ID, namespace, ctxutil.GetClientIP(ctx), ctxutil.GetUserAgent(ctx), nil)
		return &domain.TokenPair{
			MFARequired: true,
			MFAToken:    mfaToken,
		}, nil
	}

	return s.tokenService.GenerateTokenPair(ctx, user)
}

func (s *authService) Login(ctx context.Context, login, password, namespace string) (*domain.TokenPair, error) {
	if namespace == "" {
		namespace = "default"
	}

	ip := ctxutil.GetClientIP(ctx)
	ua := ctxutil.GetUserAgent(ctx)

	if ok, _ := s.rateLimiter.Allow(ctx, "login:"+ip); !ok {
		s.auditService.Log(ctx, domain.EventLoginFailed, "", namespace, ip, ua, map[string]string{"login": login, "reason": "rate_limit_exceeded"})
		return nil, status.Error(codes.ResourceExhausted, "too many login attempts")
	}

	var user *domain.User
	var err error

	// Try email first
	user, err = s.userRepo.GetByEmail(ctx, namespace, login)
	if err != nil {
		// Try username
		user, err = s.userRepo.GetByUsername(ctx, namespace, login)
		if err != nil {
			s.auditService.Log(ctx, domain.EventLoginFailed, "", namespace, ip, ua, map[string]string{"login": login, "reason": "user_not_found"})
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
	}

	if user.Status != "active" {
		s.auditService.Log(ctx, domain.EventLoginFailed, user.ID, namespace, ip, ua, map[string]string{"reason": "user_not_active"})
		return nil, status.Error(codes.PermissionDenied, "user is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.auditService.Log(ctx, domain.EventLoginFailed, user.ID, namespace, ip, ua, map[string]string{"reason": "invalid_password"})
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Check if MFA is required for this namespace
	ns, err := s.nsRepo.GetByName(ctx, namespace)
	if err == nil {
		if err := ValidateIP(ip, ns.Config.IPAllowlist, ns.Config.IPDenylist); err != nil {
			s.auditService.Log(ctx, domain.EventLoginFailed, user.ID, namespace, ip, ua, map[string]string{"reason": "ip_blocked"})
			return nil, err
		}
	}

	if err == nil && ns.Config.MFARequired {
		mfaToken, err := s.mfaService.CreateMFAToken(ctx, user.ID, namespace, domain.MFAMethodTOTP)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create MFA token")
		}
		s.auditService.Log(ctx, domain.EventMFAChallenge, user.ID, namespace, ip, ua, nil)
		return &domain.TokenPair{
			MFARequired: true,
			MFAToken:    mfaToken,
		}, nil
	}

	s.auditService.Log(ctx, domain.EventLoginSuccess, user.ID, namespace, ip, ua, nil)
	return s.tokenService.GenerateTokenPair(ctx, user)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return s.tokenService.RefreshToken(ctx, refreshToken)
}

func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	return s.tokenService.RevokeTokens(ctx, accessToken, refreshToken)
}

func (s *authService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return status.Error(codes.NotFound, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return status.Error(codes.Unauthenticated, "invalid current password")
	}

	ns, err := s.nsRepo.GetByName(ctx, user.Namespace)
	if err == nil {
		if err := ValidatePassword(newPassword, ns.Config.PasswordPolicy); err != nil {
			return err
		}
		// Check history
		for _, h := range user.PasswordHistory {
			if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(newPassword)); err == nil {
				return status.Error(codes.InvalidArgument, "password has been used before")
			}
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return status.Error(codes.Internal, "failed to hash password")
	}

	user.PasswordHash = string(hash)
	user.PasswordHistory = append(user.PasswordHistory, string(hash))
	// Keep only the last N
	if err == nil && ns.Config.PasswordPolicy.PasswordHistory > 0 {
		if len(user.PasswordHistory) > ns.Config.PasswordPolicy.PasswordHistory {
			user.PasswordHistory = user.PasswordHistory[len(user.PasswordHistory)-ns.Config.PasswordPolicy.PasswordHistory:]
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return status.Error(codes.Internal, "failed to update password")
	}

	return nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.Principal, error) {
	return s.tokenService.ValidateAccessToken(ctx, token)
}

func (s *authService) InitiateMFA(ctx context.Context, mfaTokenStr, method string) (string, string, error) {
	token, err := s.mfaService.VerifyMFAToken(ctx, mfaTokenStr)
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, "invalid MFA token")
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return "", "", status.Error(codes.NotFound, "user not found")
	}

	return s.mfaService.InitiateTOTP(ctx, user.ID, "GoGrpcAuth", user.Email)
}

func (s *authService) VerifyMFA(ctx context.Context, mfaTokenStr, code string) (*domain.TokenPair, error) {
	token, err := s.mfaService.VerifyMFAToken(ctx, mfaTokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid MFA token")
	}

	valid, err := s.mfaService.VerifyTOTP(ctx, token.UserID, code)
	if err != nil || !valid {
		s.auditService.Log(ctx, domain.EventMFAFailed, token.UserID, token.Namespace, ctxutil.GetClientIP(ctx), ctxutil.GetUserAgent(ctx), nil)
		return nil, status.Error(codes.Unauthenticated, "invalid MFA code")
	}

	// MFA successful, issue final tokens
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	s.auditService.Log(ctx, domain.EventMFAVerified, user.ID, user.Namespace, ctxutil.GetClientIP(ctx), ctxutil.GetUserAgent(ctx), nil)
	s.auditService.Log(ctx, domain.EventLoginSuccess, user.ID, user.Namespace, ctxutil.GetClientIP(ctx), ctxutil.GetUserAgent(ctx), map[string]string{"mfa": "verified"})
	
	return s.tokenService.GenerateTokenPair(ctx, user)
}
