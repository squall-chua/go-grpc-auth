package auth

import (
	"context"
	"fmt"
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

type mfaService struct {
	repo repository.MFARepository
}

func NewMFAService(repo repository.MFARepository) MFAService {
	return &mfaService{repo: repo}
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
		secret.UpdatedAt = time.Now()
		if err := s.repo.UpsertSecret(ctx, secret); err != nil {
			return false, err
		}
	}

	return valid, nil
}

func (s *mfaService) InitiateEmailOTP(ctx context.Context, userID string, email string) error {
	// Mock: generate 6-digit code and send via email
	fmt.Printf("[MFA] Sending Email OTP to %s\n", email)
	return nil
}

func (s *mfaService) VerifyEmailOTP(ctx context.Context, userID string, code string) (bool, error) {
	// Mock: verify code from repo
	return code == "123456", nil
}

func (s *mfaService) InitiateSMSOTP(ctx context.Context, userID string, phoneNumber string) error {
	// Mock: generate 6-digit code and send via SMS
	fmt.Printf("[MFA] Sending SMS OTP to %s\n", phoneNumber)
	return nil
}

func (s *mfaService) VerifySMSOTP(ctx context.Context, userID string, code string) (bool, error) {
	// Mock: verify code from repo
	return code == "123456", nil
}

func (s *mfaService) CreateMFAToken(ctx context.Context, userID string, namespace string, method domain.MFAMethod) (string, error) {
	tokenStr := uuid.New().String()
	token := &domain.MFAToken{
		Token:     tokenStr,
		UserID:    userID,
		Namespace: namespace,
		Method:    method,
		ExpiresAt: time.Now().Add(5 * time.Minute),
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

	if time.Now().After(token.ExpiresAt) {
		s.repo.DeleteToken(ctx, tokenStr)
		return nil, fmt.Errorf("MFA token expired")
	}

	return token, nil
}
