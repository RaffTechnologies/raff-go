package raff

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/rafftechnologies/raff-go/spec"
)

// SecurityGroup represents a named set of network rules.
type SecurityGroup = spec.SecurityGroup

// SecurityGroupRule is a single inbound or outbound rule.
type SecurityGroupRule = spec.SecurityGroupRule

// SecurityGroupTemplate is a pre-built rule set you can clone when creating a group.
type SecurityGroupTemplate = spec.SecurityGroupTemplate

// CreateSecurityGroupRequest is the request body for creating a security group.
type CreateSecurityGroupRequest = spec.CreateSecurityGroupRequest

// UpdateSecurityGroupRequest is the request body for updating a security group.
type UpdateSecurityGroupRequest = spec.UpdateSecurityGroupRequest

// SecurityGroupListOptions are query params for listing security groups.
type SecurityGroupListOptions = spec.ListSecurityGroupsParams

// FirewallTags is the tags a security group is linked to in its project, with the progress of
// putting it on their servers.
type FirewallTags = spec.FirewallTags

// SecurityGroupService handles communication with the security group endpoints.
type SecurityGroupService interface {
	List(ctx context.Context, opts *SecurityGroupListOptions) ([]SecurityGroup, *Response, error)
	Templates(ctx context.Context) ([]SecurityGroupTemplate, *Response, error)
	Get(ctx context.Context, sgID string) (*SecurityGroup, *Response, error)
	Create(ctx context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, *Response, error)
	Update(ctx context.Context, sgID string, req *UpdateSecurityGroupRequest) (*SecurityGroup, *Response, error)
	Delete(ctx context.Context, sgID string) (*Response, error)
	GetTags(ctx context.Context, sgID string) (*FirewallTags, *Response, error)
	SetTags(ctx context.Context, sgID string, tagIDs []string) (*FirewallTags, *Response, error)
}

// SecurityGroupServiceOp implements SecurityGroupService.
type SecurityGroupServiceOp struct {
	client *Client
}

var _ SecurityGroupService = &SecurityGroupServiceOp{}

func (s *SecurityGroupServiceOp) List(ctx context.Context, opts *SecurityGroupListOptions) ([]SecurityGroup, *Response, error) {
	if opts == nil {
		opts = &SecurityGroupListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListSecurityGroupsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var out []SecurityGroup
	if resp.JSON200.Data != nil {
		out = *resp.JSON200.Data
	}
	return out, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SecurityGroupServiceOp) Templates(ctx context.Context) ([]SecurityGroupTemplate, *Response, error) {
	resp, err := s.client.spec.ListSecurityGroupTemplatesWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var out []SecurityGroupTemplate
	if resp.JSON200.Data != nil {
		out = *resp.JSON200.Data
	}
	return out, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SecurityGroupServiceOp) Get(ctx context.Context, sgID string) (*SecurityGroup, *Response, error) {
	id, err := parseUUID(sgID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetSecurityGroupWithResponse(ctx, id, &spec.GetSecurityGroupParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SecurityGroupServiceOp) Create(ctx context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateSecurityGroupWithResponse(ctx, &spec.CreateSecurityGroupParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SecurityGroupServiceOp) Update(ctx context.Context, sgID string, req *UpdateSecurityGroupRequest) (*SecurityGroup, *Response, error) {
	id, err := parseUUID(sgID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateSecurityGroupWithResponse(ctx, id, &spec.UpdateSecurityGroupParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SecurityGroupServiceOp) Delete(ctx context.Context, sgID string) (*Response, error) {
	id, err := parseUUID(sgID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteSecurityGroupWithResponse(ctx, id, &spec.DeleteSecurityGroupParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

// GetTags returns the tags the security group is linked to in its project and its job progress.
func (s *SecurityGroupServiceOp) GetTags(ctx context.Context, sgID string) (*FirewallTags, *Response, error) {
	id, err := parseUUID(sgID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetFirewallTagsWithResponse(ctx, id, &spec.GetFirewallTagsParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// SetTags makes tagIDs the security group's tags in its own project (an empty list unlinks every
// tag). Raff then puts the security group on every server of that project with one of the tags.
func (s *SecurityGroupServiceOp) SetTags(ctx context.Context, sgID string, tagIDs []string) (*FirewallTags, *Response, error) {
	id, err := parseUUID(sgID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]openapi_types.UUID, 0, len(tagIDs))
	for _, t := range tagIDs {
		tid, err := parseUUID(t)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, tid)
	}
	resp, err := s.client.spec.SetFirewallTagsWithResponse(ctx, id, spec.SetFirewallTagsRequest{TagIds: ids})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}
