package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
)

type MFAService interface {
	InitiateTOTP(ctx context.Context, userID string, issuer string, accountName string) (secret string, qrCodeURL string, err error)
	VerifyTOTP(ctx context.Context, userID string, code string) (bool, error)

	InitiateEmailOTP(ctx context.Context, userID string, email string) error
	VerifyEmailOTP(ctx context.Context, userID string, code string) (bool, error)

	InitiateSMSOTP(ctx context.Context, userID string, phoneNumber string) error
	VerifySMSOTP(ctx context.Context, userID string, code string) (bool, error)
	
	CreateMFAToken(ctx context.Context, userID string, namespace string, method domain.MFAMethod) (string, error)
	VerifyMFAToken(ctx context.Context, tokenStr string) (*domain.MFAToken, error)
}

type MFAConfig struct {
	EmailEnabled bool
	SMSEnabled   bool
}

type mfaService struct {
	repo   repository.MFARepository
	config MFAConfig
}

func NewMFAService(repo repository.MFARepository, config MFAConfig) MFAService {
	return &mfaService{repo: repo, config: config}
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

func (s *mfaService) InitiateEmailOTP(ctx context.Context, userID string, email string) error {
	if !s.config.EmailEnabled {
		return fmt.Errorf("email OTP delivery is not enabled")
	}

	code, err := generateOTPCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
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
		return fmt.Errorf("failed to store email OTP: %w", err)
	}

	// TODO: integrate with an email delivery provider (e.g. SES, SendGrid)
	return fmt.Errorf("email OTP delivery not implemented")
}

func (s *mfaService) VerifyEmailOTP(ctx context.Context, userID string, code string) (bool, error) {
	return s.verifyOTP(ctx, userID, domain.MFAMethodEmail, code)
}

func (s *mfaService) InitiateSMSOTP(ctx context.Context, userID string, phoneNumber string) error {
	if !s.config.SMSEnabled {
		return fmt.Errorf("SMS OTP delivery is not enabled")
	}

	code, err := generateOTPCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
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
		return fmt.Errorf("failed to store SMS OTP: %w", err)
	}

	// TODO: integrate with an SMS gateway (e.g. Twilio, AWS SNS)
	return fmt.Errorf("SMS OTP delivery not implemented")
}

func (s *mfaService) VerifySMSOTP(ctx context.Context, userID string, code string) (bool, error) {
	return s.verifyOTP(ctx, userID, domain.MFAMethodSMS, code)
}

func (s *mfaService) verifyOTP(ctx context.Context, userID string, method domain.MFAMethod, code string) (bool, error) {
	secret, err := s.repo.GetSecret(ctx, userID, method)
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired: %w", err)
	}

	// OTP codes expire after 5 minutes
	if time.Since(secret.CreatedAt) > 5*time.Minute {
		s.repo.DeleteSecret(ctx, userID, method)
		return false, nil
	}

	if secret.Secret != code {
		return false, nil
	}

	// Clean up used OTP
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
