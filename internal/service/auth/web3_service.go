package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// VerifyRequest is the input to Web3AuthService.Verify.
type VerifyRequest struct {
	Namespace   string
	SIWEMessage string
	Signature   string
}

// LinkRequest is the input to Web3AuthService.LinkWallet.
type LinkRequest struct {
	UserID      string
	Namespace   string
	SIWEMessage string
	Signature   string
}

// Web3AuthService is the wallet sign-in / link / unlink surface.
type Web3AuthService interface {
	RequestNonce(ctx context.Context, namespace, wallet string) (nonce string, err error)
	Verify(ctx context.Context, req VerifyRequest) (*domain.TokenResponse, *domain.WalletInfo, error)
	ListWallets(ctx context.Context, user *domain.User) ([]*domain.WalletInfo, error)
	LinkWallet(ctx context.Context, req LinkRequest) error
	UnlinkWallet(ctx context.Context, user *domain.User, wallet string) error
}

type web3AuthService struct {
	issuer     string
	nonceStore NonceStore
	nonceTTL   time.Duration
	allowed    AllowedChainResolver
	userRepo   repository.UserRepository
	tokenSvc   tokenservice.TokenService
	auditSvc   audit.AuditService
}

// AllowedChainResolver returns the allowed chainIds for a namespace. If the
// namespace has no explicit list, the global default is returned.
type AllowedChainResolver interface {
	AllowedChainIDs(ctx context.Context, namespace string) ([]int64, error)
}

func NewWeb3AuthService(
	issuer string,
	nonceStore NonceStore,
	nonceTTL time.Duration,
	allowed AllowedChainResolver,
	userRepo repository.UserRepository,
	tokenSvc tokenservice.TokenService,
	auditSvc audit.AuditService,
) Web3AuthService {
	return &web3AuthService{
		issuer:     issuer,
		nonceStore: nonceStore,
		nonceTTL:   nonceTTL,
		allowed:    allowed,
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		auditSvc:   auditSvc,
	}
}

// method stubs (implemented in Tasks 13-16)
func (s *web3AuthService) RequestNonce(ctx context.Context, namespace, wallet string) (string, error) {
	if !common.IsHexAddress(wallet) {
		return "", errors.New("invalid wallet address")
	}
	wallet = strings.ToLower(wallet)
	nonce := util.RandomString(16)
	if err := s.nonceStore.Save(ctx, namespace, wallet, nonce, s.nonceTTL); err != nil {
		return "", err
	}
	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, domain.EventWeb3NonceIssued, "", namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), map[string]any{
			"wallet": wallet,
		})
	}
	return nonce, nil
}
func (s *web3AuthService) Verify(ctx context.Context, req VerifyRequest) (*domain.TokenResponse, *domain.WalletInfo, error) {
	// 1. Parse + verify SIWE (no domain binding in v1; we bind nonce instead).
	msg, err := siwe.ParseMessage(req.SIWEMessage)
	if err != nil {
		s.logFail(ctx, req.Namespace, "parse", err)
		return nil, nil, err
	}
	recovered, err := msg.VerifyEIP191(req.Signature)
	if err != nil {
		s.logFail(ctx, req.Namespace, "signature", err)
		return nil, nil, err
	}
	addrFromMsg := msg.GetAddress()
	if recovered == nil {
		s.logFail(ctx, req.Namespace, "recover", errors.New("nil public key"))
		return nil, nil, errors.New("signature recovery failed")
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if recoveredAddr != addrFromMsg {
		s.logFail(ctx, req.Namespace, "address_mismatch", nil)
		return nil, nil, errors.New("signature does not match address")
	}
	if !IsEOA(recoveredAddr) {
		s.logFail(ctx, req.Namespace, "smart_contract_wallet", nil)
		return nil, nil, errors.New("smart-contract wallets are not supported in v1")
	}
	checksum := recoveredAddr.Hex()
	lower := strings.ToLower(checksum)
	chainID := int64(msg.GetChainID())

	// 2. Verify nonce: must match what we issued for this (ns, wallet).
	ok, err := s.nonceStore.Consume(ctx, req.Namespace, lower, msg.GetNonce())
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		s.logFail(ctx, req.Namespace, "nonce_invalid", nil)
		return nil, nil, errors.New("nonce not found or already used")
	}

	// 3. Verify chainId against the namespace allowlist.
	allowed, err := s.allowed.AllowedChainIDs(ctx, req.Namespace)
	if err != nil {
		return nil, nil, err
	}
	if !containsInt64(allowed, chainID) {
		s.logFail(ctx, req.Namespace, "chain_not_allowed", nil)
		return nil, nil, fmt.Errorf("chain id %d not allowed for this namespace", chainID)
	}

	// 4. Find-or-create user.
	user, err := s.findOrCreateUser(ctx, req.Namespace, checksum)
	if err != nil {
		return nil, nil, err
	}

	// 5. Issue tokens with web3 claims.
	pair, err := s.tokenSvc.GenerateTokenPairWithClaims(ctx, user, "", nil, map[string]any{
		"web3_address":  checksum,
		"web3_chain_id": chainID,
	})
	if err != nil {
		return nil, nil, err
	}

	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, domain.EventWeb3SignInSuccess, user.ID.Hex(), req.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), map[string]any{
			"wallet":   checksum,
			"chain_id": chainID,
		})
	}

	return &domain.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IDToken:      pair.IDToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    "Bearer",
	}, &domain.WalletInfo{Address: checksum, ChainId: chainID}, nil
}

func (s *web3AuthService) findOrCreateUser(ctx context.Context, namespace, checksum string) (*domain.User, error) {
	lower := strings.ToLower(checksum)
	// 1. Try by SocialIdentity.
	user, err := s.userRepo.GetBySocialIdentity(ctx, namespace, domain.ProviderEthereum, lower)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	// 2. Try by synthetic email.
	synthEmail := "0x" + strings.TrimPrefix(lower, "0x") + "@wallet.local"
	user, err = s.userRepo.GetByEmail(ctx, namespace, synthEmail)
	if err == nil {
		// Link identity. Prepend so the index-0 query in
		// userRepo.GetBySocialIdentity (which only checks the first
		// element) finds it on subsequent lookups.
		user.SocialIdentities = append([]domain.SocialIdentity{
			{
				Provider:   domain.ProviderEthereum,
				ExternalID: lower,
				Email:      synthEmail,
			},
		}, user.SocialIdentities...)
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	// 3. Create.
	user = &domain.User{
		ID:        bson.NewObjectID(),
		Email:     synthEmail,
		Username:  checksum,
		Namespace: namespace,
		Status:    domain.UserStatusActive,
		SocialIdentities: []domain.SocialIdentity{
			{
				Provider:   domain.ProviderEthereum,
				ExternalID: lower,
				Email:      synthEmail,
			},
		},
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *web3AuthService) logFail(ctx context.Context, namespace, reason string, cause error) {
	if s.auditSvc == nil {
		return
	}
	meta := map[string]any{"reason": reason}
	if cause != nil {
		meta["error"] = cause.Error()
	}
	s.auditSvc.Log(ctx, domain.EventWeb3SignInFailed, "", namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), meta)
}

func containsInt64(haystack []int64, needle int64) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
func (s *web3AuthService) ListWallets(_ context.Context, user *domain.User) ([]*domain.WalletInfo, error) {
	out := make([]*domain.WalletInfo, 0, len(user.SocialIdentities))
	for _, id := range user.SocialIdentities {
		if id.Provider != domain.ProviderEthereum {
			continue
		}
		out = append(out, &domain.WalletInfo{
			Address: common.HexToAddress(id.ExternalID).Hex(),
		})
	}
	return out, nil
}

func (s *web3AuthService) LinkWallet(ctx context.Context, req LinkRequest) error {
	msg, err := siwe.ParseMessage(req.SIWEMessage)
	if err != nil {
		return err
	}
	if _, err := msg.VerifyEIP191(req.Signature); err != nil {
		return err
	}
	addr := strings.ToLower(msg.GetAddress().Hex())
	chainID := int64(msg.GetChainID())

	// Enforce chain allowlist (same as Verify).
	allowed, err := s.allowed.AllowedChainIDs(ctx, req.Namespace)
	if err != nil {
		return err
	}
	if !containsInt64(allowed, chainID) {
		return fmt.Errorf("chain id %d not allowed for this namespace", chainID)
	}

	// Reject if already linked to any user.
	if existing, err := s.userRepo.GetBySocialIdentity(ctx, req.Namespace, domain.ProviderEthereum, addr); err == nil && existing != nil {
		if s.auditSvc != nil {
			s.auditSvc.Log(ctx, domain.EventWeb3WalletLinkConflict, req.UserID, req.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), map[string]any{
				"wallet":      msg.GetAddress().Hex(),
				"linked_user": existing.ID.Hex(),
			})
		}
		return errors.New("wallet already linked to another account")
	} else if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return err
	}

	// Load current user.
	objID, err := bson.ObjectIDFromHex(req.UserID)
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, objID.Hex())
	if err != nil {
		return err
	}

	// Prepend identity so the index-0 query in userRepo.GetBySocialIdentity
	// (which only checks the first element) finds it on subsequent lookups.
	user.SocialIdentities = append([]domain.SocialIdentity{
		{
			Provider:   domain.ProviderEthereum,
			ExternalID: addr,
			Email:      "",
		},
	}, user.SocialIdentities...)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, domain.EventWeb3WalletLinked, user.ID.Hex(), req.Namespace, util.GetClientIP(ctx), util.GetUserAgent(ctx), map[string]any{
			"wallet":   msg.GetAddress().Hex(),
			"chain_id": chainID,
		})
	}
	return nil
}

func (s *web3AuthService) UnlinkWallet(ctx context.Context, user *domain.User, wallet string) error {
	target := strings.ToLower(wallet)
	filtered := user.SocialIdentities[:0]
	for _, id := range user.SocialIdentities {
		if id.Provider == domain.ProviderEthereum && strings.ToLower(id.ExternalID) == target {
			continue
		}
		filtered = append(filtered, id)
	}
	// Anti-soft-lockout: must have password or at least one other identity.
	if user.PasswordHash == "" {
		otherWallets := 0
		otherIdents := 0
		for _, id := range filtered {
			if id.Provider == domain.ProviderEthereum {
				otherWallets++
			} else {
				otherIdents++
			}
		}
		if otherWallets == 0 && otherIdents == 0 {
			return errors.New("cannot unlink last credential")
		}
	}
	user.SocialIdentities = filtered
	return s.userRepo.Update(ctx, user)
}
