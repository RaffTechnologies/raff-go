package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Region represents a Raff datacenter region.
type Region = spec.Region

// Template represents an OS template available for VM creation.
type Template = spec.Template

// TemplateListOptions are the query parameters for filtering templates.
type TemplateListOptions = spec.ListTemplatesParams

// MetadataService exposes read-only catalog endpoints (templates, regions)
// that don't fit any one resource. They power VM/volume creation flows
// and the public docs.
type MetadataService interface {
	ListRegions(ctx context.Context) ([]Region, *Response, error)
	ListTemplates(ctx context.Context, opts *TemplateListOptions) ([]Template, *Response, error)
}

// MetadataServiceOp implements MetadataService.
type MetadataServiceOp struct {
	client *Client
}

var _ MetadataService = &MetadataServiceOp{}

func (s *MetadataServiceOp) ListRegions(ctx context.Context) ([]Region, *Response, error) {
	resp, err := s.client.spec.ListRegionsWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var regions []Region
	if resp.JSON200.Regions != nil {
		regions = *resp.JSON200.Regions
	}
	return regions, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *MetadataServiceOp) ListTemplates(ctx context.Context, opts *TemplateListOptions) ([]Template, *Response, error) {
	resp, err := s.client.spec.ListTemplatesWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var templates []Template
	if resp.JSON200.Data != nil {
		templates = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return templates, responseFrom(resp.HTTPResponse, total), nil
}
