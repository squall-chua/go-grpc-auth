package server

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type authGRPCServer struct {
	auth.UnimplementedAuthServiceServer
	service       authservice.AuthService
	socialService authservice.SocialAuthService
}

func (s *authGRPCServer) GetSocialAuthURL(ctx context.Context, req *auth.SocialAuthURLRequest) (*auth.SocialAuthURLResponse, error) {
	url, err := s.socialService.GetAuthURL(domain.SocialProvider(req.Provider), req.State, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &auth.SocialAuthURLResponse{Url: url}, nil
}

func (s *authGRPCServer) HandleSocialCallback(ctx context.Context, req *auth.SocialCallbackRequest) (*auth.TokenPair, error) {
	pair, err := s.socialService.HandleCallback(ctx, domain.SocialProvider(req.Provider), req.Code, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
		TokenType:    "Bearer",
	}, nil
}

func (s *authGRPCServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.TokenPair, error) {
	pair, err := s.service.Register(ctx, req.Email, req.Username, req.Password, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
		MfaRequired:  pair.MFARequired,
		MfaToken:     pair.MFAToken,
	}, nil
}

func (s *authGRPCServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.TokenPair, error) {
	pair, err := s.service.Login(ctx, req.Login, req.Password, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
		MfaRequired:  pair.MFARequired,
		MfaToken:     pair.MFAToken,
	}, nil
}

func (s *authGRPCServer) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.TokenPair, error) {
	pair, err := s.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
	}, nil
}

func (s *authGRPCServer) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)
	err := s.service.Logout(ctx, p.UserId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) ChangePassword(ctx context.Context, req *auth.ChangePasswordRequest) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)

	err := s.service.ChangePassword(ctx, p.UserId, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.Principal, error) {
	p, err := s.service.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &auth.Principal{
		UserId:      p.UserID,
		Namespace:   p.Namespace,
		Roles:       p.Roles,
		Permissions: p.Permissions,
		ExpiresAt:   p.ExpiresAt,
	}, nil
}

func (s *authGRPCServer) InitiateMFA(ctx context.Context, req *auth.InitiateMFARequest) (*auth.InitiateMFAResponse, error) {
	secret, qrURL, err := s.service.InitiateMFA(ctx, req.MfaToken, req.Method)
	if err != nil {
		return nil, err
	}
	return &auth.InitiateMFAResponse{
		Secret:    secret,
		QrCodeUrl: qrURL,
	}, nil
}

func (s *authGRPCServer) VerifyMFA(ctx context.Context, req *auth.VerifyMFARequest) (*auth.TokenPair, error) {
	pair, err := s.service.VerifyMFA(ctx, req.MfaToken, req.Code)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IdToken:      pair.IDToken,
		ExpiresIn:    int32(pair.ExpiresIn),
	}, nil
}
