package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// ProjectMember represents a project-level member.
type ProjectMember = spec.ProjectMember

// AddProjectMemberRequest is the request body for adding a project member.
type AddProjectMemberRequest = spec.AddProjectMemberRequest

// UpdateProjectMemberRequest is the request body for updating a project
// member. The spec reuses UpdateMemberRequest for both account and project
// member updates — same fields (role, status).
type UpdateProjectMemberRequest = spec.UpdateMemberRequest

// ProjectMemberListOptions are the query parameters for listing project members.
type ProjectMemberListOptions = spec.ListProjectMembersParams

// ProjectMemberService handles communication with the project member endpoints.
//
// All operations require a project ID — these are project-scoped, not
// account-scoped.
type ProjectMemberService interface {
	List(ctx context.Context, projectID string, opts *ProjectMemberListOptions) ([]ProjectMember, *Response, error)
	Get(ctx context.Context, projectID, memberID string) (*ProjectMember, *Response, error)
	Add(ctx context.Context, projectID string, req *AddProjectMemberRequest) (*ProjectMember, *Response, error)
	Update(ctx context.Context, projectID, memberID string, req *UpdateProjectMemberRequest) (*ProjectMember, *Response, error)
	Remove(ctx context.Context, projectID, memberID string) (*Response, error)
}

// ProjectMemberServiceOp implements ProjectMemberService.
type ProjectMemberServiceOp struct {
	client *Client
}

var _ ProjectMemberService = &ProjectMemberServiceOp{}

func (s *ProjectMemberServiceOp) List(ctx context.Context, projectID string, opts *ProjectMemberListOptions) ([]ProjectMember, *Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListProjectMembersWithResponse(ctx, pid, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var members []ProjectMember
	if resp.JSON200.Data != nil {
		members = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return members, responseFrom(resp.HTTPResponse, total), nil
}

func (s *ProjectMemberServiceOp) Get(ctx context.Context, projectID, memberID string) (*ProjectMember, *Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	mid, err := parseUUID(memberID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetProjectMemberWithResponse(ctx, pid, mid)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *ProjectMemberServiceOp) Add(ctx context.Context, projectID string, req *AddProjectMemberRequest) (*ProjectMember, *Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AddProjectMemberWithResponse(ctx, pid, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *ProjectMemberServiceOp) Update(ctx context.Context, projectID, memberID string, req *UpdateProjectMemberRequest) (*ProjectMember, *Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	mid, err := parseUUID(memberID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateProjectMemberWithResponse(ctx, pid, mid, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *ProjectMemberServiceOp) Remove(ctx context.Context, projectID, memberID string) (*Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, err
	}
	mid, err := parseUUID(memberID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RemoveProjectMemberWithResponse(ctx, pid, mid)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
