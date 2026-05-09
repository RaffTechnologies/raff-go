package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Role represents an IAM role (account-level or project-level).
type Role = spec.Role

// CreateRoleRequest is the request body for creating a custom role.
type CreateRoleRequest = spec.CreateRoleRequest

// UpdateRoleRequest is the request body for updating a custom role.
type UpdateRoleRequest = spec.UpdateRoleRequest

// RoleListOptions are the query parameters for listing roles.
type RoleListOptions = spec.ListRolesParams

// RoleService handles communication with the role endpoints.
//
// System roles (Owner, Admin, Member, etc.) are immutable and cannot be
// updated or deleted; the Update / Delete methods will return an error
// from the API for system roles.
type RoleService interface {
	List(ctx context.Context, opts *RoleListOptions) ([]Role, *Response, error)
	Get(ctx context.Context, roleID string) (*Role, *Response, error)
	Create(ctx context.Context, req *CreateRoleRequest) (*Role, *Response, error)
	Update(ctx context.Context, roleID string, req *UpdateRoleRequest) (*Role, *Response, error)
	Delete(ctx context.Context, roleID string) (*Response, error)
}

// RoleServiceOp implements RoleService.
type RoleServiceOp struct {
	client *Client
}

var _ RoleService = &RoleServiceOp{}

func (s *RoleServiceOp) List(ctx context.Context, opts *RoleListOptions) ([]Role, *Response, error) {
	resp, err := s.client.spec.ListRolesWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var roles []Role
	if resp.JSON200.Data != nil {
		roles = *resp.JSON200.Data
	}
	return roles, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *RoleServiceOp) Get(ctx context.Context, roleID string) (*Role, *Response, error) {
	id, err := parseUUID(roleID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetRoleWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *RoleServiceOp) Create(ctx context.Context, req *CreateRoleRequest) (*Role, *Response, error) {
	resp, err := s.client.spec.CreateRoleWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *RoleServiceOp) Update(ctx context.Context, roleID string, req *UpdateRoleRequest) (*Role, *Response, error) {
	id, err := parseUUID(roleID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateRoleWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *RoleServiceOp) Delete(ctx context.Context, roleID string) (*Response, error) {
	id, err := parseUUID(roleID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteRoleWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
