package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Member represents an account-level member.
type Member = spec.Member

// AddMemberRequest is the request body for adding a member to the account.
type AddMemberRequest = spec.AddMemberRequest

// UpdateMemberRequest is the request body for updating a member's role/status.
type UpdateMemberRequest = spec.UpdateMemberRequest

// MemberListOptions are the query parameters for listing account members.
type MemberListOptions = spec.ListMembersParams

// MemberService handles communication with the account-level member endpoints.
type MemberService interface {
	List(ctx context.Context, opts *MemberListOptions) ([]Member, *Response, error)
	Get(ctx context.Context, memberID string) (*Member, *Response, error)
	Add(ctx context.Context, req *AddMemberRequest) (*Member, *Response, error)
	Update(ctx context.Context, memberID string, req *UpdateMemberRequest) (*Member, *Response, error)
	Remove(ctx context.Context, memberID string) (*Response, error)
}

// MemberServiceOp implements MemberService.
type MemberServiceOp struct {
	client *Client
}

var _ MemberService = &MemberServiceOp{}

func (s *MemberServiceOp) List(ctx context.Context, opts *MemberListOptions) ([]Member, *Response, error) {
	resp, err := s.client.spec.ListMembersWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var members []Member
	if resp.JSON200.Data != nil {
		members = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return members, responseFrom(resp.HTTPResponse, total), nil
}

func (s *MemberServiceOp) Get(ctx context.Context, memberID string) (*Member, *Response, error) {
	id, err := parseUUID(memberID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetMemberWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *MemberServiceOp) Add(ctx context.Context, req *AddMemberRequest) (*Member, *Response, error) {
	resp, err := s.client.spec.AddMemberWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *MemberServiceOp) Update(ctx context.Context, memberID string, req *UpdateMemberRequest) (*Member, *Response, error) {
	id, err := parseUUID(memberID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateMemberWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *MemberServiceOp) Remove(ctx context.Context, memberID string) (*Response, error) {
	id, err := parseUUID(memberID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RemoveMemberWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
