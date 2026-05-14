package ctxutil

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func GetClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if xff := md.Get("x-forwarded-for"); len(xff) > 0 {
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
		if ua := md.Get("grpcgateway-user-agent"); len(ua) > 0 {
			return ua[0]
		}
		if ua := md.Get("user-agent"); len(ua) > 0 {
			return ua[0]
		}
	}
	return ""
}
