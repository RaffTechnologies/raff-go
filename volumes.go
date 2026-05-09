package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Volume represents a block storage volume.
type Volume = spec.Volume

// CreateVolumeRequest is the request body for creating a volume.
type CreateVolumeRequest = spec.CreateVolumeRequest

// ResizeVolumeRequest is the request body for resizing a volume.
type ResizeVolumeRequest = spec.ResizeVolumeRequest

// AttachVolumeRequest is the request body for attaching a volume to a VM.
type AttachVolumeRequest = spec.AttachVolumeRequest

// VolumeListOptions are the query parameters for listing volumes.
type VolumeListOptions = spec.ListVolumesParams

// VolumeService handles communication with the volume endpoints.
type VolumeService interface {
	List(ctx context.Context, opts *VolumeListOptions) ([]Volume, *Response, error)
	Get(ctx context.Context, volumeID int) (*Volume, *Response, error)
	Create(ctx context.Context, req *CreateVolumeRequest) (*Volume, *Response, error)
	Delete(ctx context.Context, volumeID int) (*Response, error)
	Resize(ctx context.Context, volumeID int, req *ResizeVolumeRequest) (*ResizeResponse, *Response, error)
	Attach(ctx context.Context, volumeID int, req *AttachVolumeRequest) (*Volume, *Response, error)
	Detach(ctx context.Context, volumeID int) (*Response, error)
}

// VolumeServiceOp implements VolumeService.
type VolumeServiceOp struct {
	client *Client
}

var _ VolumeService = &VolumeServiceOp{}

func (s *VolumeServiceOp) List(ctx context.Context, opts *VolumeListOptions) ([]Volume, *Response, error) {
	if opts == nil {
		opts = &VolumeListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListVolumesWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var vols []Volume
	if resp.JSON200.Data != nil {
		vols = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return vols, responseFrom(resp.HTTPResponse, total), nil
}

func (s *VolumeServiceOp) Get(ctx context.Context, volumeID int) (*Volume, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetVolumeWithResponse(ctx, volumeID, &spec.GetVolumeParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VolumeServiceOp) Create(ctx context.Context, req *CreateVolumeRequest) (*Volume, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateVolumeWithResponse(ctx, &spec.CreateVolumeParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	// Volumes are async — created with HTTP 202.
	if resp.JSON202 == nil || resp.JSON202.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON202.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VolumeServiceOp) Delete(ctx context.Context, volumeID int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteVolumeWithResponse(ctx, volumeID, &spec.DeleteVolumeParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VolumeServiceOp) Resize(ctx context.Context, volumeID int, req *ResizeVolumeRequest) (*ResizeResponse, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ResizeVolumeWithResponse(ctx, volumeID, &spec.ResizeVolumeParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VolumeServiceOp) Attach(ctx context.Context, volumeID int, req *AttachVolumeRequest) (*Volume, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AttachVolumeWithResponse(ctx, volumeID, &spec.AttachVolumeParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VolumeServiceOp) Detach(ctx context.Context, volumeID int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DetachVolumeWithResponse(ctx, volumeID, &spec.DetachVolumeParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
