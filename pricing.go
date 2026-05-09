package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// VMPricingPlan represents a single VM size/plan with pricing.
type VMPricingPlan = spec.VMPricingPlan

// StoragePricing is the per-GB storage pricing shape used by volume,
// backup, and snapshot pricing endpoints.
type StoragePricing = spec.StoragePricing

// IPPricing groups IP pricing by family (ipv4, ipv6).
type IPPricing = spec.IPPricing

// VMPricingListOptions filters VM pricing by region / type.
type VMPricingListOptions = spec.ListVMPricingParams

// VolumePricingListOptions filters volume pricing by region.
type VolumePricingListOptions = spec.ListVolumePricingParams

// BackupPricingListOptions filters backup pricing by region.
type BackupPricingListOptions = spec.ListBackupPricingParams

// SnapshotPricingListOptions filters snapshot pricing by region.
type SnapshotPricingListOptions = spec.ListSnapshotPricingParams

// PricingService exposes the public pricing catalog. All endpoints are
// read-only and do not require an API key — they back the marketing
// pricing pages and `vm create` size pickers.
type PricingService interface {
	ListVM(ctx context.Context, opts *VMPricingListOptions) ([]VMPricingPlan, *Response, error)
	ListVolume(ctx context.Context, opts *VolumePricingListOptions) (*StoragePricing, *Response, error)
	ListBackup(ctx context.Context, opts *BackupPricingListOptions) (*StoragePricing, *Response, error)
	ListSnapshot(ctx context.Context, opts *SnapshotPricingListOptions) (*StoragePricing, *Response, error)
	ListIP(ctx context.Context) (*IPPricing, *Response, error)
}

// PricingServiceOp implements PricingService.
type PricingServiceOp struct {
	client *Client
}

var _ PricingService = &PricingServiceOp{}

func (s *PricingServiceOp) ListVM(ctx context.Context, opts *VMPricingListOptions) ([]VMPricingPlan, *Response, error) {
	resp, err := s.client.spec.ListVMPricingWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var plans []VMPricingPlan
	if resp.JSON200.Plans != nil {
		plans = *resp.JSON200.Plans
	}
	return plans, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *PricingServiceOp) ListVolume(ctx context.Context, opts *VolumePricingListOptions) (*StoragePricing, *Response, error) {
	resp, err := s.client.spec.ListVolumePricingWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Pricing == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Pricing, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *PricingServiceOp) ListBackup(ctx context.Context, opts *BackupPricingListOptions) (*StoragePricing, *Response, error) {
	resp, err := s.client.spec.ListBackupPricingWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Pricing == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Pricing, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *PricingServiceOp) ListSnapshot(ctx context.Context, opts *SnapshotPricingListOptions) (*StoragePricing, *Response, error) {
	resp, err := s.client.spec.ListSnapshotPricingWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Pricing == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Pricing, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *PricingServiceOp) ListIP(ctx context.Context) (*IPPricing, *Response, error) {
	resp, err := s.client.spec.ListIPPricingWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Pricing == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Pricing, responseFrom(resp.HTTPResponse, 0), nil
}
