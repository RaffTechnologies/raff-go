package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// Project represents a Raff project. Aliased to the generated spec type so
// new fields propagate automatically when the OpenAPI spec changes.
type Project = spec.Project

// CreateProjectRequest is the request body for creating a project.
type CreateProjectRequest = spec.CreateProjectRequest

// UpdateProjectRequest is the request body for updating a project.
type UpdateProjectRequest = spec.UpdateProjectRequest

// ProjectListOptions are the query parameters for listing projects.
type ProjectListOptions = spec.ListProjectsParams

// ProjectService handles communication with the project endpoints.
type ProjectService interface {
	List(ctx context.Context, opts *ProjectListOptions) ([]Project, *Response, error)
	Get(ctx context.Context, projectID string) (*Project, *Response, error)
	Create(ctx context.Context, req *CreateProjectRequest) (*Project, *Response, error)
	Update(ctx context.Context, projectID string, req *UpdateProjectRequest) (*Project, *Response, error)
	Delete(ctx context.Context, projectID string) (*Response, error)
}

// ProjectServiceOp implements ProjectService.
type ProjectServiceOp struct {
	client *Client
}

var _ ProjectService = &ProjectServiceOp{}

// List returns all projects for the authenticated account.
func (s *ProjectServiceOp) List(ctx context.Context, opts *ProjectListOptions) ([]Project, *Response, error) {
	resp, err := s.client.spec.ListProjectsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var projects []Project
	if resp.JSON200.Data != nil {
		projects = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return projects, responseFrom(resp.HTTPResponse, total), nil
}

// Get returns a single project by ID.
func (s *ProjectServiceOp) Get(ctx context.Context, projectID string) (*Project, *Response, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetProjectWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// Create creates a new project.
func (s *ProjectServiceOp) Create(ctx context.Context, req *CreateProjectRequest) (*Project, *Response, error) {
	resp, err := s.client.spec.CreateProjectWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// Update updates an existing project.
func (s *ProjectServiceOp) Update(ctx context.Context, projectID string, req *UpdateProjectRequest) (*Project, *Response, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateProjectWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// Delete deletes a project.
func (s *ProjectServiceOp) Delete(ctx context.Context, projectID string) (*Response, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteProjectWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
