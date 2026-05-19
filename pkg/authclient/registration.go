package authclient

import (
	"context"
	"fmt"

	"github.com/squall-chua/go-grpc-auth/api/v1/admin"
	"github.com/squall-chua/go-grpc-auth/api/v1/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ExtractRolesAndPermissions extracts all unique role and permission names
// declared in options.v1.rule annotations across the given service descriptors.
func ExtractRolesAndPermissions(services ...protoreflect.ServiceDescriptor) (roles, permissions []string) {
	roleSet := make(map[string]bool)
	permSet := make(map[string]bool)

	for _, sd := range services {
		methods := sd.Methods()
		for i := 0; i < methods.Len(); i++ {
			md := methods.Get(i)
			opts, ok := md.Options().(*descriptorpb.MethodOptions)
			if !ok || opts == nil {
				continue
			}
			if !proto.HasExtension(opts, options.E_Rule) {
				continue
			}
			rule, ok := proto.GetExtension(opts, options.E_Rule).(*options.AuthRule)
			if !ok || rule == nil || rule.Public {
				continue
			}
			for _, r := range rule.Roles {
				roleSet[r] = true
			}
			for _, p := range rule.Permissions {
				permSet[p] = true
			}
		}
	}

	for r := range roleSet {
		roles = append(roles, r)
	}
	for p := range permSet {
		permissions = append(permissions, p)
	}
	return roles, permissions
}

// RegisterServiceRolesAndPermissions extracts roles and permissions from the
// given service descriptors and registers them with the auth server. Existing
// roles/permissions with the same (namespace, name) are left unchanged.
func (c *Client) RegisterServiceRolesAndPermissions(
	ctx context.Context,
	namespace string,
	description string,
	services ...protoreflect.ServiceDescriptor,
) error {
	roles, permissions := ExtractRolesAndPermissions(services...)

	for _, name := range permissions {
		_, err := c.Admin.CreatePermission(ctx, &admin.CreatePermissionRequest{
			Name:        name,
			Namespace:   namespace,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("registering permission %q: %w", name, err)
		}
	}

	for _, name := range roles {
		_, err := c.Admin.CreateRole(ctx, &admin.CreateRoleRequest{
			Name:      name,
			Namespace: namespace,
		})
		if err != nil {
			return fmt.Errorf("registering role %q: %w", name, err)
		}
	}

	return nil
}
