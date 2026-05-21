package auth

import (
	"context"
	"errors"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
	"github.com/squall-chua/go-grpc-auth/internal/service/ratelimit"
	"github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/service/webhook"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(ctx context.Context, email, username, password, namespace string) (*domain.TokenPair, error)
	Login(ctx context.Context, login, password, namespace string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, userID string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ValidateToken(ctx context.Context, token string) (*domain.Principal, error)

	InitiateMFA(ctx context.Context, mfaToken, method string) (secret, qrCodeURL, maskedRecipient string, err error)
	VerifyMFA(ctx context.Context, mfaToken, code string) (*domain.TokenPair, error)

	InitiateMFAForUser(ctx context.Context, userID, method string) (secret, qrCodeURL, maskedRecipient string, err error)
	VerifyMFAForUser(ctx context.Context, userID, code string) error
	ListMFAMethods(ctx context.Context, userID string) ([]MFAMethodStatus, error)
	EnableMFAMethod(ctx context.Context, userID, method string) error
	RemoveMFAMethod(ctx context.Context, userID string, method domain.MFAMethod) error
}

type authService struct {
	issuer       string
	userRepo     repository.UserRepository
	tokenService token.TokenService
	nsRepo       repository.NamespaceRepository
	mfaService   MFAService
	auditService audit.AuditService
	rateLimiter  ratelimit.RateLimiter
	webhookSvc   webhook.WebhookService
}

func NewAuthService(
	issuer string,
	userRepo repository.UserRepository,
	tokenService token.TokenService,
	nsRepo repository.NamespaceRepository,
	mfaService MFAService,
	auditService audit.AuditService,
	rateLimiter ratelimit.RateLimiter,
	webhookSvc webhook.WebhookService,
) AuthService {
	return &authService{
		issuer:       issuer,
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
	ip := util.GetClientIP(ctx)
	if ok, _ := s.rateLimiter.Allow(ctx, "reg:"+ip); !ok {
		return nil, status.Error(codes.ResourceExhausted, "too many registration attempts")
	}

	if namespace == "" {
		namespace = "default"
	}

	ns, err := s.nsRepo.GetByName(ctx, namespace)
	if err == nil {
		if err := ValidateIP(ip, ns.Config.IPAllowList, ns.Config.IPDenyList); err != nil {
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
		ID:              bson.NewObjectID(),
		Email:           email,
		Username:        username,
		PasswordHash:    string(hash),
		Namespace:       namespace,
		Status:          domain.UserStatusActive,
		Roles:           []string{"user"},
		PasswordHistory: []string{string(hash)},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	s.auditService.Log(ctx, domain.EventRegisterSuccess, user.ID.Hex(), namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), nil)
	s.webhookSvc.Notify(ctx, namespace, domain.EventRegisterSuccess, map[string]string{"user_id": user.ID.Hex(), "username": user.Username})

	if pair, err := s.checkMFAChallenge(ctx, user, ns, util.GetClientIP(ctx), util.GetUserAgent(ctx)); pair != nil || err != nil {
		return pair, err
	}

	return s.tokenService.GenerateTokenPair(ctx, user, "", nil)
}

func (s *authService) Login(ctx context.Context, login, password, namespace string) (*domain.TokenPair, error) {
	if namespace == "" {
		namespace = "default"
	}

	ip := util.GetClientIP(ctx)
	ua := util.GetUserAgent(ctx)

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

	if user.Status != domain.UserStatusActive {
		s.auditService.Log(ctx, domain.EventLoginFailed, user.ID.Hex(), namespace, ip, ua, map[string]string{"reason": "user_not_active"})
		return nil, status.Error(codes.PermissionDenied, "user is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.auditService.Log(ctx, domain.EventLoginFailed, user.ID.Hex(), namespace, ip, ua, map[string]string{"reason": "invalid_password"})
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Check if MFA is required for this namespace
	ns, err := s.nsRepo.GetByName(ctx, namespace)
	if err == nil {
		if err := ValidateIP(ip, ns.Config.IPAllowList, ns.Config.IPDenyList); err != nil {
			s.auditService.Log(ctx, domain.EventLoginFailed, user.ID.Hex(), namespace, ip, ua, map[string]string{"reason": "ip_blocked"})
			return nil, err
		}
	}

	if pair, err := s.checkMFAChallenge(ctx, user, ns, ip, ua); pair != nil || err != nil {
		return pair, err
	}

	s.auditService.Log(ctx, domain.EventLoginSuccess, user.ID.Hex(), namespace, ip, ua, nil)
	return s.tokenService.GenerateTokenPair(ctx, user, "", nil)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return s.tokenService.RefreshToken(ctx, refreshToken)
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	return s.tokenService.RevokeAllForUser(ctx, userID)
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

	if err := s.userRepo.UpdatePassword(ctx, userID, string(hash), ns.Config.PasswordPolicy.PasswordHistory); err != nil {
		return status.Error(codes.Internal, "failed to update password")
	}

	return nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.Principal, error) {
	return s.tokenService.ValidateAccessToken(ctx, token)
}

// checkMFAChallenge decides whether to issue an MFA challenge based on the
// namespace's MFA policy and the user's enrolled methods.
//   - "required": always challenge (even if no methods enrolled — the MFA page
//     will show available methods to set up)
//   - "optional": challenge only if the user has at least one enrolled method
//   - "disabled" / "": skip MFA
//
// Returns (nil, nil) when no challenge is needed.
func (s *authService) checkMFAChallenge(ctx context.Context, user *domain.User, ns *domain.Namespace, ip, ua string) (*domain.TokenPair, error) {
	if ns == nil {
		return nil, nil
	}

	policy := ns.Config.MFAPolicy
	if policy == domain.MFAPolicyDisabled || policy == "" {
		return nil, nil
	}

	methods, _ := s.mfaService.ListMethods(ctx, user.ID.Hex())
	var enrolledNames []string
	for _, m := range methods {
		if m.Enrolled {
			enrolledNames = append(enrolledNames, m.Method)
		}
	}

	if policy == domain.MFAPolicyOptional && len(enrolledNames) == 0 {
		return nil, nil
	}

	mfaToken, err := s.mfaService.CreateMFAToken(ctx, user.ID.Hex(), ns.Name, domain.MFAMethodTOTP)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create MFA token")
	}
	s.auditService.Log(ctx, domain.EventMFAChallenge, user.ID.Hex(), ns.Name, ip, ua, nil)

	return &domain.TokenPair{
		MFARequired: true,
		MFAToken:    mfaToken,
		MFAMethods:  enrolledNames,
	}, nil
}

func (s *authService) InitiateMFA(ctx context.Context, mfaTokenStr, method string) (string, string, string, error) {
	token, err := s.mfaService.VerifyMFAToken(ctx, mfaTokenStr)
	if err != nil {
		return "", "", "", status.Error(codes.Unauthenticated, "invalid MFA token")
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return "", "", "", status.Error(codes.NotFound, "user not found")
	}

	return s.initiateMFAForUser(ctx, user, token.Namespace, method)
}

func (s *authService) InitiateMFAForUser(ctx context.Context, userID, method string) (string, string, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", "", status.Error(codes.NotFound, "user not found")
	}

	return s.initiateMFAForUser(ctx, user, user.Namespace, method)
}

func (s *authService) initiateMFAForUser(ctx context.Context, user *domain.User, namespace, method string) (string, string, string, error) {
	switch domain.MFAMethod(method) {
	case domain.MFAMethodEmail:
		masked, err := s.mfaService.InitiateEmailOTP(ctx, user.ID.Hex(), namespace, user.Email)
		if err != nil {
			return "", "", "", notifyErrToStatus("failed to initiate email OTP", err)
		}
		return "", "", masked, nil
	case domain.MFAMethodSMS:
		masked, err := s.mfaService.InitiateSMSOTP(ctx, user.ID.Hex(), namespace, "")
		if err != nil {
			return "", "", "", notifyErrToStatus("failed to initiate SMS OTP", err)
		}
		return "", "", masked, nil
	default:
		secret, qr, err := s.mfaService.InitiateTOTP(ctx, user.ID.Hex(), s.issuer, user.Email)
		if err != nil {
			return "", "", "", err
		}
		return secret, qr, "", nil
	}
}

func (s *authService) VerifyMFAForUser(ctx context.Context, userID, code string) error {
	valid, err := s.mfaService.VerifyTOTP(ctx, userID, code)
	if err != nil {
		return status.Error(codes.Internal, "verification failed")
	}
	if !valid {
		return status.Error(codes.InvalidArgument, "invalid code")
	}
	return nil
}

func (s *authService) ListMFAMethods(ctx context.Context, userID string) ([]MFAMethodStatus, error) {
	return s.mfaService.ListMethods(ctx, userID)
}

func (s *authService) EnableMFAMethod(ctx context.Context, userID, method string) error {
	return s.mfaService.EnableMethod(ctx, userID, method)
}

func (s *authService) RemoveMFAMethod(ctx context.Context, userID string, method domain.MFAMethod) error {
	return s.mfaService.RemoveMethod(ctx, userID, method)
}

// notifyErrToStatus maps notification package errors to gRPC status codes.
func notifyErrToStatus(label string, err error) error {
	switch {
	case errors.Is(err, notification.ErrInvalidRecipient):
		return status.Errorf(codes.InvalidArgument, "%s: invalid recipient: %v", label, err)
	case errors.Is(err, notification.ErrProviderUnavailable):
		return status.Errorf(codes.Unavailable, "%s: notification provider unavailable: %v", label, err)
	case errors.Is(err, notification.ErrRateLimited):
		return status.Errorf(codes.ResourceExhausted, "%s: rate limited: %v", label, err)
	default:
		return status.Errorf(codes.Internal, "%s: %v", label, err)
	}
}

func (s *authService) VerifyMFA(ctx context.Context, mfaTokenStr, code string) (*domain.TokenPair, error) {
	token, err := s.mfaService.VerifyMFAToken(ctx, mfaTokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid MFA token")
	}

	var valid bool
	switch token.Method {
	case domain.MFAMethodEmail:
		valid, err = s.mfaService.VerifyEmailOTP(ctx, token.UserID, code)
	case domain.MFAMethodSMS:
		valid, err = s.mfaService.VerifySMSOTP(ctx, token.UserID, code)
	default:
		valid, err = s.mfaService.VerifyTOTP(ctx, token.UserID, code)
	}
	if err != nil || !valid {
		s.auditService.Log(ctx, domain.EventMFAFailed, token.UserID, token.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), nil)
		return nil, status.Error(codes.Unauthenticated, "invalid MFA code")
	}

	// MFA successful, issue final tokens
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	s.auditService.Log(ctx, domain.EventMFAVerified, user.ID.Hex(), user.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), nil)
	s.auditService.Log(ctx, domain.EventLoginSuccess, user.ID.Hex(), user.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), map[string]string{"mfa": "verified"})

	return s.tokenService.GenerateTokenPair(ctx, user, "", nil)
}
