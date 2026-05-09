package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Snapshot represents a point-in-time snapshot of a VM disk or volume.
type Snapshot = spec.Snapshot

// CreateSnapshotRequest is the request body for creating a snapshot.
type CreateSnapshotRequest = spec.CreateSnapshotRequest

// RenameSnapshotRequest is the request body for renaming a snapshot.
type RenameSnapshotRequest = spec.RenameSnapshotRequest

// SnapshotListOptions are the query parameters for listing snapshots.
type SnapshotListOptions = spec.ListSnapshotsParams

// SnapshotService handles communication with the snapshot endpoints.
type SnapshotService interface {
	List(ctx context.Context, opts *SnapshotListOptions) ([]Snapshot, *Response, error)
	Get(ctx context.Context, snapshotID int) (*Snapshot, *Response, error)
	Create(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, *Response, error)
	Rename(ctx context.Context, snapshotID int, req *RenameSnapshotRequest) (*Snapshot, *Response, error)
	Restore(ctx context.Context, snapshotID int) (*Response, error)
	Delete(ctx context.Context, snapshotID int) (*Response, error)
}

// SnapshotServiceOp implements SnapshotService.
type SnapshotServiceOp struct {
	client *Client
}

var _ SnapshotService = &SnapshotServiceOp{}

func (s *SnapshotServiceOp) List(ctx context.Context, opts *SnapshotListOptions) ([]Snapshot, *Response, error) {
	if opts == nil {
		opts = &SnapshotListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListSnapshotsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var snaps []Snapshot
	if resp.JSON200.Data != nil {
		snaps = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return snaps, responseFrom(resp.HTTPResponse, total), nil
}

func (s *SnapshotServiceOp) Get(ctx context.Context, snapshotID int) (*Snapshot, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetSnapshotWithResponse(ctx, snapshotID, &spec.GetSnapshotParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SnapshotServiceOp) Create(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateSnapshotWithResponse(ctx, &spec.CreateSnapshotParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SnapshotServiceOp) Rename(ctx context.Context, snapshotID int, req *RenameSnapshotRequest) (*Snapshot, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.RenameSnapshotWithResponse(ctx, snapshotID, &spec.RenameSnapshotParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SnapshotServiceOp) Restore(ctx context.Context, snapshotID int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RestoreSnapshotWithResponse(ctx, snapshotID, &spec.RestoreSnapshotParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SnapshotServiceOp) Delete(ctx context.Context, snapshotID int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteSnapshotWithResponse(ctx, snapshotID, &spec.DeleteSnapshotParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
