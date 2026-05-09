package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Permission represents a single permission identifier (e.g. "vm.create")
// along with its description and applicable scope.
type Permission = spec.Permission

// PermissionListOptions are the query parameters for listing permissions.
type PermissionListOptions = spec.ListPermissionsParams

// PermissionService handles communication with the permissions catalog
// endpoint. Permissions themselves are read-only — they're built into the
// platform; only roles bind a set of permissions to members.
type PermissionService interface {
	List(ctx context.Context, opts *PermissionListOptions) ([]Permission, *Response, error)
}

// PermissionServiceOp implements PermissionService.
type PermissionServiceOp struct {
	client *Client
}

var _ PermissionService = &PermissionServiceOp{}

func (s *PermissionServiceOp) List(ctx context.Context, opts *PermissionListOptions) ([]Permission, *Response, error) {
	resp, err := s.client.spec.ListPermissionsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var perms []Permission
	if resp.JSON200.Data != nil {
		perms = *resp.JSON200.Data
	}
	return perms, responseFrom(resp.HTTPResponse, 0), nil
}
