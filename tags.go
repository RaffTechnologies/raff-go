package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Tag is an account tag: one name, colour and optional priority shared by every server that has it,
// with the firewalls it carries per project.
type Tag = spec.AccountTag

// TagColor is a tag's colour: gray, red, orange, amber, green, blue, violet or pink.
type TagColor = spec.TagColor

// TagFirewallLink is a tag that carries a firewall in one project.
type TagFirewallLink = spec.TagFirewallLink

// CreateTagRequest is the request body for creating a tag.
type CreateTagRequest = spec.CreateTagRequest

// UpdateTagRequest is the request body for updating a tag. A Priority of 0 removes the priority.
type UpdateTagRequest = spec.UpdateTagRequest

// TagListOptions are the optional parameters for listing tags. XProjectID counts servers of
// one project only.
type TagListOptions = spec.ListTagsParams

// TagService handles communication with the account tag endpoints.
type TagService interface {
	List(ctx context.Context, opts *TagListOptions) ([]Tag, *Response, error)
	Create(ctx context.Context, req *CreateTagRequest) (*Tag, *Response, error)
	Update(ctx context.Context, tagID string, req *UpdateTagRequest) (*Tag, *Response, error)
	Delete(ctx context.Context, tagID string) (*Response, error)
}

// TagServiceOp implements TagService.
type TagServiceOp struct {
	client *Client
}

var _ TagService = &TagServiceOp{}

func (s *TagServiceOp) List(ctx context.Context, opts *TagListOptions) ([]Tag, *Response, error) {
	if opts == nil {
		opts = &TagListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListTagsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var out []Tag
	if resp.JSON200.Data != nil {
		out = *resp.JSON200.Data
	}
	return out, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *TagServiceOp) Create(ctx context.Context, req *CreateTagRequest) (*Tag, *Response, error) {
	resp, err := s.client.spec.CreateTagWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *TagServiceOp) Update(ctx context.Context, tagID string, req *UpdateTagRequest) (*Tag, *Response, error) {
	id, err := parseUUID(tagID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateTagWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *TagServiceOp) Delete(ctx context.Context, tagID string) (*Response, error) {
	id, err := parseUUID(tagID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteTagWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
