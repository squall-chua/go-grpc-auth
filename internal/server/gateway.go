package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/squall-chua/go-grpc-auth/api/swagger"
	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
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

func NewGatewayServer(ctx context.Context, conn *grpc.ClientConn, uiCfg UIConfig) (*http.Server, error) {
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

	registerHandlers := []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
		auth.RegisterAuthServiceHandler,
		admin.RegisterAdminServiceHandler,
		admin.RegisterNamespaceServiceHandler,
		auth.RegisterOIDCServiceHandler,
		admin.RegisterOIDCClientServiceHandler,
	}
	for _, register := range registerHandlers {
		if err := register(ctx, mux, conn); err != nil {
			return nil, err
		}
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
	if err := ServeUI(handler, uiCfg); err != nil {
		return nil, err
	}

	return &http.Server{
		Handler: corsMiddleware(handler),
	}, nil
}
