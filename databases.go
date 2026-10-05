package raff

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rafftechnologies/raff-go/spec"
)

// Database is a managed database instance.
type Database = spec.Database

// DatabaseStatus is a database lifecycle status.
type DatabaseStatus = spec.DatabaseStatus

// Database lifecycle statuses.
const (
	DatabaseStatusPending   = spec.DatabaseStatusPending
	DatabaseStatusDeploying = spec.DatabaseStatusDeploying
	DatabaseStatusRunning   = spec.DatabaseStatusRunning
	DatabaseStatusWarning   = spec.DatabaseStatusWarning
	DatabaseStatusResizing  = spec.DatabaseStatusResizing
	DatabaseStatusRestoring = spec.DatabaseStatusRestoring
	DatabaseStatusSuspended = spec.DatabaseStatusSuspended
	DatabaseStatusFailed    = spec.DatabaseStatusFailed
	DatabaseStatusDeleting  = spec.DatabaseStatusDeleting
	DatabaseStatusDeleted   = spec.DatabaseStatusDeleted
)

// DatabaseEngine is a database engine identifier (postgres, mysql, valkey,
// clickhouse, kafka).
type DatabaseEngine = spec.DatabaseEngine

// DatabaseEngineInfo is one engine of the catalog with its versions.
type DatabaseEngineInfo = spec.DatabaseEngineInfo

// DatabasePricingPlan is a database plan with its price and size.
type DatabasePricingPlan = spec.DatabasePricingPlan

// DatabaseListOptions are the parameters for listing databases.
type DatabaseListOptions = spec.ListDatabasesParams

// CreateDatabaseRequest is the request body for creating a database.
type CreateDatabaseRequest = spec.CreateDatabaseJSONRequestBody

// ScaleDatabaseRequest is the request body for scaling a database.
type ScaleDatabaseRequest = spec.ScaleDatabaseJSONRequestBody

// RestoreDatabaseRequest is the request body for a restore.
type RestoreDatabaseRequest = spec.RestoreDatabaseJSONRequestBody

// DatabaseConnection holds the connection details of a database.
type DatabaseConnection = spec.DatabaseConnection

// DatabaseMetrics is the latest usage snapshot of a database.
type DatabaseMetrics = spec.DatabaseMetrics

// DatabaseMetricsPoint is one bucketed point of the metrics history.
type DatabaseMetricsPoint = spec.DatabaseMetricsPoint

// DatabaseClientTraffic is the traffic of one database user over a window.
type DatabaseClientTraffic = spec.DatabaseClientTraffic

// DatabaseSlowQuery is one slow-query entry.
type DatabaseSlowQuery = spec.DatabaseSlowQuery

// DatabaseUser is a user (role) inside the database engine.
type DatabaseUser = spec.DatabaseUser

// DatabaseUserRole is the access a new database user gets.
type DatabaseUserRole = spec.CreateDatabaseUserRequestRole

// Database user roles.
const (
	DatabaseUserReadOnly  = spec.Readonly
	DatabaseUserReadWrite = spec.Readwrite
)

// DatabaseQueryRequest is the request body for running a query.
type DatabaseQueryRequest = spec.RunDatabaseQueryJSONRequestBody

// DatabaseQueryResult is the result of a query.
type DatabaseQueryResult = spec.DatabaseQueryResult

// DatabaseBrowseRequest is the request body for browsing a database.
type DatabaseBrowseRequest = spec.BrowseDatabaseJSONRequestBody

// DatabaseBrowseResult is one level of a database: a list of nodes or a page
// of rows.
type DatabaseBrowseResult = spec.DatabaseBrowseResult

// DatabaseRow is one result row: values as text plus a NULL marker per value.
type DatabaseRow = spec.DatabaseRow

// DatabaseSavedQuery is a named query saved for a database.
type DatabaseSavedQuery = spec.DatabaseSavedQuery

// DatabaseExtension is a PostgreSQL extension with its installed state.
type DatabaseExtension = spec.DatabaseExtension

// DatabaseParameter is an engine setting in effect.
type DatabaseParameter = spec.DatabaseParameter

// DatabaseConsumer is an app or function bound to a database.
type DatabaseConsumer = spec.DatabaseConsumer

// DatabaseBackup is a database backup.
type DatabaseBackup = spec.DatabaseBackup

// DatabasePlans bundles database plans with HA, replica and storage pricing.
type DatabasePlans struct {
	Plans          []DatabasePricingPlan
	HAPricing      []spec.DatabaseHAPricing
	ReplicaPricing []spec.DatabaseReplicaPricing
	StoragePricing []spec.DatabaseStoragePricing
}

// DatabaseMetricsHistory is a metrics time series.
type DatabaseMetricsHistory struct {
	Points        []DatabaseMetricsPoint
	BucketSeconds int
}

// DatabaseClientTrafficHistory is the traffic per database user over a window.
type DatabaseClientTrafficHistory struct {
	Clients       []DatabaseClientTraffic
	BucketSeconds int
}

// DatabaseSlowQueries lists the slowest recent queries.
type DatabaseSlowQueries struct {
	Queries []DatabaseSlowQuery
	// Collecting is false until slow-query statistics are being collected.
	Collecting bool
}

// DatabaseCredential is a database user's username and password.
type DatabaseCredential struct {
	Username string
	Password string
}

// DatabaseService handles communication with the managed database endpoints.
type DatabaseService interface {
	ListEngines(ctx context.Context) ([]DatabaseEngineInfo, *Response, error)
	// ListPlans returns the plans; engine filters them ("" for all).
	ListPlans(ctx context.Context, engine DatabaseEngine) (*DatabasePlans, *Response, error)

	List(ctx context.Context, opts *DatabaseListOptions) ([]Database, *Response, error)
	Get(ctx context.Context, databaseID string) (*Database, *Response, error)
	Create(ctx context.Context, req *CreateDatabaseRequest) (*Database, *Response, error)
	Rename(ctx context.Context, databaseID, name string) (*Database, *Response, error)
	Delete(ctx context.Context, databaseID string) (*Response, error)
	Scale(ctx context.Context, databaseID string, req *ScaleDatabaseRequest) (*Database, *Response, error)
	// Resume wakes a free-tier database paused for idleness.
	Resume(ctx context.Context, databaseID string) (*Database, *Response, error)

	// Connection returns the connection details; reveal includes the password.
	Connection(ctx context.Context, databaseID string, reveal bool) (*DatabaseConnection, *Response, error)
	RotateCredentials(ctx context.Context, databaseID string) (*Response, error)
	ConnectVPC(ctx context.Context, databaseID, vpcID string) (*Database, *Response, error)
	DisconnectVPC(ctx context.Context, databaseID string) (*Database, *Response, error)
	// SetPublicAccess opens or closes public access. allowlist nil keeps the
	// stored list; an empty slice allows all sources.
	SetPublicAccess(ctx context.Context, databaseID string, enabled bool, allowlist []string) (*Database, *Response, error)

	ListUsers(ctx context.Context, databaseID string) ([]DatabaseUser, *Response, error)
	// CreateUser adds a user and returns it with its password.
	CreateUser(ctx context.Context, databaseID, name string, role DatabaseUserRole) (*DatabaseUser, string, *Response, error)
	// RotateUser gives a user a new password and returns it.
	RotateUser(ctx context.Context, databaseID, name string) (string, *Response, error)
	UserCredential(ctx context.Context, databaseID, name string) (*DatabaseCredential, *Response, error)
	DeleteUser(ctx context.Context, databaseID, name string) (*Response, error)

	Query(ctx context.Context, databaseID string, req *DatabaseQueryRequest) (*DatabaseQueryResult, *Response, error)
	Browse(ctx context.Context, databaseID string, req *DatabaseBrowseRequest) (*DatabaseBrowseResult, *Response, error)
	ListSavedQueries(ctx context.Context, databaseID string) ([]DatabaseSavedQuery, *Response, error)
	SaveQuery(ctx context.Context, databaseID, name, statement string) (*DatabaseSavedQuery, *Response, error)
	DeleteSavedQuery(ctx context.Context, databaseID, name string) (*Response, error)

	Metrics(ctx context.Context, databaseID string) (*DatabaseMetrics, *Response, error)
	MetricsHistory(ctx context.Context, databaseID string, windowHours int) (*DatabaseMetricsHistory, *Response, error)
	ClientTraffic(ctx context.Context, databaseID string, windowHours int) (*DatabaseClientTrafficHistory, *Response, error)
	SlowQueries(ctx context.Context, databaseID string, limit int) (*DatabaseSlowQueries, *Response, error)
	Logs(ctx context.Context, databaseID string, tail int) ([]string, *Response, error)

	ListExtensions(ctx context.Context, databaseID string) ([]DatabaseExtension, *Response, error)
	SetExtension(ctx context.Context, databaseID, name string, enabled bool) (*Response, error)
	ListParameters(ctx context.Context, databaseID string) ([]DatabaseParameter, *Response, error)
	ListConsumers(ctx context.Context, databaseID string) ([]DatabaseConsumer, *Response, error)

	ListBackups(ctx context.Context, databaseID string) ([]DatabaseBackup, *Response, error)
	CreateBackup(ctx context.Context, databaseID string) (*DatabaseBackup, *Response, error)
	// Restore starts a restore and returns the new database's short id (empty
	// for an in-place restore).
	Restore(ctx context.Context, databaseID string, req *RestoreDatabaseRequest) (string, *Response, error)

	// WaitForStatus polls the database until it reaches the target status,
	// enters failed, or ctx is done. Poll interval is 2 seconds. Returns the
	// database in its final observed state.
	WaitForStatus(ctx context.Context, databaseID string, target DatabaseStatus) (*Database, error)
	// WaitForDeleted polls until Get returns 404 or ctx is done.
	WaitForDeleted(ctx context.Context, databaseID string) error
}

// DatabaseServiceOp implements DatabaseService.
type DatabaseServiceOp struct {
	client *Client
}

var _ DatabaseService = &DatabaseServiceOp{}

const databaseWaitPollInterval = 2 * time.Second

// databaseResult returns the database of a {success, database} response, or
// the API error when the body has none.
func databaseResult(d *Database, status int, httpResp *Response, body []byte) (*Database, *Response, error) {
	if d == nil {
		return nil, httpResp, errorFromResponse(status, body)
	}
	return d, httpResp, nil
}

// statusOnly turns a response without a useful body into (*Response, error).
func statusOnly(status int, httpResp *Response, body []byte) (*Response, error) {
	if status >= 400 {
		return httpResp, errorFromResponse(status, body)
	}
	return httpResp, nil
}

// decodeFlat decodes a 200 body whose fields sit at the top level next to
// "success" (the spec's allOf responses).
func decodeFlat[T any](status int, httpResp *Response, body []byte) (*T, *Response, error) {
	if status != 200 {
		return nil, httpResp, errorFromResponse(status, body)
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, httpResp, fmt.Errorf("decode response: %w", err)
	}
	return &out, httpResp, nil
}

func deref[T any](p *[]T) []T {
	if p == nil {
		return nil
	}
	return *p
}

func (s *DatabaseServiceOp) ListEngines(ctx context.Context) ([]DatabaseEngineInfo, *Response, error) {
	resp, err := s.client.spec.ListDatabaseEnginesWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Engines), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) ListPlans(ctx context.Context, engine DatabaseEngine) (*DatabasePlans, *Response, error) {
	params := &spec.ListDatabasePlansParams{}
	if engine != "" {
		params.Engine = &engine
	}
	resp, err := s.client.spec.ListDatabasePlansWithResponse(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return &DatabasePlans{
		Plans:          deref(resp.JSON200.Plans),
		HAPricing:      deref(resp.JSON200.HaPricing),
		ReplicaPricing: deref(resp.JSON200.ReplicaPricing),
		StoragePricing: deref(resp.JSON200.StoragePricing),
	}, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) List(ctx context.Context, opts *DatabaseListOptions) ([]Database, *Response, error) {
	if opts == nil {
		opts = &DatabaseListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListDatabasesWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	dbs := deref(resp.JSON200.Databases)
	return dbs, responseFrom(resp.HTTPResponse, len(dbs)), nil
}

func (s *DatabaseServiceOp) Get(ctx context.Context, databaseID string) (*Database, *Response, error) {
	resp, err := s.client.spec.GetDatabaseWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Create(ctx context.Context, req *CreateDatabaseRequest) (*Database, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateDatabaseWithResponse(ctx, &spec.CreateDatabaseParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON201 != nil {
		d = resp.JSON201.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Rename(ctx context.Context, databaseID, name string) (*Database, *Response, error) {
	resp, err := s.client.spec.UpdateDatabaseWithResponse(ctx, databaseID, spec.UpdateDatabaseJSONRequestBody{Name: name})
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Delete(ctx context.Context, databaseID string) (*Response, error) {
	resp, err := s.client.spec.DeleteDatabaseWithResponse(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	return statusOnly(resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Scale(ctx context.Context, databaseID string, req *ScaleDatabaseRequest) (*Database, *Response, error) {
	resp, err := s.client.spec.ScaleDatabaseWithResponse(ctx, databaseID, *req)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Resume(ctx context.Context, databaseID string) (*Database, *Response, error) {
	resp, err := s.client.spec.ResumeDatabaseWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Connection(ctx context.Context, databaseID string, reveal bool) (*DatabaseConnection, *Response, error) {
	resp, err := s.client.spec.GetDatabaseConnectionWithResponse(ctx, databaseID, &spec.GetDatabaseConnectionParams{Reveal: &reveal})
	if err != nil {
		return nil, nil, err
	}
	return decodeFlat[DatabaseConnection](resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) RotateCredentials(ctx context.Context, databaseID string) (*Response, error) {
	resp, err := s.client.spec.RotateDatabaseCredentialsWithResponse(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	return statusOnly(resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) ConnectVPC(ctx context.Context, databaseID, vpcID string) (*Database, *Response, error) {
	id, err := parseUUID(vpcID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ConnectDatabaseVPCWithResponse(ctx, databaseID, spec.ConnectDatabaseVPCJSONRequestBody{VpcID: id})
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) DisconnectVPC(ctx context.Context, databaseID string) (*Database, *Response, error) {
	resp, err := s.client.spec.DisconnectDatabaseVPCWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) SetPublicAccess(ctx context.Context, databaseID string, enabled bool, allowlist []string) (*Database, *Response, error) {
	body := spec.SetDatabasePublicAccessJSONRequestBody{Enabled: enabled}
	if allowlist != nil {
		body.Allowlist = &allowlist
	}
	resp, err := s.client.spec.SetDatabasePublicAccessWithResponse(ctx, databaseID, body)
	if err != nil {
		return nil, nil, err
	}
	var d *Database
	if resp.JSON200 != nil {
		d = resp.JSON200.Database
	}
	return databaseResult(d, resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) ListUsers(ctx context.Context, databaseID string) ([]DatabaseUser, *Response, error) {
	resp, err := s.client.spec.ListDatabaseUsersWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Users), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) CreateUser(ctx context.Context, databaseID, name string, role DatabaseUserRole) (*DatabaseUser, string, *Response, error) {
	resp, err := s.client.spec.CreateDatabaseUserWithResponse(ctx, databaseID, spec.CreateDatabaseUserJSONRequestBody{Name: name, Role: role})
	if err != nil {
		return nil, "", nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.User == nil {
		return nil, "", responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	password := ""
	if resp.JSON201.Password != nil {
		password = *resp.JSON201.Password
	}
	return resp.JSON201.User, password, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) RotateUser(ctx context.Context, databaseID, name string) (string, *Response, error) {
	resp, err := s.client.spec.RotateDatabaseUserWithResponse(ctx, databaseID, name)
	if err != nil {
		return "", nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Password == nil {
		return "", responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return *resp.JSON200.Password, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) UserCredential(ctx context.Context, databaseID, name string) (*DatabaseCredential, *Response, error) {
	resp, err := s.client.spec.GetDatabaseUserCredentialWithResponse(ctx, databaseID, name)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	cred := &DatabaseCredential{}
	if resp.JSON200.Username != nil {
		cred.Username = *resp.JSON200.Username
	}
	if resp.JSON200.Password != nil {
		cred.Password = *resp.JSON200.Password
	}
	return cred, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) DeleteUser(ctx context.Context, databaseID, name string) (*Response, error) {
	resp, err := s.client.spec.DeleteDatabaseUserWithResponse(ctx, databaseID, name)
	if err != nil {
		return nil, err
	}
	return statusOnly(resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Query(ctx context.Context, databaseID string, req *DatabaseQueryRequest) (*DatabaseQueryResult, *Response, error) {
	resp, err := s.client.spec.RunDatabaseQueryWithResponse(ctx, databaseID, *req)
	if err != nil {
		return nil, nil, err
	}
	return decodeFlat[DatabaseQueryResult](resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Browse(ctx context.Context, databaseID string, req *DatabaseBrowseRequest) (*DatabaseBrowseResult, *Response, error) {
	resp, err := s.client.spec.BrowseDatabaseWithResponse(ctx, databaseID, *req)
	if err != nil {
		return nil, nil, err
	}
	return decodeFlat[DatabaseBrowseResult](resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) ListSavedQueries(ctx context.Context, databaseID string) ([]DatabaseSavedQuery, *Response, error) {
	resp, err := s.client.spec.ListDatabaseSavedQueriesWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Queries), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) SaveQuery(ctx context.Context, databaseID, name, statement string) (*DatabaseSavedQuery, *Response, error) {
	resp, err := s.client.spec.SaveDatabaseSavedQueryWithResponse(ctx, databaseID, spec.SaveDatabaseSavedQueryJSONRequestBody{Name: name, Statement: statement})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Query == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Query, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) DeleteSavedQuery(ctx context.Context, databaseID, name string) (*Response, error) {
	resp, err := s.client.spec.DeleteDatabaseSavedQueryWithResponse(ctx, databaseID, name)
	if err != nil {
		return nil, err
	}
	return statusOnly(resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) Metrics(ctx context.Context, databaseID string) (*DatabaseMetrics, *Response, error) {
	resp, err := s.client.spec.GetDatabaseMetricsWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	return decodeFlat[DatabaseMetrics](resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) MetricsHistory(ctx context.Context, databaseID string, windowHours int) (*DatabaseMetricsHistory, *Response, error) {
	params := &spec.GetDatabaseMetricsHistoryParams{}
	if windowHours > 0 {
		params.WindowHours = &windowHours
	}
	resp, err := s.client.spec.GetDatabaseMetricsHistoryWithResponse(ctx, databaseID, params)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	h := &DatabaseMetricsHistory{Points: deref(resp.JSON200.Points)}
	if resp.JSON200.BucketSeconds != nil {
		h.BucketSeconds = *resp.JSON200.BucketSeconds
	}
	return h, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) ClientTraffic(ctx context.Context, databaseID string, windowHours int) (*DatabaseClientTrafficHistory, *Response, error) {
	params := &spec.GetDatabaseClientTrafficParams{}
	if windowHours > 0 {
		params.WindowHours = &windowHours
	}
	resp, err := s.client.spec.GetDatabaseClientTrafficWithResponse(ctx, databaseID, params)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	h := &DatabaseClientTrafficHistory{Clients: deref(resp.JSON200.Clients)}
	if resp.JSON200.BucketSeconds != nil {
		h.BucketSeconds = *resp.JSON200.BucketSeconds
	}
	return h, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) SlowQueries(ctx context.Context, databaseID string, limit int) (*DatabaseSlowQueries, *Response, error) {
	params := &spec.GetDatabaseSlowQueriesParams{}
	if limit > 0 {
		params.Limit = &limit
	}
	resp, err := s.client.spec.GetDatabaseSlowQueriesWithResponse(ctx, databaseID, params)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	out := &DatabaseSlowQueries{Queries: deref(resp.JSON200.Queries)}
	if resp.JSON200.Collecting != nil {
		out.Collecting = *resp.JSON200.Collecting
	}
	return out, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) Logs(ctx context.Context, databaseID string, tail int) ([]string, *Response, error) {
	params := &spec.GetDatabaseLogsParams{}
	if tail > 0 {
		params.Tail = &tail
	}
	resp, err := s.client.spec.GetDatabaseLogsWithResponse(ctx, databaseID, params)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Lines), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) ListExtensions(ctx context.Context, databaseID string) ([]DatabaseExtension, *Response, error) {
	resp, err := s.client.spec.ListDatabaseExtensionsWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Extensions), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) SetExtension(ctx context.Context, databaseID, name string, enabled bool) (*Response, error) {
	resp, err := s.client.spec.SetDatabaseExtensionWithResponse(ctx, databaseID, spec.SetDatabaseExtensionJSONRequestBody{Name: name, Enabled: enabled})
	if err != nil {
		return nil, err
	}
	return statusOnly(resp.StatusCode(), responseFrom(resp.HTTPResponse, 0), resp.Body)
}

func (s *DatabaseServiceOp) ListParameters(ctx context.Context, databaseID string) ([]DatabaseParameter, *Response, error) {
	resp, err := s.client.spec.ListDatabaseParametersWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Parameters), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) ListConsumers(ctx context.Context, databaseID string) ([]DatabaseConsumer, *Response, error) {
	resp, err := s.client.spec.ListDatabaseConsumersWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Consumers), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) ListBackups(ctx context.Context, databaseID string) ([]DatabaseBackup, *Response, error) {
	resp, err := s.client.spec.ListDatabaseBackupsWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return deref(resp.JSON200.Backups), responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) CreateBackup(ctx context.Context, databaseID string) (*DatabaseBackup, *Response, error) {
	resp, err := s.client.spec.CreateDatabaseBackupWithResponse(ctx, databaseID)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Backup == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Backup, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) Restore(ctx context.Context, databaseID string, req *RestoreDatabaseRequest) (string, *Response, error) {
	resp, err := s.client.spec.RestoreDatabaseWithResponse(ctx, databaseID, *req)
	if err != nil {
		return "", nil, err
	}
	if resp.JSON200 == nil {
		return "", responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	newID := ""
	if resp.JSON200.NewDatabaseID != nil {
		newID = *resp.JSON200.NewDatabaseID
	}
	return newID, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *DatabaseServiceOp) WaitForStatus(ctx context.Context, databaseID string, target DatabaseStatus) (*Database, error) {
	t := time.NewTicker(databaseWaitPollInterval)
	defer t.Stop()
	for {
		db, _, err := s.Get(ctx, databaseID)
		if err == nil && db.Status != nil {
			switch *db.Status {
			case target:
				return db, nil
			case DatabaseStatusFailed:
				msg := ""
				if db.StatusMessage != nil && *db.StatusMessage != "" {
					msg = ": " + *db.StatusMessage
				}
				return db, fmt.Errorf("database %s entered status %s%s", databaseID, *db.Status, msg)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
}

func (s *DatabaseServiceOp) WaitForDeleted(ctx context.Context, databaseID string) error {
	t := time.NewTicker(databaseWaitPollInterval)
	defer t.Stop()
	for {
		_, resp, err := s.Get(ctx, databaseID)
		if err != nil && resp != nil && resp.StatusCode == 404 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}
