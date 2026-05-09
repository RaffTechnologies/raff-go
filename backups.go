package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Backup represents a managed backup of a VM or volume.
type Backup = spec.Backup

// CreateBackupRequest is the request body for taking an on-demand backup.
type CreateBackupRequest = spec.CreateBackupRequest

// BackupListOptions are the query parameters for listing backups.
type BackupListOptions = spec.ListBackupsParams

// BackupSchedule defines a recurring backup policy.
type BackupSchedule = spec.BackupSchedule

// CreateBackupScheduleRequest is the request body for creating a schedule.
type CreateBackupScheduleRequest = spec.CreateBackupScheduleRequest

// UpdateBackupScheduleRequest is the request body for updating a schedule.
type UpdateBackupScheduleRequest = spec.UpdateBackupScheduleRequest

// BackupScheduleListOptions are the query parameters for listing schedules.
type BackupScheduleListOptions = spec.ListBackupSchedulesParams

// BackupService handles communication with the backup endpoints. Restore
// is asynchronous and returns the in-progress backup record (HTTP 202).
type BackupService interface {
	List(ctx context.Context, opts *BackupListOptions) ([]Backup, *Response, error)
	Get(ctx context.Context, backupID string) (*Backup, *Response, error)
	Create(ctx context.Context, req *CreateBackupRequest) (*Backup, *Response, error)
	Restore(ctx context.Context, backupID string) (*Backup, *Response, error)
	Delete(ctx context.Context, backupID string) (*Response, error)
}

// BackupScheduleService handles backup schedule CRUD.
type BackupScheduleService interface {
	List(ctx context.Context, opts *BackupScheduleListOptions) ([]BackupSchedule, *Response, error)
	Get(ctx context.Context, scheduleID int) (*BackupSchedule, *Response, error)
	Create(ctx context.Context, req *CreateBackupScheduleRequest) (*BackupSchedule, *Response, error)
	Update(ctx context.Context, scheduleID int, req *UpdateBackupScheduleRequest) (*BackupSchedule, *Response, error)
	Delete(ctx context.Context, scheduleID int) (*Response, error)
}

// BackupServiceOp implements BackupService.
type BackupServiceOp struct {
	client *Client
}

var _ BackupService = &BackupServiceOp{}

func (s *BackupServiceOp) List(ctx context.Context, opts *BackupListOptions) ([]Backup, *Response, error) {
	if opts == nil {
		opts = &BackupListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListBackupsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var items []Backup
	if resp.JSON200.Data != nil {
		items = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return items, responseFrom(resp.HTTPResponse, total), nil
}

func (s *BackupServiceOp) Get(ctx context.Context, backupID string) (*Backup, *Response, error) {
	id, err := parseUUID(backupID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetBackupWithResponse(ctx, id, &spec.GetBackupParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupServiceOp) Create(ctx context.Context, req *CreateBackupRequest) (*Backup, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateBackupWithResponse(ctx, &spec.CreateBackupParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	// Backups are async (HTTP 202).
	if resp.JSON202 == nil || resp.JSON202.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON202.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupServiceOp) Restore(ctx context.Context, backupID string) (*Backup, *Response, error) {
	id, err := parseUUID(backupID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.RestoreBackupWithResponse(ctx, id, &spec.RestoreBackupParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON202 == nil || resp.JSON202.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON202.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupServiceOp) Delete(ctx context.Context, backupID string) (*Response, error) {
	id, err := parseUUID(backupID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteBackupWithResponse(ctx, id, &spec.DeleteBackupParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

// BackupScheduleServiceOp implements BackupScheduleService.
type BackupScheduleServiceOp struct {
	client *Client
}

var _ BackupScheduleService = &BackupScheduleServiceOp{}

func (s *BackupScheduleServiceOp) List(ctx context.Context, opts *BackupScheduleListOptions) ([]BackupSchedule, *Response, error) {
	if opts == nil {
		opts = &BackupScheduleListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListBackupSchedulesWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var items []BackupSchedule
	if resp.JSON200.Data != nil {
		items = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return items, responseFrom(resp.HTTPResponse, total), nil
}

func (s *BackupScheduleServiceOp) Get(ctx context.Context, scheduleID int) (*BackupSchedule, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetBackupScheduleWithResponse(ctx, scheduleID, &spec.GetBackupScheduleParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupScheduleServiceOp) Create(ctx context.Context, req *CreateBackupScheduleRequest) (*BackupSchedule, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateBackupScheduleWithResponse(ctx, &spec.CreateBackupScheduleParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupScheduleServiceOp) Update(ctx context.Context, scheduleID int, req *UpdateBackupScheduleRequest) (*BackupSchedule, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateBackupScheduleWithResponse(ctx, scheduleID, &spec.UpdateBackupScheduleParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *BackupScheduleServiceOp) Delete(ctx context.Context, scheduleID int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteBackupScheduleWithResponse(ctx, scheduleID, &spec.DeleteBackupScheduleParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
