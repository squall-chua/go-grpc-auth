package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/api/v1/options"
	authservice "github.com/squall-chua/go-grpc-auth/internal/service/auth"
	"github.com/squall-chua/go-grpc-auth/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func AuthUnaryInterceptor(authService authservice.AuthService) grpc.UnaryServerInterceptor {
	methodRules := make(map[string]*options.AuthRule)

	// Dynamically discover rules from proto descriptors
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			sd := services.Get(i)
			methods := sd.Methods()
			for j := 0; j < methods.Len(); j++ {
				md := methods.Get(j)
				opts := md.Options().(*descriptorpb.MethodOptions)
				if proto.HasExtension(opts, options.E_Rule) {
					rule := proto.GetExtension(opts, options.E_Rule).(*options.AuthRule)
					if rule != nil {
						fullMethod := fmt.Sprintf("/%s/%s", sd.FullName(), md.Name())
						methodRules[fullMethod] = rule
					}
				}
			}
		}
		return true
	})

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		rule := methodRules[info.FullMethod]

		// 1. Skip auth if public
		if rule != nil && rule.Public {
			return handler(ctx, req)
		}

		// 2. Authentication
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		authStr := strings.Split(authHeader[0], " ")
		var token string
		if len(authStr) == 2 && strings.EqualFold(authStr[0], "Bearer") {
			token = authStr[1]
		} else {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		p, err := authService.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		principal := &auth.Principal{
			UserId:      p.UserID,
			Namespace:   p.Namespace,
			Roles:       p.Roles,
			Permissions: p.Permissions,
			ExpiresAt:   p.ExpiresAt,
		}

		// 3. Authorization
		if rule != nil {
			// Check Roles
			if len(rule.Roles) > 0 {
				allowed := false
				for _, r := range rule.Roles {
					for _, pr := range principal.Roles {
						if r == pr {
							allowed = true
							break
						}
					}
					if allowed {
						break
					}
				}
				if !allowed {
					return nil, status.Error(codes.PermissionDenied, "insufficient roles")
				}
			}

			// Check Permissions
			if len(rule.Permissions) > 0 {
				allowed := false
				for _, perm := range rule.Permissions {
					for _, pp := range principal.Permissions {
						if perm == pp {
							allowed = true
							break
						}
					}
					if allowed {
						break
					}
				}
				if !allowed {
					return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
				}
			}
		}

		ctx = util.WithPrincipal(ctx, principal)

		// 4. Extract Origin Principal
		if origins := md.Get("origin-principal"); len(origins) > 0 {
			var op auth.Principal
			if err := json.Unmarshal([]byte(origins[0]), &op); err == nil {
				ctx = util.WithOriginPrincipal(ctx, &op)
			}
		}

		return handler(ctx, req)
	}
}
