package server

import (
	"context"
	"strings"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type web3GRPCServer struct {
	auth.UnimplementedWeb3AuthServiceServer
	service authservice.Web3AuthService
	issuer  string
	users   repository.UserRepository
}

func (s *web3GRPCServer) RequestNonce(ctx context.Context, req *auth.RequestNonceRequest) (*auth.RequestNonceResponse, error) {
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if !strings.HasPrefix(req.Wallet, "0x") {
		return nil, status.Error(codes.InvalidArgument, "wallet must be a 0x-prefixed address")
	}
	nonce, err := s.service.RequestNonce(ctx, req.Namespace, req.Wallet)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nonce: %v", err)
	}
	return &auth.RequestNonceResponse{
		Nonce:  nonce,
		Domain: s.issuer,
		Uri:    s.issuer + "/v1/auth/web3/verify",
	}, nil
}

func (s *web3GRPCServer) Verify(ctx context.Context, req *auth.VerifyRequest) (*auth.TokenPair, error) {
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	pair, _, err := s.service.Verify(ctx, authservice.VerifyRequest{
		Namespace:   req.Namespace,
		SIWEMessage: req.Message,
		Signature:   req.Signature,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "verify: %v", err)
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
		TokenType:    "Bearer",
	}, nil
}

func (s *web3GRPCServer) ListWallets(ctx context.Context, _ *emptypb.Empty) (*auth.ListWalletsResponse, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	user, err := loadUserForPrincipal(ctx, p, s.users)
	if err != nil {
		return nil, err
	}
	wallets, err := s.service.ListWallets(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list wallets: %v", err)
	}
	out := make([]*auth.Wallet, 0, len(wallets))
	for _, w := range wallets {
		out = append(out, &auth.Wallet{Address: w.Address, ChainId: w.ChainId})
	}
	return &auth.ListWalletsResponse{Wallets: out}, nil
}

func (s *web3GRPCServer) LinkWallet(ctx context.Context, req *auth.LinkWalletRequest) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	if err := s.service.LinkWallet(ctx, authservice.LinkRequest{
		UserID:      p.UserId,
		Namespace:   p.Namespace,
		SIWEMessage: req.Message,
		Signature:   req.Signature,
	}); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "link: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *web3GRPCServer) UnlinkWallet(ctx context.Context, req *auth.UnlinkWalletRequest) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	user, err := loadUserForPrincipal(ctx, p, s.users)
	if err != nil {
		return nil, err
	}
	if err := s.service.UnlinkWallet(ctx, user, req.Wallet); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unlink: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// loadUserForPrincipal fetches the user record backing the principal.
func loadUserForPrincipal(ctx context.Context, p *auth.Principal, repo repository.UserRepository) (*domain.User, error) {
	if repo == nil {
		return nil, status.Error(codes.Internal, "user repository not wired")
	}
	return repo.GetByID(ctx, p.UserId)
}
