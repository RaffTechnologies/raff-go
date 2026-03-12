package raff

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const projectsBasePath = "/api/v1/projects"

// ProjectService handles communication with the project related methods of the Raff API.
type ProjectService interface {
	List(ctx context.Context, opts *ListOptions) ([]Project, *Response, error)
	Get(ctx context.Context, projectID string) (*Project, *Response, error)
	Create(ctx context.Context, req *ProjectCreateRequest) (*Project, *Response, error)
	Update(ctx context.Context, projectID string, req *ProjectUpdateRequest) (*Project, *Response, error)
	Delete(ctx context.Context, projectID string) (*Response, error)
}

// ProjectServiceOp implements ProjectService.
type ProjectServiceOp struct {
	client *Client
}

var _ ProjectService = &ProjectServiceOp{}

// Project represents a Raff project.
type Project struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"account_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	DefaultRegion string    `json:"default_region"`
	IsDefault     bool      `json:"is_default"`
	IsActive      bool      `json:"is_active"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProjectCreateRequest represents a request to create a project.
type ProjectCreateRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	DefaultRegion string `json:"default_region,omitempty"`
}

// ProjectUpdateRequest represents a request to update a project.
type ProjectUpdateRequest struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	DefaultRegion string `json:"default_region,omitempty"`
}

// List returns all projects for the authenticated account.
func (s *ProjectServiceOp) List(ctx context.Context, opts *ListOptions) ([]Project, *Response, error) {
	path := addListOptions(projectsBasePath, opts)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var projects []Project
	resp, err := s.client.Do(ctx, req, &projects)
	if err != nil {
		return nil, resp, err
	}

	return projects, resp, nil
}

// Get returns a single project by ID.
func (s *ProjectServiceOp) Get(ctx context.Context, projectID string) (*Project, *Response, error) {
	path := fmt.Sprintf("%s/%s", projectsBasePath, projectID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var project Project
	resp, err := s.client.Do(ctx, req, &project)
	if err != nil {
		return nil, resp, err
	}

	return &project, resp, nil
}

// Create creates a new project.
func (s *ProjectServiceOp) Create(ctx context.Context, createReq *ProjectCreateRequest) (*Project, *Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, projectsBasePath, createReq)
	if err != nil {
		return nil, nil, err
	}

	var project Project
	resp, err := s.client.Do(ctx, req, &project)
	if err != nil {
		return nil, resp, err
	}

	return &project, resp, nil
}

// Update updates an existing project.
func (s *ProjectServiceOp) Update(ctx context.Context, projectID string, updateReq *ProjectUpdateRequest) (*Project, *Response, error) {
	path := fmt.Sprintf("%s/%s", projectsBasePath, projectID)

	req, err := s.client.NewRequest(ctx, http.MethodPut, path, updateReq)
	if err != nil {
		return nil, nil, err
	}

	var project Project
	resp, err := s.client.Do(ctx, req, &project)
	if err != nil {
		return nil, resp, err
	}

	return &project, resp, nil
}

// Delete deletes a project.
func (s *ProjectServiceOp) Delete(ctx context.Context, projectID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", projectsBasePath, projectID)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
