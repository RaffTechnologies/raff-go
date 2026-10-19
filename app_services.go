package raff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Apps service (Raff Apps / PaaS) — hand-written wrappers, wire-identical to
// the published spec (docs/api-reference/openapi.yaml). spec.gen.go carries the
// generated client; migrating these wrappers onto it is a mechanical follow-up
// — the public types and behavior here already match the spec.

// AppService represents a deployed app service (web, private, worker, cron, or
// job) on Raff Apps.
type AppService struct {
	ID                  string `json:"id"`
	ServiceID           string `json:"service_id"`
	ProjectID           string `json:"project_id"`
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	Description         string `json:"description"`
	ServiceType         string `json:"service_type"`
	SourceType          string `json:"source_type"`
	Builder             string `json:"builder"`
	RepoFullName        string `json:"repo_full_name"`
	RepoBranch          string `json:"repo_branch"`
	RepoRootDir         string `json:"repo_root_dir"`
	DockerfilePath      string `json:"dockerfile_path"`
	ImageRef            string `json:"image_ref"`
	StartCommand        string `json:"start_command"`
	TierID              int    `json:"tier_id"`
	TierName            string `json:"tier_name"`
	Replicas            int    `json:"replicas"`
	AutoscalingEnabled  bool   `json:"autoscaling_enabled"`
	MinReplicas         int    `json:"min_replicas"`
	MaxReplicas         int    `json:"max_replicas"`
	AutoscalingMetric   string `json:"autoscaling_metric"`
	AutoscalingTarget   int    `json:"autoscaling_target"`
	ScaleToZero         bool   `json:"scale_to_zero"`
	HTTPPort            int    `json:"http_port"`
	HealthCheckPath     string `json:"health_check_path"`
	CronSchedule        string `json:"cron_schedule"`
	CronTimezone        string `json:"cron_timezone"`
	Region              string `json:"region"`
	URL                 string `json:"url"`
	InternalHost        string `json:"internal_host"`
	Status              string `json:"status"`
	StatusMessage       string `json:"status_message"`
	PausedByCap         bool   `json:"paused_by_cap"`
	BillingType         string `json:"billing_type"`
	CurrentDeploymentID string `json:"current_deployment_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`

	CronConcurrency         string `json:"cron_concurrency"`
	PrimaryDomain           string `json:"primary_domain"`
	PreviewsEnabled         bool   `json:"previews_enabled"`
	PreviewCount            int    `json:"preview_count"`
	MaxPreviews             int    `json:"max_previews"`
	WaitingForDatabase      bool   `json:"waiting_for_database"`
	GitHubStatus            string `json:"github_status"`
	GitHubLogin             string `json:"github_login"`
	GitHubSettingsURL       string `json:"github_settings_url"`
	CurrentDeploymentNumber int    `json:"current_deployment_number"`
}

// CreateAppServiceRequest is the request body for creating an app service.
type CreateAppServiceRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	ServiceType     string `json:"service_type"`
	SourceType      string `json:"source_type,omitempty"`
	Builder         string `json:"builder,omitempty"`
	ImageRef        string `json:"image_ref,omitempty"`
	RepoFullName    string `json:"repo_full_name,omitempty"`
	RepoBranch      string `json:"repo_branch,omitempty"`
	RepoRootDir     string `json:"repo_root_dir,omitempty"`
	DockerfilePath  string `json:"dockerfile_path,omitempty"`
	StartCommand    string `json:"start_command,omitempty"`
	TierID          int    `json:"tier_id,omitempty"`
	Replicas        int    `json:"replicas,omitempty"`
	Autoscaling     bool   `json:"autoscaling_enabled,omitempty"`
	MinReplicas     int    `json:"min_replicas,omitempty"`
	MaxReplicas     int    `json:"max_replicas,omitempty"`
	ScaleToZero     bool   `json:"scale_to_zero,omitempty"`
	HTTPPort        int    `json:"http_port,omitempty"`
	HealthCheckPath string `json:"health_check_path,omitempty"`
	CronSchedule    string `json:"cron_schedule,omitempty"`
	CronTimezone    string `json:"cron_timezone,omitempty"`
	Region          string `json:"region,omitempty"`
	// DeployNow immediately queues a build+rollout for prebuilt-image
	// services. Source-based services deploy after their source is uploaded.
	DeployNow bool `json:"deploy_now,omitempty"`

	// GitHubInstallationID is the ID of a GitHub connection (ListGitHubConnections)
	// for SourceType "git".
	GitHubInstallationID string `json:"github_installation_id,omitempty"`
	// CronConcurrency is forbid (default), allow or replace.
	CronConcurrency string `json:"cron_concurrency,omitempty"`
	// EnvVars are set before the first deploy.
	EnvVars []AppEnvVarInput `json:"env_vars,omitempty"`
	// LinkDatabaseID links a managed database: the app deploys once the
	// database runs, with DATABASE_URL and the other DATABASE_* variables set.
	LinkDatabaseID string `json:"link_database_id,omitempty"`
}

// AppEnvVarInput is one environment variable set at create. Scope is
// runtime (default), build or both.
type AppEnvVarInput struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret,omitempty"`
	Scope    string `json:"scope,omitempty"`
}

// UpdateAppServiceRequest changes an app's settings. Only the fields you set
// are sent. A change that affects how a live app runs (port, start command,
// health check, size) always rolls out as a new deploy; Redeploy true also
// redeploys after a change that would not need it.
type UpdateAppServiceRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	TierID          *int    `json:"tier_id,omitempty"`
	StartCommand    *string `json:"start_command,omitempty"`
	HTTPPort        *int    `json:"http_port,omitempty"`
	HealthCheckPath *string `json:"health_check_path,omitempty"`
	CronSchedule    *string `json:"cron_schedule,omitempty"`
	CronTimezone    *string `json:"cron_timezone,omitempty"`
	CronConcurrency *string `json:"cron_concurrency,omitempty"`
	RepoBranch      *string `json:"repo_branch,omitempty"`
	RepoRootDir     *string `json:"repo_root_dir,omitempty"`
	DockerfilePath  *string `json:"dockerfile_path,omitempty"`
	PreviewsEnabled *bool   `json:"previews_enabled,omitempty"`
	Redeploy        *bool   `json:"redeploy,omitempty"`
}

// AppTemplate is a ready-to-run app: create it with SourceType "image" and
// the template's ImageRef, HTTPPort, Env and HealthCheckPath.
type AppTemplate struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	ServiceType      string            `json:"service_type"`
	ImageRef         string            `json:"image_ref"`
	Builder          string            `json:"builder"`
	DefaultTierID    int               `json:"default_tier_id"`
	MinTierID        int               `json:"min_tier_id"`
	HTTPPort         int               `json:"http_port"`
	HealthCheckPath  string            `json:"health_check_path"`
	Companions       []string          `json:"companions"`
	DatabaseOptional bool              `json:"database_optional"`
	Env              map[string]string `json:"env"`
	Popular          bool              `json:"popular"`
}

// AppGitHubConnection is a GitHub account or organization connected to Raff.
type AppGitHubConnection struct {
	ID             string `json:"id"`
	InstallationID int64  `json:"installation_id"`
	GitHubLogin    string `json:"github_login"`
	AccountType    string `json:"account_type"`
	Status         string `json:"status"`
	Suspended      bool   `json:"suspended"`
	SettingsURL    string `json:"settings_url"`
	CreatedAt      string `json:"created_at"`
}

// AppBinding links an app to a managed database.
type AppBinding struct {
	ID            string   `json:"id"`
	ServiceID     string   `json:"service_id"`
	BindingType   string   `json:"binding_type"`
	ResourceID    string   `json:"resource_id"`
	ResourceName  string   `json:"resource_name"`
	EnvPrefix     string   `json:"env_prefix"`
	InjectedKeys  []string `json:"injected_keys"`
	Status        string   `json:"status"`
	StatusMessage string   `json:"status_message"`
	CreatedAt     string   `json:"created_at"`
}

// CreateAppBindingRequest links a database to an app. BindingType is
// "database"; EnvPrefix is optional (DATABASE by default).
type CreateAppBindingRequest struct {
	BindingType string `json:"binding_type"`
	ResourceID  string `json:"resource_id"`
	EnvPrefix   string `json:"env_prefix,omitempty"`
}

// AppPreview is a pull request preview of a web app from GitHub.
type AppPreview struct {
	ID                  string `json:"id"`
	ServiceID           string `json:"service_id"`
	PRNumber            int    `json:"pr_number"`
	PRTitle             string `json:"pr_title"`
	Branch              string `json:"branch"`
	HeadSHA             string `json:"head_sha"`
	Status              string `json:"status"`
	StatusMessage       string `json:"status_message"`
	URL                 string `json:"url"`
	CurrentDeploymentID string `json:"current_deployment_id"`
	ExpiresAt           string `json:"expires_at"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

// AppEvent is one line of an app's activity timeline.
type AppEvent struct {
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	EventType string `json:"event_type"`
	Message   string `json:"message"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
}

// AppInstance is one running instance of an app.
type AppInstance struct {
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Ready     bool   `json:"ready"`
	Restarts  int    `json:"restarts"`
	StartedAt string `json:"started_at"`
}

// AppTier is one Apps pricing tier (compute size + price).
type AppTier struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	VCPU          float64 `json:"vcpu"`
	MemoryMiB     int     `json:"memory_mib"`
	EphemeralGiB  int     `json:"ephemeral_gib"`
	PricePerHour  float64 `json:"price_per_hour"`
	PricePerMonth float64 `json:"price_per_month"`
	YearlyPrice   float64 `json:"yearly_price"`
	Region        string  `json:"region"`
}

// AppDeployment represents one build+rollout of an app service.
type AppDeployment struct {
	ID               string `json:"id"`
	ServiceID        string `json:"service_id"`
	DeploymentNumber int    `json:"deployment_number"`
	Channel          string `json:"channel"`
	Status           string `json:"status"`
	RollbackOf       string `json:"rollback_of"`
	ImageDigest      string `json:"image_digest"`
	CommitSHA        string `json:"commit_sha"`
	CommitMessage    string `json:"commit_message"`
	Error            string `json:"error"`
	BuildDurationMS  int    `json:"build_duration_ms"`
	StartedAt        string `json:"started_at"`
	FinishedAt       string `json:"finished_at"`
	CreatedAt        string `json:"created_at"`
}

// CreateAppDeploymentRequest starts a deployment from staged source or a
// prebuilt image.
type CreateAppDeploymentRequest struct {
	Channel       string `json:"channel,omitempty"`
	SourceBlobKey string `json:"source_blob_key,omitempty"`
	ImageRef      string `json:"image_ref,omitempty"`
}

// RollbackAppServiceRequest pins a service back to an older deployment's image.
type RollbackAppServiceRequest struct {
	TargetDeploymentID string `json:"target_deployment_id"`
}

// AppSourceUpload is a presigned upload slot for a source tarball.
type AppSourceUpload struct {
	UploadURL        string `json:"upload_url"`
	BlobKey          string `json:"blob_key"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// App env var scopes. Build-scoped values are passed to the image build, so
// frameworks that inline configuration at build time (Next.js NEXT_PUBLIC_*,
// Vite VITE_*, Create React App REACT_APP_*, SvelteKit and Astro PUBLIC_*) can
// see them. Changing one rebuilds the image rather than restarting replicas.
const (
	AppEnvScopeRuntime = "runtime"
	AppEnvScopeBuild   = "build"
	AppEnvScopeBoth    = "both"
)

// AppEnvVar is one environment variable on an app service.
type AppEnvVar struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
	IsSystem bool   `json:"is_system"`
	// Scope is runtime, build or both; empty responses predate build scopes
	// and mean runtime.
	Scope string `json:"scope,omitempty"`
}

// SetAppEnvVarRequest sets one environment variable. Redeploy defaults to true
// (nil) — env changes roll out to running replicas.
type SetAppEnvVarRequest struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret,omitempty"`
	// Scope is runtime (default), build or both. Build arguments are
	// recoverable from image history, so prefer runtime for credentials.
	Scope    string `json:"scope,omitempty"`
	Redeploy *bool  `json:"redeploy,omitempty"`
}

// BulkSetAppEnvVarsRequest upserts many environment variables at once.
type BulkSetAppEnvVarsRequest struct {
	Vars     []SetAppEnvVarRequest `json:"vars"`
	Redeploy *bool                 `json:"redeploy,omitempty"`
}

// ScaleAppServiceRequest updates replicas / autoscaling / scale-to-zero.
type ScaleAppServiceRequest struct {
	Replicas          int    `json:"replicas,omitempty"`
	AutoscalingSet    bool   `json:"autoscaling_set,omitempty"`
	Autoscaling       bool   `json:"autoscaling_enabled,omitempty"`
	MinReplicas       int    `json:"min_replicas,omitempty"`
	MaxReplicas       int    `json:"max_replicas,omitempty"`
	AutoscalingMetric string `json:"autoscaling_metric,omitempty"`
	AutoscalingTarget int    `json:"autoscaling_target,omitempty"`
	ScaleToZero       bool   `json:"scale_to_zero,omitempty"`
}

// AppJobRun is one run of a cron or one-off job service.
type AppJobRun struct {
	ID           string `json:"id"`
	ServiceID    string `json:"service_id"`
	DeploymentID string `json:"deployment_id"`
	Trigger      string `json:"trigger"`
	Status       string `json:"status"`
	ExitCode     int    `json:"exit_code"`
	Error        string `json:"error"`
	ScheduledFor string `json:"scheduled_for"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
}

// AppLogLine is one runtime log line.
type AppLogLine struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Stream    string `json:"stream"`
	Component string `json:"component"`
}

// AppBuildLogLine is one line of deployment build output.
type AppBuildLogLine struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Component string `json:"component"`
	Stream    string `json:"stream"`
}

// AppLogOptions filters a runtime log query.
type AppLogOptions struct {
	Level  string
	Search string
	Since  string
	Until  string
	Limit  int
}

// AppCustomDomain is a custom hostname attached to a web app service.
type AppCustomDomain struct {
	ID                   string `json:"id"`
	ServiceID            string `json:"service_id"`
	Domain               string `json:"domain"`
	Status               string `json:"status"`
	StatusMessage        string `json:"status_message"`
	CNAMETarget          string `json:"cname_target"`
	VerificationTXTName  string `json:"verification_txt_name"`
	VerificationTXTValue string `json:"verification_txt_value"`
	CreatedAt            string `json:"created_at"`
	LastCheckedAt        string `json:"last_checked_at"`
}

// AddAppCustomDomainRequest attaches a custom domain to a web app service.
type AddAppCustomDomainRequest struct {
	Domain string `json:"domain"`
}

// AppServiceUsage is one service's month-to-date usage and estimated charge.
type AppServiceUsage struct {
	ServiceID        string  `json:"service_id"`
	ServiceName      string  `json:"service_name"`
	TierID           int     `json:"tier_id"`
	ReplicaHours     float64 `json:"replica_hours"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// AppUsage is the account's Apps usage for a billing period.
type AppUsage struct {
	Services              []AppServiceUsage `json:"services"`
	TotalEstimatedCostUSD float64           `json:"total_estimated_cost_usd"`
	PeriodStart           string            `json:"period_start"`
	PeriodEnd             string            `json:"period_end"`
}

// AppSpendSettings is the account's Apps spend cap.
type AppSpendSettings struct {
	Enabled       bool    `json:"enabled"`
	MonthlyCapUSD float64 `json:"monthly_cap_usd"`
	CapAction     string  `json:"cap_action"`
	PausedByCap   bool    `json:"paused_by_cap"`
}

// UpdateAppSpendSettingsRequest changes the account's Apps spend cap. Only
// the fields you set are sent; the others keep their value. MonthlyCapUSD must
// be at least 1 (to remove the cap, set Enabled to false). CapAction is
// "notify_only" (email at 75%, 90% and 100%) or "pause" (also pause every
// pay-as-you-go app at 100%; prepaid apps keep running).
type UpdateAppSpendSettingsRequest struct {
	Enabled       *bool    `json:"enabled,omitempty"`
	MonthlyCapUSD *float64 `json:"monthly_cap_usd,omitempty"`
	CapAction     *string  `json:"cap_action,omitempty"`
}

// AppServiceService handles communication with the Raff Apps endpoints.
type AppServiceService interface {
	ListTiers(ctx context.Context) ([]AppTier, *Response, error)
	List(ctx context.Context) ([]AppService, *Response, error)
	Get(ctx context.Context, ref string) (*AppService, *Response, error)
	Create(ctx context.Context, req *CreateAppServiceRequest) (*AppService, *Response, error)
	Delete(ctx context.Context, ref string) (*Response, error)
	Pause(ctx context.Context, ref string) (*AppService, *Response, error)
	Resume(ctx context.Context, ref string) (*AppService, *Response, error)
	Scale(ctx context.Context, ref string, req *ScaleAppServiceRequest) (*AppService, *Response, error)
	RequestSourceUpload(ctx context.Context, ref string) (*AppSourceUpload, *Response, error)
	UploadSource(ctx context.Context, uploadURL string, tarData []byte) error
	ListDeployments(ctx context.Context, ref string, limit int) ([]AppDeployment, *Response, error)
	GetDeployment(ctx context.Context, deploymentID string) (*AppDeployment, *Response, error)
	CreateDeployment(ctx context.Context, ref string, req *CreateAppDeploymentRequest) (*AppDeployment, *Response, error)
	Rollback(ctx context.Context, ref string, req *RollbackAppServiceRequest) (*AppDeployment, *Response, error)
	GetDeploymentLogs(ctx context.Context, deploymentID string) ([]AppBuildLogLine, bool, *Response, error)
	ListLogs(ctx context.Context, ref string, opts *AppLogOptions) ([]AppLogLine, *Response, error)
	GetMetrics(ctx context.Context, ref, window string) ([]AppMetricSeries, *Response, error)
	ListEnvVars(ctx context.Context, ref string, reveal bool) ([]AppEnvVar, *Response, error)
	SetEnvVar(ctx context.Context, ref string, req *SetAppEnvVarRequest) ([]AppEnvVar, *Response, error)
	BulkSetEnvVars(ctx context.Context, ref string, req *BulkSetAppEnvVarsRequest) ([]AppEnvVar, *Response, error)
	DeleteEnvVar(ctx context.Context, ref, key string) ([]AppEnvVar, *Response, error)
	ListCustomDomains(ctx context.Context, ref string) ([]AppCustomDomain, *Response, error)
	AddCustomDomain(ctx context.Context, ref string, req *AddAppCustomDomainRequest) (*AppCustomDomain, *Response, error)
	DeleteCustomDomain(ctx context.Context, domainID string) (*Response, error)
	RetryDomainVerification(ctx context.Context, domainID string) (*AppCustomDomain, *Response, error)
	RunJob(ctx context.Context, ref string) (*AppJobRun, *Response, error)
	ListJobRuns(ctx context.Context, ref string) ([]AppJobRun, *Response, error)
	CancelJobRun(ctx context.Context, runID string) (*AppJobRun, *Response, error)
	GetUsage(ctx context.Context, ref string) (*AppUsage, *Response, error)
	GetSpendSettings(ctx context.Context) (*AppSpendSettings, *Response, error)
	UpdateSpendSettings(ctx context.Context, req *UpdateAppSpendSettingsRequest) (*AppSpendSettings, *Response, error)
	Update(ctx context.Context, ref string, req *UpdateAppServiceRequest) (*AppService, string, *Response, error)
	ListTemplates(ctx context.Context) ([]AppTemplate, *Response, error)
	ListGitHubConnections(ctx context.Context) ([]AppGitHubConnection, *Response, error)
	ListBindings(ctx context.Context, ref string) ([]AppBinding, *Response, error)
	CreateBinding(ctx context.Context, ref string, req *CreateAppBindingRequest) (*AppBinding, *Response, error)
	DeleteBinding(ctx context.Context, bindingID string) (*Response, error)
	RotateBinding(ctx context.Context, bindingID string) (*AppBinding, *Response, error)
	ListPreviews(ctx context.Context, ref string) ([]AppPreview, *Response, error)
	DeletePreview(ctx context.Context, previewID string) (*Response, error)
	RedeployPreview(ctx context.Context, previewID string) (*AppPreview, string, *Response, error)
	ListEvents(ctx context.Context, ref string, limit int) ([]AppEvent, *Response, error)
	ListInstances(ctx context.Context, ref string) ([]AppInstance, *Response, error)
}

// AppMetricSeries is one named metric series (values aligned to timestamps).
type AppMetricSeries struct {
	Name       string    `json:"name"`
	Values     []float64 `json:"values"`
	Timestamps []string  `json:"timestamps"`
}

// AppServiceServiceOp implements AppServiceService.
type AppServiceServiceOp struct {
	client *Client
}

var _ AppServiceService = &AppServiceServiceOp{}

// doJSON performs a hand-written request against the Apps API and decodes the
// response into out. Apps responses are flat `{success, <key>: ...}` envelopes
// (the gateway merges the payload keys alongside `success`), so out is
// unmarshaled from the top-level body rather than a nested `data` object.
func (s *AppServiceServiceOp) doJSON(ctx context.Context, method, path string, body any, out any) (*Response, error) {
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
		if err := json.Unmarshal(raw, out); err != nil {
			return responseFrom(resp, 0), fmt.Errorf("decode response: %w", err)
		}
	}
	return responseFrom(resp, 0), nil
}

// ListTiers returns the Apps sizes and prices.
func (s *AppServiceServiceOp) ListTiers(ctx context.Context) ([]AppTier, *Response, error) {
	var data struct {
		Tiers []AppTier `json:"tiers"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/tiers", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Tiers, resp, nil
}

func (s *AppServiceServiceOp) List(ctx context.Context) ([]AppService, *Response, error) {
	var data struct {
		Services []AppService `json:"services"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Services, resp, nil
}

func (s *AppServiceServiceOp) Get(ctx context.Context, ref string) (*AppService, *Response, error) {
	var data struct {
		Service *AppService `json:"service"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref), nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Service, resp, nil
}

func (s *AppServiceServiceOp) Create(ctx context.Context, req *CreateAppServiceRequest) (*AppService, *Response, error) {
	var data struct {
		Service *AppService `json:"service"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Service, resp, nil
}

// Delete removes an app service (asynchronous; the name and URL are released
// once deletion completes).
func (s *AppServiceServiceOp) Delete(ctx context.Context, ref string) (*Response, error) {
	return s.doJSON(ctx, http.MethodDelete, "/api/v1/apps/services/"+url.PathEscape(ref), nil, nil)
}

func (s *AppServiceServiceOp) Pause(ctx context.Context, ref string) (*AppService, *Response, error) {
	var data struct {
		Service *AppService `json:"service"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/pause", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Service, resp, nil
}

func (s *AppServiceServiceOp) Resume(ctx context.Context, ref string) (*AppService, *Response, error) {
	var data struct {
		Service *AppService `json:"service"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/resume", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Service, resp, nil
}

func (s *AppServiceServiceOp) Scale(ctx context.Context, ref string, req *ScaleAppServiceRequest) (*AppService, *Response, error) {
	var data struct {
		Service *AppService `json:"service"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/scale", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Service, resp, nil
}

func (s *AppServiceServiceOp) RequestSourceUpload(ctx context.Context, ref string) (*AppSourceUpload, *Response, error) {
	var data AppSourceUpload
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/source-upload", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return &data, resp, nil
}

// UploadSource PUTs the source tarball to a presigned URL (no API auth involved).
func (s *AppServiceServiceOp) UploadSource(ctx context.Context, uploadURL string, tarData []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(tarData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/gzip")
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("source upload failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *AppServiceServiceOp) ListDeployments(ctx context.Context, ref string, limit int) ([]AppDeployment, *Response, error) {
	var data struct {
		Deployments []AppDeployment `json:"deployments"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/deployments"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployments, resp, nil
}

func (s *AppServiceServiceOp) GetDeployment(ctx context.Context, deploymentID string) (*AppDeployment, *Response, error) {
	var data struct {
		Deployment *AppDeployment `json:"deployment"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/deployments/"+url.PathEscape(deploymentID), nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployment, resp, nil
}

func (s *AppServiceServiceOp) CreateDeployment(ctx context.Context, ref string, req *CreateAppDeploymentRequest) (*AppDeployment, *Response, error) {
	var data struct {
		Deployment *AppDeployment `json:"deployment"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/deployments", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployment, resp, nil
}

// Rollback pins a service back to an older deployment's image (creates a new
// deployment from the target's digest and config snapshot).
func (s *AppServiceServiceOp) Rollback(ctx context.Context, ref string, req *RollbackAppServiceRequest) (*AppDeployment, *Response, error) {
	var data struct {
		Deployment *AppDeployment `json:"deployment"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/rollback", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Deployment, resp, nil
}

func (s *AppServiceServiceOp) GetDeploymentLogs(ctx context.Context, deploymentID string) ([]AppBuildLogLine, bool, *Response, error) {
	var data struct {
		Lines    []AppBuildLogLine `json:"lines"`
		Complete bool              `json:"complete"`
	}
	path := "/api/v1/apps/deployments/" + url.PathEscape(deploymentID) + "/logs"
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, false, resp, err
	}
	return data.Lines, data.Complete, resp, nil
}

func (s *AppServiceServiceOp) ListLogs(ctx context.Context, ref string, opts *AppLogOptions) ([]AppLogLine, *Response, error) {
	var data struct {
		Lines []AppLogLine `json:"lines"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/logs"
	if opts != nil {
		q := url.Values{}
		if opts.Level != "" {
			q.Set("level", opts.Level)
		}
		if opts.Search != "" {
			q.Set("search", opts.Search)
		}
		if opts.Since != "" {
			q.Set("since", opts.Since)
		}
		if opts.Until != "" {
			q.Set("until", opts.Until)
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Lines, resp, nil
}

// GetMetrics returns per-service metric series over the given window
// (e.g. "1h", "24h"). An empty window uses the server default.
func (s *AppServiceServiceOp) GetMetrics(ctx context.Context, ref, window string) ([]AppMetricSeries, *Response, error) {
	var data struct {
		Series []AppMetricSeries `json:"series"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/metrics"
	if window != "" {
		path += "?window=" + url.QueryEscape(window)
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Series, resp, nil
}

// ListEnvVars lists a service's environment variables. Secret values are
// masked unless reveal is true (which requires elevated permission).
func (s *AppServiceServiceOp) ListEnvVars(ctx context.Context, ref string, reveal bool) ([]AppEnvVar, *Response, error) {
	var data struct {
		EnvVars []AppEnvVar `json:"env_vars"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/env"
	if reveal {
		path += "/reveal"
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.EnvVars, resp, nil
}

func (s *AppServiceServiceOp) SetEnvVar(ctx context.Context, ref string, req *SetAppEnvVarRequest) ([]AppEnvVar, *Response, error) {
	var data struct {
		EnvVars []AppEnvVar `json:"env_vars"`
	}
	resp, err := s.doJSON(ctx, http.MethodPut, "/api/v1/apps/services/"+url.PathEscape(ref)+"/env", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.EnvVars, resp, nil
}

func (s *AppServiceServiceOp) BulkSetEnvVars(ctx context.Context, ref string, req *BulkSetAppEnvVarsRequest) ([]AppEnvVar, *Response, error) {
	var data struct {
		EnvVars []AppEnvVar `json:"env_vars"`
	}
	resp, err := s.doJSON(ctx, http.MethodPut, "/api/v1/apps/services/"+url.PathEscape(ref)+"/env/bulk", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.EnvVars, resp, nil
}

func (s *AppServiceServiceOp) DeleteEnvVar(ctx context.Context, ref, key string) ([]AppEnvVar, *Response, error) {
	var data struct {
		EnvVars []AppEnvVar `json:"env_vars"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/env/" + url.PathEscape(key)
	resp, err := s.doJSON(ctx, http.MethodDelete, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.EnvVars, resp, nil
}

func (s *AppServiceServiceOp) ListCustomDomains(ctx context.Context, ref string) ([]AppCustomDomain, *Response, error) {
	var data struct {
		Domains []AppCustomDomain `json:"domains"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref)+"/domains", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Domains, resp, nil
}

// AddCustomDomain attaches a custom domain to a web app service. The returned
// domain carries the DNS records to create (CNAME target and verification TXT).
func (s *AppServiceServiceOp) AddCustomDomain(ctx context.Context, ref string, req *AddAppCustomDomainRequest) (*AppCustomDomain, *Response, error) {
	var data struct {
		Domain *AppCustomDomain `json:"domain"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/domains", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Domain, resp, nil
}

func (s *AppServiceServiceOp) DeleteCustomDomain(ctx context.Context, domainID string) (*Response, error) {
	return s.doJSON(ctx, http.MethodDelete, "/api/v1/apps/domains/"+url.PathEscape(domainID), nil, nil)
}

func (s *AppServiceServiceOp) RetryDomainVerification(ctx context.Context, domainID string) (*AppCustomDomain, *Response, error) {
	var data struct {
		Domain *AppCustomDomain `json:"domain"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/domains/"+url.PathEscape(domainID)+"/retry", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Domain, resp, nil
}

// RunJob triggers a manual run of a cron or one-off job service.
func (s *AppServiceServiceOp) RunJob(ctx context.Context, ref string) (*AppJobRun, *Response, error) {
	var data struct {
		Run *AppJobRun `json:"run"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/jobs/run", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Run, resp, nil
}

func (s *AppServiceServiceOp) ListJobRuns(ctx context.Context, ref string) ([]AppJobRun, *Response, error) {
	var data struct {
		Runs []AppJobRun `json:"runs"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref)+"/jobs/runs", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Runs, resp, nil
}

func (s *AppServiceServiceOp) CancelJobRun(ctx context.Context, runID string) (*AppJobRun, *Response, error) {
	var data struct {
		Run *AppJobRun `json:"run"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/jobs/runs/"+url.PathEscape(runID)+"/cancel", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Run, resp, nil
}

// GetUsage returns the account's month-to-date Apps usage and estimated
// charge. Pass an empty ref for all services, or a service ref to scope it.
func (s *AppServiceServiceOp) GetUsage(ctx context.Context, ref string) (*AppUsage, *Response, error) {
	var data AppUsage
	path := "/api/v1/apps/usage"
	if ref != "" {
		path += "?service_id=" + url.QueryEscape(ref)
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return &data, resp, nil
}

func (s *AppServiceServiceOp) GetSpendSettings(ctx context.Context) (*AppSpendSettings, *Response, error) {
	var data struct {
		Settings *AppSpendSettings `json:"settings"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/spend-settings", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Settings, resp, nil
}

func (s *AppServiceServiceOp) UpdateSpendSettings(ctx context.Context, req *UpdateAppSpendSettingsRequest) (*AppSpendSettings, *Response, error) {
	var data struct {
		Settings *AppSpendSettings `json:"settings"`
	}
	resp, err := s.doJSON(ctx, http.MethodPatch, "/api/v1/apps/spend-settings", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Settings, resp, nil
}

// Update changes an app's settings. The returned deployment ID is set when
// the change rolled out as a new deploy.
func (s *AppServiceServiceOp) Update(ctx context.Context, ref string, req *UpdateAppServiceRequest) (*AppService, string, *Response, error) {
	var data struct {
		Service      *AppService `json:"service"`
		DeploymentID string      `json:"deployment_id"`
	}
	resp, err := s.doJSON(ctx, http.MethodPatch, "/api/v1/apps/services/"+url.PathEscape(ref), req, &data)
	if err != nil {
		return nil, "", resp, err
	}
	return data.Service, data.DeploymentID, resp, nil
}

// ListTemplates returns the ready-to-run app templates.
func (s *AppServiceServiceOp) ListTemplates(ctx context.Context) ([]AppTemplate, *Response, error) {
	var data struct {
		Templates []AppTemplate `json:"templates"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/templates", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Templates, resp, nil
}

// ListGitHubConnections returns the account's GitHub connections; the ID of
// one is CreateAppServiceRequest.GitHubInstallationID.
func (s *AppServiceServiceOp) ListGitHubConnections(ctx context.Context) ([]AppGitHubConnection, *Response, error) {
	var data struct {
		Installations []AppGitHubConnection `json:"installations"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/github/installations", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Installations, resp, nil
}

// ListBindings returns the databases linked to an app.
func (s *AppServiceServiceOp) ListBindings(ctx context.Context, ref string) ([]AppBinding, *Response, error) {
	var data struct {
		Bindings []AppBinding `json:"bindings"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref)+"/bindings", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Bindings, resp, nil
}

// CreateBinding links a database to an app and redeploys it with the
// DATABASE_* variables set.
func (s *AppServiceServiceOp) CreateBinding(ctx context.Context, ref string, req *CreateAppBindingRequest) (*AppBinding, *Response, error) {
	var data struct {
		Binding *AppBinding `json:"binding"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/services/"+url.PathEscape(ref)+"/bindings", req, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Binding, resp, nil
}

// DeleteBinding unlinks a database: the app's database user is removed and
// the DATABASE_* variables are cleared.
func (s *AppServiceServiceOp) DeleteBinding(ctx context.Context, bindingID string) (*Response, error) {
	return s.doJSON(ctx, http.MethodDelete, "/api/v1/apps/bindings/"+url.PathEscape(bindingID), nil, nil)
}

// RotateBinding gives the app new database credentials and redeploys it.
func (s *AppServiceServiceOp) RotateBinding(ctx context.Context, bindingID string) (*AppBinding, *Response, error) {
	var data struct {
		Binding *AppBinding `json:"binding"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/bindings/"+url.PathEscape(bindingID)+"/rotate", struct{}{}, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Binding, resp, nil
}

// ListPreviews returns an app's pull request previews.
func (s *AppServiceServiceOp) ListPreviews(ctx context.Context, ref string) ([]AppPreview, *Response, error) {
	var data struct {
		Previews []AppPreview `json:"previews"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref)+"/previews", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Previews, resp, nil
}

// DeletePreview removes a pull request preview.
func (s *AppServiceServiceOp) DeletePreview(ctx context.Context, previewID string) (*Response, error) {
	return s.doJSON(ctx, http.MethodDelete, "/api/v1/apps/previews/"+url.PathEscape(previewID), nil, nil)
}

// RedeployPreview builds and deploys the pull request's latest commit again.
func (s *AppServiceServiceOp) RedeployPreview(ctx context.Context, previewID string) (*AppPreview, string, *Response, error) {
	var data struct {
		Preview      *AppPreview `json:"preview"`
		DeploymentID string      `json:"deployment_id"`
	}
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/v1/apps/previews/"+url.PathEscape(previewID)+"/redeploy", struct{}{}, &data)
	if err != nil {
		return nil, "", resp, err
	}
	return data.Preview, data.DeploymentID, resp, nil
}

// ListEvents returns an app's activity timeline, newest first. A limit of 0
// uses the API default.
func (s *AppServiceServiceOp) ListEvents(ctx context.Context, ref string, limit int) ([]AppEvent, *Response, error) {
	var data struct {
		Events []AppEvent `json:"events"`
	}
	path := "/api/v1/apps/services/" + url.PathEscape(ref) + "/events"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	resp, err := s.doJSON(ctx, http.MethodGet, path, nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Events, resp, nil
}

// ListInstances returns an app's running instances.
func (s *AppServiceServiceOp) ListInstances(ctx context.Context, ref string) ([]AppInstance, *Response, error) {
	var data struct {
		Instances []AppInstance `json:"instances"`
	}
	resp, err := s.doJSON(ctx, http.MethodGet, "/api/v1/apps/services/"+url.PathEscape(ref)+"/instances", nil, &data)
	if err != nil {
		return nil, resp, err
	}
	return data.Instances, resp, nil
}
