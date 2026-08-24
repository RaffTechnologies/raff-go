package raff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Functions service — hand-written wrappers, wire-identical to the published
// spec (docs/api-reference/openapi.yaml, which now includes all Functions
// endpoints; spec.gen.go carries the generated client). Migrating these
// wrappers onto the generated client is a mechanical follow-up — the public
// types and behavior here already match the spec.

// Function represents a deployed serverless function.
type Function struct {
	ID                    string `json:"id"`
	FunctionID            string `json:"function_id"`
	Name                  string `json:"name"`
	Slug                  string `json:"slug"`
	Description           string `json:"description"`
	Runtime               string `json:"runtime"`
	Region                string `json:"region"`
	MemoryMB              int    `json:"memory_mb"`
	TimeoutSeconds        int    `json:"timeout_seconds"`
	ExtendedTimeout       bool   `json:"extended_timeout"`
	MinScale              int    `json:"min_scale"`
	MaxScale              int    `json:"max_scale"`
	SourceType            string `json:"source_type"`
	RepoFullName          string `json:"repo_full_name"`
	RepoBranch            string `json:"repo_branch"`
	RepoRootDir           string `json:"repo_root_dir"`
	URL                   string `json:"url"`
	Status                string `json:"status"`
	StatusMessage         string `json:"status_message"`
	CurrentRevisionNumber int    `json:"current_revision_number"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

// CreateFunctionRequest is the request body for creating a function.
type CreateFunctionRequest struct {
	Name           string `json:"name"`
	Slug           string `json:"slug,omitempty"`
	Description    string `json:"description,omitempty"`
	Runtime        string `json:"runtime"`
	Region         string `json:"region,omitempty"`
	MemoryMB       int    `json:"memory_mb,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	MinScale       int    `json:"min_scale,omitempty"`
	MaxScale       int    `json:"max_scale,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
	// ExtendedTimeout opts into timeouts above 3600s (max 86400s / 24h).
	ExtendedTimeout bool `json:"extended_timeout,omitempty"`
}

// FunctionDeployment represents one build+rollout of a function.
type FunctionDeployment struct {
	ID              string `json:"id"`
	Channel         string `json:"channel"`
	Status          string `json:"status"`
	CommitSHA       string `json:"commit_sha"`
	Error           string `json:"error"`
	BuildDurationMS int    `json:"build_duration_ms"`
	CreatedAt       string `json:"created_at"`
	FinishedAt      string `json:"finished_at"`
}

// CreateFunctionDeploymentRequest starts a deployment from staged source.
type CreateFunctionDeploymentRequest struct {
	Channel       string `json:"channel,omitempty"`
	SourceBlobKey string `json:"source_blob_key,omitempty"`
	Template      string `json:"template,omitempty"`
	Message       string `json:"message,omitempty"`
}

// FunctionSourceUpload is a presigned upload slot for a source zip.
type FunctionSourceUpload struct {
	UploadURL        string `json:"upload_url"`
	SourceBlobKey    string `json:"source_blob_key"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// FunctionBuildLogLine is one line of build output.
type FunctionBuildLogLine struct {
	Timestamp string `json:"timestamp"`
	Step      string `json:"step"`
	Message   string `json:"message"`
}

// FunctionEnvVar is one environment variable on a function.
type FunctionEnvVar struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
	IsSystem bool   `json:"is_system"`
}

// SetFunctionEnvVarRequest sets one environment variable.
type SetFunctionEnvVarRequest struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret,omitempty"`
}

// ConvertLambdaRequest is the Lambda import request (stateless conversion).
type ConvertLambdaRequest struct {
	LambdaRuntime     string            `json:"lambda_runtime"`
	Files             []SourceFile      `json:"files"`
	Handler           string            `json:"handler,omitempty"`
	MemoryMB          int               `json:"memory_mb,omitempty"`
	TimeoutSeconds    int               `json:"timeout_seconds,omitempty"`
	EnvVars           map[string]string `json:"env_vars,omitempty"`
	TriggerConfigJSON string            `json:"trigger_config_json,omitempty"`
	// Mode: "rewrite" (AI, default) or "adapter" (run as-is, no AI).
	Mode string `json:"mode,omitempty"`
}

// SourceFile is one source file (input or converted output).
type SourceFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ConvertLambdaResult is the conversion output — review before deploying.
type ConvertLambdaResult struct {
	Runtime          string       `json:"runtime"`
	ConvertedFiles   []SourceFile `json:"converted_files"`
	RaffToml         string       `json:"raff_toml"`
	ProposedTriggers []string     `json:"proposed_triggers"`
	Warnings         []string     `json:"warnings"`
	OutOfScope       []string     `json:"out_of_scope"`
}

// UpdateFunctionRequest partially updates runtime settings (nil = keep).
// Live functions roll changes out as a zero-downtime config redeploy.
type UpdateFunctionRequest struct {
	Description     *string `json:"description,omitempty"`
	MemoryMB        *int    `json:"memory_mb,omitempty"`
	TimeoutSeconds  *int    `json:"timeout_seconds,omitempty"`
	ExtendedTimeout *bool   `json:"extended_timeout,omitempty"`
	MinScale        *int    `json:"min_scale,omitempty"`
	MaxScale        *int    `json:"max_scale,omitempty"`
}

// FunctionService handles communication with the Functions endpoints.
type FunctionService interface {
	List(ctx context.Context) ([]Function, *Response, error)
	Get(ctx context.Context, ref string) (*Function, *Response, error)
	Create(ctx context.Context, req *CreateFunctionRequest) (*Function, *Response, error)
	Update(ctx context.Context, ref string, req *UpdateFunctionRequest) (*Function, *Response, error)
	Delete(ctx context.Context, ref string) (*Response, error)
	RequestSourceUpload(ctx context.Context, ref string) (*FunctionSourceUpload, *Response, error)
	UploadSource(ctx context.Context, uploadURL string, zipData []byte) error
	CreateDeployment(ctx context.Context, ref string, req *CreateFunctionDeploymentRequest) (*FunctionDeployment, *Response, error)
	GetDeployment(ctx context.Context, ref, deploymentID string) (*FunctionDeployment, *Response, error)
	GetDeploymentLogs(ctx context.Context, ref, deploymentID string) ([]FunctionBuildLogLine, bool, *Response, error)
	ListEnvVars(ctx context.Context, ref string, reveal bool) ([]FunctionEnvVar, *Response, error)
	SetEnvVar(ctx context.Context, ref string, req *SetFunctionEnvVarRequest) (*Response, error)
	ConvertLambda(ctx context.Context, req *ConvertLambdaRequest) (*ConvertLambdaResult, *Response, error)
}

// FunctionServiceOp implements FunctionService.
type FunctionServiceOp struct {
	client *Client
}

var _ FunctionService = &FunctionServiceOp{}

// doJSON performs a hand-written request against the Functions API and
// decodes the standard `{success, data}` envelope into out.
func (s *FunctionServiceOp) doJSON(ctx context.Context, method, path string, body any, out any) (*Response, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.client.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if err := s.client.injectHeaders(ctx, req); err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.client.projectID != "" {
		req.Header.Set("X-Project-ID", s.client.projectID)
	}
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseFrom(resp, 0), err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return responseFrom(resp, 0), errorFromResponse(resp.StatusCode, raw)
	}
	if out != nil {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return responseFrom(resp, 0), fmt.Errorf("decode response: %w", err)
		}
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return responseFrom(resp, 0), fmt.Errorf("decode response data: %w", err)
		}
	}
	return responseFrom(resp, 0), nil
}

func (s *FunctionServiceOp) List(ctx context.Context) ([]Function, *Response, error) {
	var data struct {
		Functions []Function `json:"functions"`
		Total     int        `json:"total"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/functions", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	if resp != nil {
		resp.Total = data.Total
	}
	return data.Functions, resp, nil
}

func (s *FunctionServiceOp) Get(ctx context.Context, ref string) (*Function, *Response, error) {
	var data struct {
		Function *Function `json:"function"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/functions/"+url.PathEscape(ref), nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Function, resp, nil
}

func (s *FunctionServiceOp) Create(ctx context.Context, req *CreateFunctionRequest) (*Function, *Response, error) {
	var data struct {
		Function *Function `json:"function"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/functions", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Function, resp, nil
}

func (s *FunctionServiceOp) RequestSourceUpload(ctx context.Context, ref string) (*FunctionSourceUpload, *Response, error) {
	var data FunctionSourceUpload
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/functions/"+url.PathEscape(ref)+"/source-upload", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return &data, resp, nil
}

// UploadSource PUTs the source zip to a presigned URL (no API auth involved).
func (s *FunctionServiceOp) UploadSource(ctx context.Context, uploadURL string, zipData []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(zipData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/zip")
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("source upload failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *FunctionServiceOp) CreateDeployment(ctx context.Context, ref string, req *CreateFunctionDeploymentRequest) (*FunctionDeployment, *Response, error) {
	var data struct {
		Deployment *FunctionDeployment `json:"deployment"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/functions/"+url.PathEscape(ref)+"/deployments", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployment, resp, nil
}

func (s *FunctionServiceOp) GetDeployment(ctx context.Context, ref, deploymentID string) (*FunctionDeployment, *Response, error) {
	var data struct {
		Deployment *FunctionDeployment `json:"deployment"`
	}
	path := "/api/v1/functions/" + url.PathEscape(ref) + "/deployments/" + url.PathEscape(deploymentID)
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployment, resp, nil
}

func (s *FunctionServiceOp) GetDeploymentLogs(ctx context.Context, ref, deploymentID string) ([]FunctionBuildLogLine, bool, *Response, error) {
	var data struct {
		Lines    []FunctionBuildLogLine `json:"lines"`
		Complete bool                   `json:"complete"`
	}
	path := "/api/v1/functions/" + url.PathEscape(ref) + "/deployments/" + url.PathEscape(deploymentID) + "/logs"
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, false, resp, err
	}
	return data.Lines, data.Complete, resp, nil
}

func (s *FunctionServiceOp) ListEnvVars(ctx context.Context, ref string, reveal bool) ([]FunctionEnvVar, *Response, error) {
	var data struct {
		EnvVars []FunctionEnvVar `json:"env_vars"`
	}
	path := "/api/v1/functions/" + url.PathEscape(ref) + "/env"
	if reveal {
		path += "/reveal"
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.EnvVars, resp, nil
}

func (s *FunctionServiceOp) SetEnvVar(ctx context.Context, ref string, req *SetFunctionEnvVarRequest) (*Response, error) {
	return s.doJSON(ctx, http.MethodPut, "/api/v1/functions/"+url.PathEscape(ref)+"/env", req, nil)
}

// ConvertLambda converts AWS Lambda source to a portable Raff handler.
// Stateless — nothing is stored; review the output, then deploy it.
func (s *FunctionServiceOp) ConvertLambda(ctx context.Context, req *ConvertLambdaRequest) (*ConvertLambdaResult, *Response, error) {
	var data ConvertLambdaResult
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/functions/import/lambda", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return &data, resp, nil
}

// Update partially updates function settings; live functions get a
// zero-downtime config redeploy.
func (s *FunctionServiceOp) Update(ctx context.Context, ref string, req *UpdateFunctionRequest) (*Function, *Response, error) {
	var data struct {
		Function *Function `json:"function"`
	}
	resp, err := s.doJSON(ctx, http.MethodPatch, "/api/v1/functions/"+url.PathEscape(ref), req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Function, resp, nil
}

// Delete removes a function and all revisions (asynchronous; the name and
// URL are released once deletion completes).
func (s *FunctionServiceOp) Delete(ctx context.Context, ref string) (*Response, error) {
	return s.doJSON(ctx, http.MethodDelete, "/api/v1/functions/"+url.PathEscape(ref), nil, nil)
}
