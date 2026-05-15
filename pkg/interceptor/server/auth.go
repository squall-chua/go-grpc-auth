package server

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/api/v1/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	PrincipalKey       string = "principal"
	OriginPrincipalKey string = "origin-principal"
)

type cacheEntry struct {
	principal *auth.Principal
	expiresAt time.Time
}

type AuthInterceptor struct {
	authClient auth.AuthServiceClient
	cache      map[string]cacheEntry
	rules      map[string]*options.AuthRule
	mu         sync.RWMutex
}

func NewAuthInterceptor(authClient auth.AuthServiceClient) *AuthInterceptor {
	return &AuthInterceptor{
		authClient: authClient,
		cache:      make(map[string]cacheEntry),
		rules:      make(map[string]*options.AuthRule),
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. Get Rule from Method Options
		rule := i.getRule(info.FullMethod)

		// 2. Auth Check
		principal, err := i.authenticate(ctx, rule)
		if err != nil {
			return nil, err
		}

		// 3. Add Principal to Context
		if principal != nil {
			ctx = context.WithValue(ctx, PrincipalKey, principal)
		}

		// 4. Extract Origin Principal
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if origins := md.Get("origin-principal"); len(origins) > 0 {
				var op auth.Principal
				if err := json.Unmarshal([]byte(origins[0]), &op); err == nil {
					ctx = context.WithValue(ctx, OriginPrincipalKey, &op)
				}
			}
		}

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) authenticate(ctx context.Context, rule *options.AuthRule) (*auth.Principal, error) {
	if rule != nil && rule.Public {
		return nil, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header is missing")
	}

	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := parts[1]
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "token is missing")
	}

	// Check Cache
	i.mu.RLock()
	entry, found := i.cache[token]
	i.mu.RUnlock()

	if found && time.Now().UTC().Before(entry.expiresAt) {
		return i.authorize(entry.principal, rule)
	}

	// Validate with Auth Service
	resp, err := i.authClient.ValidateToken(ctx, &auth.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}

	// Update Cache
	i.mu.Lock()
	i.cache[token] = cacheEntry{
		principal: resp,
		expiresAt: time.Now().UTC().Add(30 * time.Second),
	}
	i.mu.Unlock()

	return i.authorize(resp, rule)
}

func (i *AuthInterceptor) authorize(p *auth.Principal, rule *options.AuthRule) (*auth.Principal, error) {
	if rule == nil {
		return p, nil
	}

	// Check Roles
	if len(rule.Roles) > 0 {
		allowed := false
		for _, r := range rule.Roles {
			for _, pr := range p.Roles {
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
			for _, pp := range p.Permissions {
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

	return p, nil
}

func (i *AuthInterceptor) getRule(fullMethod string) *options.AuthRule {
	i.mu.RLock()
	if rule, ok := i.rules[fullMethod]; ok {
		i.mu.RUnlock()
		return rule
	}
	i.mu.RUnlock()

	// Parse /Service/Method
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	if len(parts) != 2 {
		return nil
	}

	serviceName := protoreflect.FullName(parts[0])
	methodName := protoreflect.Name(parts[1])

	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(serviceName)
	if err != nil {
		return nil
	}

	serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil
	}

	methodDesc := serviceDesc.Methods().ByName(methodName)
	if methodDesc == nil {
		return nil
	}

	opts := methodDesc.Options()
	if opts == nil {
		return nil
	}

	ext := proto.GetExtension(opts, options.E_Rule)
	if rule, ok := ext.(*options.AuthRule); ok {
		i.mu.Lock()
		if i.rules == nil {
			i.rules = make(map[string]*options.AuthRule)
		}
		i.rules[fullMethod] = rule
		i.mu.Unlock()
		return rule
	}

	return nil
}

func GetPrincipal(ctx context.Context) *auth.Principal {
	p, ok := ctx.Value(PrincipalKey).(*auth.Principal)
	if !ok {
		return nil
	}
	return p
}
