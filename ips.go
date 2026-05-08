package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// FloatingIP represents a floating public IP address.
type FloatingIP = spec.FloatingIP

// ReserveIPRequest is the request body for reserving a floating IP.
type ReserveIPRequest = spec.ReserveIPRequest

// IPListOptions are the query parameters for listing floating IPs.
type IPListOptions = spec.ListIPsParams

// ChangeIPResponse is the response shape for swapping a reserved IP.
type ChangeIPResponse struct {
	Success bool        `json:"success"`
	OldIP   *FloatingIP `json:"old_ip,omitempty"`
	NewIP   *FloatingIP `json:"new_ip,omitempty"`
}

// IPService handles communication with the floating IP endpoints.
type IPService interface {
	List(ctx context.Context, opts *IPListOptions) ([]FloatingIP, *Response, error)
	Get(ctx context.Context, ipID string) (*FloatingIP, *Response, error)
	Reserve(ctx context.Context, req *ReserveIPRequest) (*FloatingIP, *Response, error)
	Release(ctx context.Context, ipID string) (*Response, error)
	Change(ctx context.Context, ipID string) (*ChangeIPResponse, *Response, error)
}

// IPServiceOp implements IPService.
type IPServiceOp struct {
	client *Client
}

var _ IPService = &IPServiceOp{}

func (s *IPServiceOp) List(ctx context.Context, opts *IPListOptions) ([]FloatingIP, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	if opts == nil {
		opts = &IPListOptions{}
	}
	opts.XProjectID = projectID
	resp, err := s.client.spec.ListIPsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var ips []FloatingIP
	if resp.JSON200.Data != nil {
		ips = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return ips, responseFrom(resp.HTTPResponse, total), nil
}

func (s *IPServiceOp) Get(ctx context.Context, ipID string) (*FloatingIP, *Response, error) {
	id, err := parseUUID(ipID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetIPWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *IPServiceOp) Reserve(ctx context.Context, req *ReserveIPRequest) (*FloatingIP, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	body := ReserveIPRequest{}
	if req != nil {
		body = *req
	}
	resp, err := s.client.spec.ReserveIPWithResponse(ctx, &spec.ReserveIPParams{XProjectID: projectID}, body)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *IPServiceOp) Release(ctx context.Context, ipID string) (*Response, error) {
	id, err := parseUUID(ipID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.ReleaseIPWithResponse(ctx, id, &spec.ReleaseIPParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *IPServiceOp) Change(ctx context.Context, ipID string) (*ChangeIPResponse, *Response, error) {
	id, err := parseUUID(ipID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ChangeIPWithResponse(ctx, id, &spec.ChangeIPParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	out := &ChangeIPResponse{}
	if resp.JSON200.Success != nil {
		out.Success = *resp.JSON200.Success
	}
	out.OldIP = resp.JSON200.OldIP
	out.NewIP = resp.JSON200.NewIP
	return out, responseFrom(resp.HTTPResponse, 0), nil
}
