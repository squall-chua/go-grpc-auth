package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

type MFAMethodStatus struct {
	Method    string
	Enrolled  bool
	Available bool
}

type MFAService interface {
	InitiateTOTP(ctx context.Context, userID string, issuer string, accountName string) (secret string, qrCodeURL string, err error)
	VerifyTOTP(ctx context.Context, userID string, code string) (bool, error)

	InitiateEmailOTP(ctx context.Context, userID, namespace, email string) (string, error)
	VerifyEmailOTP(ctx context.Context, userID string, code string) (bool, error)

	InitiateSMSOTP(ctx context.Context, userID, namespace, phoneNumber string) (string, error)
	VerifySMSOTP(ctx context.Context, userID string, code string) (bool, error)

	CreateMFAToken(ctx context.Context, userID string, namespace string, method domain.MFAMethod) (string, error)
	VerifyMFAToken(ctx context.Context, tokenStr string) (*domain.MFAToken, error)

	ListMethods(ctx context.Context, userID string) ([]MFAMethodStatus, error)
	EnableMethod(ctx context.Context, userID, method string) error
	RemoveMethod(ctx context.Context, userID string, method domain.MFAMethod) error
}

type mfaService struct {
	repo     repository.MFARepository
	notifier notification.Service
	appName  string
}

func NewMFAService(repo repository.MFARepository, notifier notification.Service, appName string) MFAService {
	return &mfaService{repo: repo, notifier: notifier, appName: appName}
}

func (s *mfaService) InitiateTOTP(ctx context.Context, userID string, issuer string, accountName string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}

	secret := &domain.MFASecret{
		UserID:    userID,
		Method:    domain.MFAMethodTOTP,
		Secret:    key.Secret(),
		Confirmed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.UpsertSecret(ctx, secret); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *mfaService) VerifyTOTP(ctx context.Context, userID string, code string) (bool, error) {
	secret, err := s.repo.GetSecret(ctx, userID, domain.MFAMethodTOTP)
	if err != nil {
		return false, err
	}

	valid := totp.Validate(code, secret.Secret)
	if valid && !secret.Confirmed {
		secret.Confirmed = true
		secret.UpdatedAt = time.Now().UTC()
		if err := s.repo.UpsertSecret(ctx, secret); err != nil {
			return false, err
		}
	}

	return valid, nil
}

func (s *mfaService) InitiateEmailOTP(ctx context.Context, userID, namespace, email string) (string, error) {
	code, err := generateOTPCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	secret := &domain.MFASecret{
		UserID:    userID,
		Method:    domain.MFAMethodEmail,
		Secret:    code,
		Confirmed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.UpsertSecret(ctx, secret); err != nil {
		return "", fmt.Errorf("failed to store email OTP: %w", err)
	}

	if err := s.notifier.SendEmail(ctx, namespace, userID, "mfa_email_otp", email, map[string]any{
		"Code":       code,
		"TTLMinutes": 5,
		"AppName":    s.appName,
	}); err != nil {
		// Roll back the stored OTP so a failed delivery does not leave a guessable secret.
		s.repo.DeleteSecret(ctx, userID, domain.MFAMethodEmail)
		return "", err
	}
	return maskEmail(email), nil
}

func (s *mfaService) VerifyEmailOTP(ctx context.Context, userID string, code string) (bool, error) {
	return s.verifyOTP(ctx, userID, domain.MFAMethodEmail, code)
}

func (s *mfaService) InitiateSMSOTP(ctx context.Context, userID, namespace, phoneNumber string) (string, error) {
	code, err := generateOTPCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	secret := &domain.MFASecret{
		UserID:    userID,
		Method:    domain.MFAMethodSMS,
		Secret:    code,
		Confirmed: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.UpsertSecret(ctx, secret); err != nil {
		return "", fmt.Errorf("failed to store SMS OTP: %w", err)
	}

	if err := s.notifier.SendSMS(ctx, namespace, userID, "mfa_sms_otp", phoneNumber, map[string]any{
		"Code":       code,
		"TTLMinutes": 5,
	}); err != nil {
		s.repo.DeleteSecret(ctx, userID, domain.MFAMethodSMS)
		return "", err
	}
	return maskPhone(phoneNumber), nil
}

func (s *mfaService) VerifySMSOTP(ctx context.Context, userID string, code string) (bool, error) {
	return s.verifyOTP(ctx, userID, domain.MFAMethodSMS, code)
}

func (s *mfaService) verifyOTP(ctx context.Context, userID string, method domain.MFAMethod, code string) (bool, error) {
	secret, err := s.repo.GetSecret(ctx, userID, method)
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired: %w", err)
	}

	if time.Since(secret.CreatedAt) > 5*time.Minute {
		s.repo.DeleteSecret(ctx, userID, method)
		return false, nil
	}

	if secret.Secret != code {
		return false, nil
	}

	s.repo.DeleteSecret(ctx, userID, method)
	return true, nil
}

func generateOTPCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	name := parts[0]
	if len(name) <= 2 {
		return name[:1] + "***@" + parts[1]
	}
	return name[:1] + strings.Repeat("*", len(name)-2) + name[len(name)-1:] + "@" + parts[1]
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[:len(phone)-4] + " ****"
}

func (s *mfaService) ListMethods(ctx context.Context, userID string) ([]MFAMethodStatus, error) {
	methods, err := s.repo.ListEnrolledMethods(ctx, userID)
	if err != nil {
		return nil, err
	}
	enrolled := make(map[domain.MFAMethod]bool)
	for _, m := range methods {
		enrolled[m] = true
	}
	return []MFAMethodStatus{
		{Method: "totp", Enrolled: enrolled[domain.MFAMethodTOTP], Available: true},
		{Method: "email", Enrolled: enrolled[domain.MFAMethodEmail], Available: true},
		{Method: "sms", Enrolled: enrolled[domain.MFAMethodSMS], Available: false},
	}, nil
}

func (s *mfaService) EnableMethod(ctx context.Context, userID, method string) error {
	switch domain.MFAMethod(method) {
	case domain.MFAMethodEmail:
		secret := &domain.MFASecret{
			UserID:    userID,
			Method:    domain.MFAMethodEmail,
			Secret:    "",
			Confirmed: true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		return s.repo.UpsertSecret(ctx, secret)
	case domain.MFAMethodSMS:
		return fmt.Errorf("SMS MFA requires a phone number on the user profile")
	case domain.MFAMethodTOTP:
		return fmt.Errorf("use InitiateMFA + VerifyMFA to set up TOTP")
	default:
		return fmt.Errorf("unknown MFA method: %s", method)
	}
}

func (s *mfaService) RemoveMethod(ctx context.Context, userID string, method domain.MFAMethod) error {
	return s.repo.DeleteSecret(ctx, userID, method)
}

func (s *mfaService) CreateMFAToken(ctx context.Context, userID string, namespace string, method domain.MFAMethod) (string, error) {
	tokenStr := uuid.New().String()
	token := &domain.MFAToken{
		Token:     tokenStr,
		UserID:    userID,
		Namespace: namespace,
		Method:    method,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}

	if err := s.repo.CreateToken(ctx, token); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *mfaService) VerifyMFAToken(ctx context.Context, tokenStr string) (*domain.MFAToken, error) {
	token, err := s.repo.GetToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		s.repo.DeleteToken(ctx, tokenStr)
		return nil, fmt.Errorf("MFA token expired")
	}

	return token, nil
}
