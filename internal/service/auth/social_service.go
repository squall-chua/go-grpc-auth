package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SocialAuthService interface {
	GetAuthURL(provider domain.SocialProvider, state string, namespace string) (string, error)
	HandleCallback(ctx context.Context, provider domain.SocialProvider, code string, namespace string) (*domain.TokenResponse, error)
}

type socialAuthService struct {
	providers map[domain.SocialProvider]domain.SocialProviderInterface
	userRepo  repository.UserRepository
	tokenSvc  tokenservice.TokenService
}

func NewSocialAuthService(
	userRepo repository.UserRepository,
	tokenSvc tokenservice.TokenService,
	providers []domain.SocialProviderInterface,
) SocialAuthService {
	pMap := make(map[domain.SocialProvider]domain.SocialProviderInterface)
	for _, p := range providers {
		pMap[p.GetProvider()] = p
	}
	return &socialAuthService{
		providers: pMap,
		userRepo:  userRepo,
		tokenSvc:  tokenSvc,
	}
}

func (s *socialAuthService) GetAuthURL(provider domain.SocialProvider, state string, namespace string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", fmt.Errorf("provider not supported: %s", provider)
	}
	// Note: namespace could be encoded into state or used for namespace-specific config
	return p.GetAuthURL(state), nil
}

func (s *socialAuthService) HandleCallback(ctx context.Context, provider domain.SocialProvider, code string, namespace string) (*domain.TokenResponse, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not supported")
	}

	socialUser, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Find or create user
	user, err := s.userRepo.GetByEmail(ctx, socialUser.Email, namespace)
	if err != nil {
		// Auto-provision user
		user = &domain.User{
			ID:        bson.NewObjectID(),
			Email:     socialUser.Email,
			Username:  socialUser.Email, // Default to email
			Namespace: namespace,
			Status:    "active",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			SocialIdentities: []domain.SocialIdentity{
				{
					Provider:   provider,
					ExternalID: socialUser.ID,
					Email:      socialUser.Email,
					Name:       socialUser.Name,
					AvatarURL:  socialUser.AvatarURL,
				},
			},
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create social user: %w", err)
		}
	} else {
		// Link identity if not already linked
		linked := false
		for _, id := range user.SocialIdentities {
			if id.Provider == provider && id.ExternalID == socialUser.ID {
				linked = true
				break
			}
		}
		if !linked {
			user.SocialIdentities = append(user.SocialIdentities, domain.SocialIdentity{
				Provider:   provider,
				ExternalID: socialUser.ID,
				Email:      socialUser.Email,
				Name:       socialUser.Name,
				AvatarURL:  socialUser.AvatarURL,
			})
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to link social identity: %w", err)
			}
		}
	}

	// Issue tokens
	pair, err := s.tokenSvc.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &domain.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IDToken:      pair.IDToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}
