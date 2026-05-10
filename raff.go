// Package raff provides a Go client library for the Raff Cloud API.
//
// Usage:
//
//	client := raff.NewFromToken("raff_pub_xxx")
//	projects, _, err := client.Projects.List(ctx, nil)
//
// The library is a thin idiomatic wrapper over types and an HTTP client
// generated from the public OpenAPI spec at docs/api-reference/openapi.yaml.
// Run `make generate` after editing the spec.
package raff

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/rafftechnologies/raff-go/spec"
)

const (
	// Version is the client library version.
	Version = "0.3.2"

	defaultBaseURL   = "https://api.rafftechnologies.com"
	defaultUserAgent = "raff-go/" + Version
)

// Client manages communication with the Raff Cloud API.
type Client struct {
	spec      *spec.ClientWithResponses
	apiKey    string
	projectID string
	userAgent string

	// Services
	Projects        ProjectService
	VMs             VMService
	VPCs            VPCService
	IPs             IPService
	SecurityGroups  SecurityGroupService
	SSHKeys         SSHKeyService
	APIKeys         APIKeyService
	Members         MemberService
	ProjectMembers  ProjectMemberService
	Roles           RoleService
	Permissions     PermissionService
	Invitations     InvitationService
	Volumes         VolumeService
	Snapshots       SnapshotService
	Backups         BackupService
	BackupSchedules BackupScheduleService
	Metadata        MetadataService
	Pricing         PricingService
}

// NewFromToken creates a new Raff API client with the given API key.
func NewFromToken(apiKey string, opts ...ClientOpt) *Client {
	return New(&http.Client{Timeout: 30 * time.Second}, apiKey, opts...)
}

// New creates a new Raff API client with a custom HTTP client.
func New(httpClient *http.Client, apiKey string, opts ...ClientOpt) *Client {
	c := &Client{
		apiKey:    apiKey,
		userAgent: defaultUserAgent,
	}

	cfg := clientConfig{baseURL: defaultBaseURL}
	for _, opt := range opts {
		opt(&cfg, c)
	}

	specClient, err := spec.NewClientWithResponses(
		cfg.baseURL,
		spec.WithHTTPClient(httpClient),
		spec.WithRequestEditorFn(c.injectHeaders),
	)
	if err != nil {
		// Only fails on invalid base URL — fall back to default.
		specClient, _ = spec.NewClientWithResponses(
			defaultBaseURL,
			spec.WithHTTPClient(httpClient),
			spec.WithRequestEditorFn(c.injectHeaders),
		)
	}
	c.spec = specClient

	c.Projects = &ProjectServiceOp{client: c}
	c.VMs = &VMServiceOp{client: c}
	c.VPCs = &VPCServiceOp{client: c}
	c.IPs = &IPServiceOp{client: c}
	c.SecurityGroups = &SecurityGroupServiceOp{client: c}
	c.SSHKeys = &SSHKeyServiceOp{client: c}
	c.APIKeys = &APIKeyServiceOp{client: c}
	c.Members = &MemberServiceOp{client: c}
	c.ProjectMembers = &ProjectMemberServiceOp{client: c}
	c.Roles = &RoleServiceOp{client: c}
	c.Permissions = &PermissionServiceOp{client: c}
	c.Invitations = &InvitationServiceOp{client: c}
	c.Volumes = &VolumeServiceOp{client: c}
	c.Snapshots = &SnapshotServiceOp{client: c}
	c.Backups = &BackupServiceOp{client: c}
	c.BackupSchedules = &BackupScheduleServiceOp{client: c}
	c.Metadata = &MetadataServiceOp{client: c}
	c.Pricing = &PricingServiceOp{client: c}

	return c
}

// clientConfig captures options that affect spec client construction.
type clientConfig struct {
	baseURL string
}

// ClientOpt is a functional option for configuring the client.
type ClientOpt func(*clientConfig, *Client)

// SetBaseURL sets the API base URL.
func SetBaseURL(baseURL string) ClientOpt {
	return func(cfg *clientConfig, _ *Client) {
		baseURL = strings.TrimRight(baseURL, "/")
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "https://" + baseURL
		}
		cfg.baseURL = baseURL
	}
}

// SetUserAgent sets the User-Agent header.
func SetUserAgent(ua string) ClientOpt {
	return func(_ *clientConfig, c *Client) {
		c.userAgent = ua
	}
}

// SetProjectID sets the default project ID sent via X-Project-ID header for
// mutating operations. Each service method may override it per call.
func SetProjectID(id string) ClientOpt {
	return func(_ *clientConfig, c *Client) {
		c.projectID = id
	}
}

func (c *Client) injectHeaders(_ context.Context, req *http.Request) error {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	return nil
}

// requireProjectID returns the configured project ID as a UUID, or an error
// if not set. Used by service methods that need X-Project-ID.
//
// The error message is intentionally neutral so it reads well when surfaced
// by raff-cli (which would otherwise leak `raff.SetProjectID` into terminal
// output, confusing users who aren't Go developers). CLI consumers should
// pre-check via their own helper so this rarely fires; when it does the
// message points at the underlying X-Project-ID requirement rather than at
// any specific Go API.
func (c *Client) requireProjectID() (openapi_types.UUID, error) {
	if c.projectID == "" {
		return openapi_types.UUID{}, fmt.Errorf("project ID required (X-Project-ID header). " +
			"Set --project-id, RAFF_PROJECT_ID, or run `raff configure` if using the CLI; " +
			"set raff.SetProjectID(...) when using the SDK directly")
	}
	return parseUUID(c.projectID)
}

// optionalProjectID returns a pointer to the configured project ID UUID, or
// (nil, nil) when no project is configured. Used by list endpoints that allow
// listing across all accessible projects when X-Project-ID is omitted.
//
// CONVENTION for `List` methods: every List wrapper that has an optional
// project-scope filter (whether the spec exposes it as an `X-Project-ID`
// header or a `project_id` query param) MUST auto-fill it from the client
// before sending. Pattern:
//
//	if opts == nil {
//	    opts = &XListOptions{}
//	}
//	if opts.XProjectID == nil { // or opts.ProjectID for query-param variants
//	    if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
//	        opts.XProjectID = pid
//	    }
//	}
//
// Without this, a project-scoped client (raff.SetProjectID(...)) silently
// returns resources across every project the API key can see — a data-leak
// bug across customers. The `make verify` CI guards spec drift but not
// wrapper-correctness — keep this pattern consistent when adding new
// services (volumes, snapshots, ssh keys, …).
func (c *Client) optionalProjectID() (*openapi_types.UUID, error) {
	if c.projectID == "" {
		return nil, nil
	}
	id, err := parseUUID(c.projectID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// parseUUID converts a string ID into the UUID type used by the spec.
func parseUUID(s string) (openapi_types.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid UUID %q: %w", s, err)
	}
	return id, nil
}

// Response wraps the underlying http.Response for callers that need access
// to status codes, headers, or pagination metadata.
type Response struct {
	*http.Response
	Total int
}

func responseFrom(httpResp *http.Response, total int) *Response {
	return &Response{Response: httpResp, Total: total}
}

// ErrorResponse is returned when the API returns a non-2xx status code.
type ErrorResponse struct {
	StatusCode int
	Message    string
	Reason     string
	Body       string
}

func (e *ErrorResponse) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Body)
}

// errorFromResponse builds an ErrorResponse from raw HTTP response data.
func errorFromResponse(statusCode int, body []byte) error {
	var apiErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	}

	errResp := &ErrorResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}
	if json.Unmarshal(body, &apiErr) == nil {
		if apiErr.Message != "" {
			errResp.Message = apiErr.Message
		} else {
			errResp.Message = apiErr.Error
		}
		errResp.Reason = apiErr.Reason
	}
	return errResp
}

// String returns a pointer to v. Use it when constructing request types
// that have optional string fields.
func String(v string) *string { return &v }

// Int returns a pointer to v. Use it when constructing request types
// that have optional int fields.
func Int(v int) *int { return &v }

// Bool returns a pointer to v. Use it when constructing request types
// that have optional bool fields.
func Bool(v bool) *bool { return &v }

// StringValue dereferences p, returning "" if nil.
func StringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// IntValue dereferences p, returning 0 if nil.
func IntValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// BoolValue dereferences p, returning false if nil.
func BoolValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

