package auth

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"github.com/squall-chua/go-grpc-auth/internal/service/audit"
	tokenservice "github.com/squall-chua/go-grpc-auth/internal/service/token"
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
	ListWallets(ctx context.Context, userID string) ([]*domain.WalletInfo, error)
	LinkWallet(ctx context.Context, req LinkRequest) error
	UnlinkWallet(ctx context.Context, userID, namespace, wallet string) error
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
func (s *web3AuthService) RequestNonce(_ context.Context, _ string, wallet string) (string, error) {
	_ = common.HexToAddress(wallet) // compile-time check that the eth dep is wired
	return "", nil
}
func (s *web3AuthService) Verify(_ context.Context, _ VerifyRequest) (*domain.TokenResponse, *domain.WalletInfo, error) {
	return nil, nil, nil
}
func (s *web3AuthService) ListWallets(_ context.Context, _ string) ([]*domain.WalletInfo, error) {
	return nil, nil
}
func (s *web3AuthService) LinkWallet(_ context.Context, _ LinkRequest) error {
	return nil
}
func (s *web3AuthService) UnlinkWallet(_ context.Context, _, _, _ string) error {
	return nil
}
