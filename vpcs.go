package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// VPC represents a virtual private cloud.
type VPC = spec.VPC

// VPCDetail is the richer response returned by GET /api/v1/vpcs/{id}: the
// VPC plus its allocatable IP range and the active leases on attached NICs.
type VPCDetail = spec.VPCDetail

// VPCLease is a single IP lease on a VPC NIC.
type VPCLease = spec.VPCLease

// CreateVPCRequest is the request body for creating a VPC.
type CreateVPCRequest = spec.CreateVPCRequest

// UpdateVPCRequest is the request body for updating a VPC.
type UpdateVPCRequest = spec.UpdateVPCRequest

// CIDRSuggestionsResponse contains a recommended CIDR and alternatives.
type CIDRSuggestionsResponse = spec.CIDRSuggestionsResponse

// CIDRSuggestion is a single CIDR suggestion entry.
type CIDRSuggestion = spec.CIDRSuggestion

// VPCListOptions are the query parameters for listing VPCs.
type VPCListOptions = spec.ListVPCsParams

// VPCService handles communication with the VPC endpoints.
type VPCService interface {
	List(ctx context.Context, opts *VPCListOptions) ([]VPC, *Response, error)
	Get(ctx context.Context, vpcID string) (*VPC, *Response, error)
	GetDetail(ctx context.Context, vpcID string) (*VPCDetail, *Response, error)
	Create(ctx context.Context, req *CreateVPCRequest) (*VPC, *Response, error)
	Update(ctx context.Context, vpcID string, req *UpdateVPCRequest) (*VPC, *Response, error)
	Delete(ctx context.Context, vpcID string) (*Response, error)
	CIDRSuggestions(ctx context.Context) (*CIDRSuggestionsResponse, *Response, error)
}

// VPCServiceOp implements VPCService.
type VPCServiceOp struct {
	client *Client
}

var _ VPCService = &VPCServiceOp{}

func (s *VPCServiceOp) List(ctx context.Context, opts *VPCListOptions) ([]VPC, *Response, error) {
	if opts == nil {
		opts = &VPCListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListVPCsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var vpcs []VPC
	if resp.JSON200.Data != nil {
		vpcs = *resp.JSON200.Data
	}
	return vpcs, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VPCServiceOp) Get(ctx context.Context, vpcID string) (*VPC, *Response, error) {
	detail, r, err := s.GetDetail(ctx, vpcID)
	if err != nil {
		return nil, r, err
	}
	return &detail.Vpc, r, nil
}

func (s *VPCServiceOp) GetDetail(ctx context.Context, vpcID string) (*VPCDetail, *Response, error) {
	id, err := parseUUID(vpcID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetVPCWithResponse(ctx, id, &spec.GetVPCParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VPCServiceOp) Create(ctx context.Context, req *CreateVPCRequest) (*VPC, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateVPCWithResponse(ctx, &spec.CreateVPCParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VPCServiceOp) Update(ctx context.Context, vpcID string, req *UpdateVPCRequest) (*VPC, *Response, error) {
	id, err := parseUUID(vpcID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateVPCWithResponse(ctx, id, &spec.UpdateVPCParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VPCServiceOp) Delete(ctx context.Context, vpcID string) (*Response, error) {
	id, err := parseUUID(vpcID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteVPCWithResponse(ctx, id, &spec.DeleteVPCParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VPCServiceOp) CIDRSuggestions(ctx context.Context) (*CIDRSuggestionsResponse, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListVPCCIDRSuggestionsWithResponse(ctx, &spec.ListVPCCIDRSuggestionsParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}
