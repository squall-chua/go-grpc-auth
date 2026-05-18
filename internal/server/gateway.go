package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"github.com/squall-chua/go-grpc-auth/api/swagger"
	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewGatewayServer(ctx context.Context, grpcPort string) (*http.Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts)
	if err != nil {
		return nil, err
	}

	err = admin.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts)
	if err != nil {
		return nil, err
	}

	err = admin.RegisterNamespaceServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts)
	if err != nil {
		return nil, err
	}

	err = auth.RegisterOIDCServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts)
	if err != nil {
		return nil, err
	}

	err = admin.RegisterOIDCClientServiceHandlerFromEndpoint(ctx, mux, ":"+grpcPort, opts)
	if err != nil {
		return nil, err
	}

	// Wrapper for other HTTP routes (Swagger, Metrics, etc.)
	handler := http.NewServeMux()

	// Register gRPC Gateway handlers with their prefixes
	handler.Handle("/v1/", mux)
	handler.Handle("/oauth2/", mux)
	handler.Handle("/.well-known/", mux)

	// Serve Swagger JSONs
	swagger.RegisterRoutes(handler)

	// Serve Frontend SPA (Root fallback)
	if err := ServeUI(handler); err != nil {
		return nil, err
	}

	return &http.Server{
		Handler: corsMiddleware(handler),
	}, nil
}
