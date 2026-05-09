package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Invitation represents a pending invite to join an account or project.
type Invitation = spec.Invitation

// CreateInvitationRequest is the request body for both account and project
// invitations — same shape (email + role_id).
type CreateInvitationRequest = spec.CreateInvitationRequest

// InvitationService handles communication with the invitation endpoints.
//
// The public spec exposes Create (account-scoped or project-scoped) and
// Cancel. There is no List endpoint — invitations are observed via the
// invited member appearing in the relevant Members or ProjectMembers list
// with status "pending".
type InvitationService interface {
	CreateAccount(ctx context.Context, req *CreateInvitationRequest) (*Invitation, *Response, error)
	CreateProject(ctx context.Context, projectID string, req *CreateInvitationRequest) (*Invitation, *Response, error)
	Cancel(ctx context.Context, invitationID string) (*Response, error)
}

// InvitationServiceOp implements InvitationService.
type InvitationServiceOp struct {
	client *Client
}

var _ InvitationService = &InvitationServiceOp{}

func (s *InvitationServiceOp) CreateAccount(ctx context.Context, req *CreateInvitationRequest) (*Invitation, *Response, error) {
	resp, err := s.client.spec.CreateAccountInvitationWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *InvitationServiceOp) CreateProject(ctx context.Context, projectID string, req *CreateInvitationRequest) (*Invitation, *Response, error) {
	pid, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateProjectInvitationWithResponse(ctx, pid, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *InvitationServiceOp) Cancel(ctx context.Context, invitationID string) (*Response, error) {
	id, err := parseUUID(invitationID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.CancelInvitationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
