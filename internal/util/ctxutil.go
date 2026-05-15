package util

import (
	"context"
	"strings"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	ContextKeyPrincipal            = "principal"
	ContextKeyOriginPrincipal      = "origin-principal"
	HTTPHeaderXForwardedFor        = "x-forwarded-for"
	HTTPHeaderUserAgent            = "user-agent"
	HTTPHeaderGrpcGatewayUserAgent = "grpcgateway-user-agent"
)

func GetClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if xff := md.Get(HTTPHeaderXForwardedFor); len(xff) > 0 {
			return strings.Split(xff[0], ",")[0]
		}
	}

	p, ok := peer.FromContext(ctx)
	if ok {
		return p.Addr.String()
	}

	return ""
}

func GetUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ua := md.Get(HTTPHeaderGrpcGatewayUserAgent); len(ua) > 0 {
			return ua[0]
		}
		if ua := md.Get(HTTPHeaderUserAgent); len(ua) > 0 {
			return ua[0]
		}
	}
	return ""
}

func WithPrincipal(ctx context.Context, principal *auth.Principal) context.Context {
	return context.WithValue(ctx, ContextKeyPrincipal, principal)
}

func GetPrincipal(ctx context.Context) *auth.Principal {
	p, _ := ctx.Value(ContextKeyPrincipal).(*auth.Principal)
	return p
}

func WithOriginPrincipal(ctx context.Context, principal *auth.Principal) context.Context {
	return context.WithValue(ctx, ContextKeyOriginPrincipal, principal)
}

func GetOriginPrincipal(ctx context.Context) *auth.Principal {
	p, _ := ctx.Value(ContextKeyOriginPrincipal).(*auth.Principal)
	return p
}
