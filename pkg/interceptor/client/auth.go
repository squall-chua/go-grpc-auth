package client

import (
	"context"
	"encoding/json"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

const (
	PrincipalKey    = "principal"
	OriginHeaderKey = "origin-principal"
)

func UnaryAuthInterceptor(provider TokenProvider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token, err := provider.GetToken(ctx)
		if err != nil {
			return err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		// Propagate origin principal if present
		if p, ok := ctx.Value(PrincipalKey).(*auth.Principal); ok {
			if data, err := json.Marshal(p); err == nil {
				ctx = metadata.AppendToOutgoingContext(ctx, OriginHeaderKey, string(data))
			}
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamAuthInterceptor(provider TokenProvider) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		token, err := provider.GetToken(ctx)
		if err != nil {
			return nil, err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		// Propagate origin principal if present
		if p, ok := ctx.Value(PrincipalKey).(*auth.Principal); ok {
			if data, err := json.Marshal(p); err == nil {
				ctx = metadata.AppendToOutgoingContext(ctx, OriginHeaderKey, string(data))
			}
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
