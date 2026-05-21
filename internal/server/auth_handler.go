package server

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		MfaMethods:   pair.MFAMethods,
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
		MfaMethods:   pair.MFAMethods,
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
	var secret, qrCodeURL, maskedRecipient string
	var err error

	if req.MfaToken != "" {
		secret, qrCodeURL, maskedRecipient, err = s.service.InitiateMFA(ctx, req.MfaToken, req.Method)
	} else {
		p := util.GetPrincipal(ctx)
		if p == nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		secret, qrCodeURL, maskedRecipient, err = s.service.InitiateMFAForUser(ctx, p.UserId, req.Method)
	}
	if err != nil {
		return nil, err
	}
	return &auth.InitiateMFAResponse{
		Secret:          secret,
		QrCodeUrl:       qrCodeURL,
		MaskedRecipient: maskedRecipient,
	}, nil
}

func (s *authGRPCServer) VerifyMFA(ctx context.Context, req *auth.VerifyMFARequest) (*auth.TokenPair, error) {
	if req.MfaToken == "" {
		p := util.GetPrincipal(ctx)
		if p == nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		if err := s.service.VerifyMFAForUser(ctx, p.UserId, req.Code); err != nil {
			return nil, err
		}
		return &auth.TokenPair{}, nil
	}

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

func (s *authGRPCServer) ListMFAMethods(ctx context.Context, req *auth.ListMFAMethodsRequest) (*auth.ListMFAMethodsResponse, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	methods, err := s.service.ListMFAMethods(ctx, p.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list MFA methods: %v", err)
	}
	var statuses []*auth.MFAMethodStatus
	for _, m := range methods {
		statuses = append(statuses, &auth.MFAMethodStatus{
			Method:    m.Method,
			Enrolled:  m.Enrolled,
			Available: m.Available,
		})
	}
	return &auth.ListMFAMethodsResponse{Methods: statuses}, nil
}

func (s *authGRPCServer) RemoveMFAMethod(ctx context.Context, req *auth.RemoveMFAMethodRequest) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	if err := s.service.RemoveMFAMethod(ctx, p.UserId, domain.MFAMethod(req.Method)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove MFA method: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) EnableMFAMethod(ctx context.Context, req *auth.EnableMFAMethodRequest) (*emptypb.Empty, error) {
	p := util.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	if err := s.service.EnableMFAMethod(ctx, p.UserId, req.Method); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to enable MFA method: %v", err)
	}
	return &emptypb.Empty{}, nil
}
