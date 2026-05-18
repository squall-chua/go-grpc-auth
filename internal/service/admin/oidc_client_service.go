package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type OIDCClientService interface {
	RegisterClient(ctx context.Context, client *domain.OIDCClient) (string, error) // returns plain secret
	GetClient(ctx context.Context, clientID string) (*domain.OIDCClient, error)
	UpdateClient(ctx context.Context, clientID string, fields repository.ClientUpdateFields) (*domain.OIDCClient, error)
	DeleteClient(ctx context.Context, clientID string) error
	ListClients(ctx context.Context, namespace, query string, page, pageSize int) ([]*domain.OIDCClient, int64, error)
	RotateSecret(ctx context.Context, clientID string) (string, error)
}

type oidcClientService struct {
	repo repository.ClientRepository
}

func NewOIDCClientService(repo repository.ClientRepository) OIDCClientService {
	return &oidcClientService{repo: repo}
}

func (s *oidcClientService) RegisterClient(ctx context.Context, client *domain.OIDCClient) (string, error) {
	client.ClientID = uuid.New().String()
	
	secret, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	client.ClientSecret = string(hashed)
	client.CreatedAt = time.Now().UTC()
	client.UpdatedAt = time.Now().UTC()

	if err := s.repo.Create(ctx, client); err != nil {
		return "", err
	}

	return secret, nil
}

func (s *oidcClientService) GetClient(ctx context.Context, clientID string) (*domain.OIDCClient, error) {
	return s.repo.GetByID(ctx, clientID)
}

func (s *oidcClientService) UpdateClient(ctx context.Context, clientID string, fields repository.ClientUpdateFields) (*domain.OIDCClient, error) {
	return s.repo.Update(ctx, clientID, fields)
}

func (s *oidcClientService) DeleteClient(ctx context.Context, clientID string) error {
	return s.repo.Delete(ctx, clientID)
}

func (s *oidcClientService) ListClients(ctx context.Context, namespace, query string, page, pageSize int) ([]*domain.OIDCClient, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, namespace, query, offset, pageSize)
}

func (s *oidcClientService) RotateSecret(ctx context.Context, clientID string) (string, error) {
	secret, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := s.repo.UpdateSecret(ctx, clientID, string(hashed)); err != nil {
		return "", err
	}

	return secret, nil
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
