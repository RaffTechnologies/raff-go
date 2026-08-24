package raff

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/rafftechnologies/raff-go/spec"
)

// parsePoolID converts a node pool ID string to the UUID path type.
func parsePoolID(poolID string) (openapi_types.UUID, error) {
	id, err := uuid.Parse(poolID)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid node pool ID %q: %w", poolID, err)
	}
	return id, nil
}

// K8sCluster represents a managed Kubernetes cluster.
type K8sCluster = spec.K8SCluster

// K8sClusterStatus is a cluster lifecycle status.
type K8sClusterStatus = spec.K8SClusterStatus

// Cluster lifecycle statuses.
const (
	K8sClusterStatusPending      = spec.K8SClusterStatusPending
	K8sClusterStatusDeploying    = spec.K8SClusterStatusDeploying
	K8sClusterStatusRunning      = spec.K8SClusterStatusRunning
	K8sClusterStatusWarning      = spec.K8SClusterStatusWarning
	K8sClusterStatusFailed       = spec.K8SClusterStatusFailed
	K8sClusterStatusDeleting     = spec.K8SClusterStatusDeleting
	K8sClusterStatusDeleteFailed = spec.K8SClusterStatusDeleteFailed
	K8sClusterStatusDeleted      = spec.K8SClusterStatusDeleted
)

// K8sNodePool represents a group of identical worker nodes.
type K8sNodePool = spec.K8SNodePool

// K8sClusterNode is one node of a cluster with live kubelet state.
type K8sClusterNode = spec.K8SClusterNode

// K8sVersion is an available Kubernetes version.
type K8sVersion = spec.K8SVersion

// K8sNodePlan is a worker node plan with pricing.
type K8sNodePlan = spec.K8SNodePlan

// K8sClusterEvent is one lifecycle event of a cluster.
type K8sClusterEvent = spec.K8SClusterEvent

// K8sNodePoolInput describes a node pool in a cluster create request.
type K8sNodePoolInput = spec.K8SNodePoolInput

// CreateK8sClusterStorageNodeCount is the storage-node count in a create request.
type CreateK8sClusterStorageNodeCount = spec.CreateK8SClusterRequestStorageNodeCount

// CreateK8sClusterRequest is the request body for creating a cluster.
type CreateK8sClusterRequest = spec.CreateK8SClusterJSONRequestBody

// AddK8sNodePoolRequest is the request body for adding a node pool.
type AddK8sNodePoolRequest = spec.AddK8SNodePoolJSONRequestBody

// UpdateK8sNodePoolRequest is the request body for updating a node pool.
type UpdateK8sNodePoolRequest = spec.UpdateK8SNodePoolJSONRequestBody

// K8sClusterListOptions are the query parameters for listing clusters.
type K8sClusterListOptions = spec.ListK8SClustersParams

// K8sKubeconfig is a cluster's kubeconfig with its API endpoint.
type K8sKubeconfig struct {
	Kubeconfig  string
	APIEndpoint string
}

// K8sNodePlans bundles worker plans with HA and storage pricing.
type K8sNodePlans struct {
	Plans          []K8sNodePlan
	HAPricing      *spec.K8SHAPricing
	StoragePricing *spec.K8SStoragePricing
}

// KubernetesService handles communication with the managed Kubernetes
// endpoints.
type KubernetesService interface {
	List(ctx context.Context, opts *K8sClusterListOptions) ([]K8sCluster, *Response, error)
	Get(ctx context.Context, clusterID string) (*K8sCluster, *Response, error)
	// Create provisions a cluster. idempotencyKey (optional, max 128 chars)
	// makes retries safe: the same key returns the same cluster instead of
	// creating a second one.
	Create(ctx context.Context, req *CreateK8sClusterRequest, idempotencyKey string) (*K8sCluster, *Response, error)
	Rename(ctx context.Context, clusterID, name string) (*Response, error)
	Delete(ctx context.Context, clusterID string) (*Response, error)
	Kubeconfig(ctx context.Context, clusterID string) (*K8sKubeconfig, *Response, error)
	UpgradeHA(ctx context.Context, clusterID string) (*Response, error)

	ListNodePools(ctx context.Context, clusterID string) ([]K8sNodePool, *Response, error)
	AddNodePool(ctx context.Context, clusterID string, req *AddK8sNodePoolRequest) (*K8sNodePool, *Response, error)
	UpdateNodePool(ctx context.Context, clusterID, poolID string, req *UpdateK8sNodePoolRequest) (*K8sNodePool, *Response, error)
	ScaleNodePool(ctx context.Context, clusterID, poolID string, nodeCount int) (*Response, error)
	DeleteNodePool(ctx context.Context, clusterID, poolID string) (*Response, error)

	ListNodes(ctx context.Context, clusterID string) ([]K8sClusterNode, *Response, error)
	ListEvents(ctx context.Context, clusterID string) ([]K8sClusterEvent, *Response, error)

	Upgrades(ctx context.Context, clusterID string) (*K8sUpgradeInfo, *Response, error)
	Upgrade(ctx context.Context, clusterID string, versionID int, confirmSingleMaster bool) (*Response, error)
	SetMaintenance(ctx context.Context, clusterID, mode string, day, startHour *int) (*Response, error)

	ListVersions(ctx context.Context) ([]K8sVersion, *Response, error)
	ListNodePlans(ctx context.Context) (*K8sNodePlans, *Response, error)

	// WaitForStatus polls the cluster until it reaches the target status,
	// the cluster enters failed/delete_failed, or ctx is done. Poll interval
	// is 10 seconds. Returns the cluster in its final observed state.
	WaitForStatus(ctx context.Context, clusterID string, target K8sClusterStatus) (*K8sCluster, error)
	// WaitForDeleted polls until Get returns 404 or ctx is done.
	WaitForDeleted(ctx context.Context, clusterID string) error
}

// KubernetesServiceOp implements KubernetesService.
type KubernetesServiceOp struct {
	client *Client
}

var _ KubernetesService = &KubernetesServiceOp{}

func (s *KubernetesServiceOp) List(ctx context.Context, opts *K8sClusterListOptions) ([]K8sCluster, *Response, error) {
	if opts == nil {
		opts = &K8sClusterListOptions{}
	}
	if opts.XProjectID == nil {
		if pid, err := s.client.optionalProjectID(); err == nil && pid != nil {
			opts.XProjectID = pid
		}
	}
	resp, err := s.client.spec.ListK8SClustersWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var clusters []K8sCluster
	if resp.JSON200.Clusters != nil {
		clusters = *resp.JSON200.Clusters
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return clusters, responseFrom(resp.HTTPResponse, total), nil
}

func (s *KubernetesServiceOp) Get(ctx context.Context, clusterID string) (*K8sCluster, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetK8SClusterWithResponse(ctx, clusterID, &spec.GetK8SClusterParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Cluster == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Cluster, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) Create(ctx context.Context, req *CreateK8sClusterRequest, idempotencyKey string) (*K8sCluster, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	params := &spec.CreateK8SClusterParams{XProjectID: projectID}
	if idempotencyKey != "" {
		params.IdempotencyKey = &idempotencyKey
	}
	resp, err := s.client.spec.CreateK8SClusterWithResponse(ctx, params, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Cluster == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Cluster, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) Rename(ctx context.Context, clusterID, name string) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RenameK8SClusterWithResponse(ctx, clusterID,
		&spec.RenameK8SClusterParams{XProjectID: projectID},
		spec.RenameK8SClusterJSONRequestBody{Name: name})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) Delete(ctx context.Context, clusterID string) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteK8SClusterWithResponse(ctx, clusterID, &spec.DeleteK8SClusterParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) Kubeconfig(ctx context.Context, clusterID string) (*K8sKubeconfig, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetK8SKubeconfigWithResponse(ctx, clusterID, &spec.GetK8SKubeconfigParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Kubeconfig == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	kc := &K8sKubeconfig{Kubeconfig: *resp.JSON200.Kubeconfig}
	if resp.JSON200.APIEndpoint != nil {
		kc.APIEndpoint = *resp.JSON200.APIEndpoint
	}
	return kc, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) UpgradeHA(ctx context.Context, clusterID string) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.UpgradeK8SClusterHAWithResponse(ctx, clusterID, &spec.UpgradeK8SClusterHAParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ListNodePools(ctx context.Context, clusterID string) ([]K8sNodePool, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListK8SNodePoolsWithResponse(ctx, clusterID, &spec.ListK8SNodePoolsParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var pools []K8sNodePool
	if resp.JSON200.NodePools != nil {
		pools = *resp.JSON200.NodePools
	}
	return pools, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) AddNodePool(ctx context.Context, clusterID string, req *AddK8sNodePoolRequest) (*K8sNodePool, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AddK8SNodePoolWithResponse(ctx, clusterID, &spec.AddK8SNodePoolParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.NodePool == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.NodePool, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) UpdateNodePool(ctx context.Context, clusterID, poolID string, req *UpdateK8sNodePoolRequest) (*K8sNodePool, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	pid, err := parsePoolID(poolID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateK8SNodePoolWithResponse(ctx, clusterID, pid, &spec.UpdateK8SNodePoolParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.NodePool == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.NodePool, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ScaleNodePool(ctx context.Context, clusterID, poolID string, nodeCount int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	pid, err := parsePoolID(poolID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.ScaleK8SNodePoolWithResponse(ctx, clusterID, pid,
		&spec.ScaleK8SNodePoolParams{XProjectID: projectID},
		spec.ScaleK8SNodePoolJSONRequestBody{NodeCount: nodeCount})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) DeleteNodePool(ctx context.Context, clusterID, poolID string) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	pid, err := parsePoolID(poolID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteK8SNodePoolWithResponse(ctx, clusterID, pid, &spec.DeleteK8SNodePoolParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ListNodes(ctx context.Context, clusterID string) ([]K8sClusterNode, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListK8SClusterNodesWithResponse(ctx, clusterID, &spec.ListK8SClusterNodesParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var nodes []K8sClusterNode
	if resp.JSON200.Nodes != nil {
		nodes = *resp.JSON200.Nodes
	}
	return nodes, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ListEvents(ctx context.Context, clusterID string) ([]K8sClusterEvent, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListK8SClusterEventsWithResponse(ctx, clusterID, &spec.ListK8SClusterEventsParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var events []K8sClusterEvent
	if resp.JSON200.Events != nil {
		events = *resp.JSON200.Events
	}
	return events, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ListVersions(ctx context.Context) ([]K8sVersion, *Response, error) {
	resp, err := s.client.spec.ListK8SVersionsWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var versions []K8sVersion
	if resp.JSON200.Versions != nil {
		versions = *resp.JSON200.Versions
	}
	return versions, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *KubernetesServiceOp) ListNodePlans(ctx context.Context) (*K8sNodePlans, *Response, error) {
	resp, err := s.client.spec.ListK8SNodePlansWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	out := &K8sNodePlans{
		HAPricing:      resp.JSON200.HaPricing,
		StoragePricing: resp.JSON200.StoragePricing,
	}
	if resp.JSON200.Plans != nil {
		out.Plans = *resp.JSON200.Plans
	}
	return out, responseFrom(resp.HTTPResponse, 0), nil
}

// k8sWaitPollInterval is how often the waiters poll.
const k8sWaitPollInterval = 10 * time.Second

func (s *KubernetesServiceOp) WaitForStatus(ctx context.Context, clusterID string, target K8sClusterStatus) (*K8sCluster, error) {
	t := time.NewTicker(k8sWaitPollInterval)
	defer t.Stop()
	for {
		cluster, _, err := s.Get(ctx, clusterID)
		if err == nil {
			switch cluster.Status {
			case target:
				return cluster, nil
			case K8sClusterStatusFailed, K8sClusterStatusDeleteFailed:
				msg := ""
				if cluster.StatusMessage != nil {
					msg = ": " + *cluster.StatusMessage
				}
				return cluster, fmt.Errorf("cluster %s entered status %s%s", clusterID, cluster.Status, msg)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
}

func (s *KubernetesServiceOp) WaitForDeleted(ctx context.Context, clusterID string) error {
	t := time.NewTicker(k8sWaitPollInterval)
	defer t.Stop()
	for {
		_, resp, err := s.Get(ctx, clusterID)
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

// K8sAvailableUpgrade is one version a cluster can upgrade to.
type K8sAvailableUpgrade struct {
	VersionID   int
	Version     string
	RKE2Version string
	IsMinor     bool
	IsDefault   bool
}

// K8sUpgradeInfo is a cluster's upgrade and maintenance state.
type K8sUpgradeInfo struct {
	CurrentVersion   string
	UpgradeStatus    string
	TargetVersion    string
	UpgradeMode      string
	MaintenanceDay   *int
	MaintenanceStart *int
	Available        []K8sAvailableUpgrade
}

// Upgrades lists the versions a cluster can upgrade to plus its upgrade state.
func (s *KubernetesServiceOp) Upgrades(ctx context.Context, clusterID string) (*K8sUpgradeInfo, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListK8SClusterUpgradesWithResponse(ctx, clusterID, &spec.ListK8SClusterUpgradesParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	info := &K8sUpgradeInfo{
		MaintenanceDay:   resp.JSON200.MaintenanceDay,
		MaintenanceStart: resp.JSON200.MaintenanceStart,
	}
	if resp.JSON200.CurrentVersion != nil {
		info.CurrentVersion = *resp.JSON200.CurrentVersion
	}
	if resp.JSON200.UpgradeStatus != nil {
		info.UpgradeStatus = string(*resp.JSON200.UpgradeStatus)
	}
	if resp.JSON200.TargetVersion != nil {
		info.TargetVersion = *resp.JSON200.TargetVersion
	}
	if resp.JSON200.UpgradeMode != nil {
		info.UpgradeMode = string(*resp.JSON200.UpgradeMode)
	}
	if resp.JSON200.Available != nil {
		for _, a := range *resp.JSON200.Available {
			u := K8sAvailableUpgrade{}
			if a.VersionID != nil {
				u.VersionID = *a.VersionID
			}
			if a.Version != nil {
				u.Version = *a.Version
			}
			if a.Rke2Version != nil {
				u.RKE2Version = *a.Rke2Version
			}
			if a.IsMinor != nil {
				u.IsMinor = *a.IsMinor
			}
			if a.IsDefault != nil {
				u.IsDefault = *a.IsDefault
			}
			info.Available = append(info.Available, u)
		}
	}
	return info, responseFrom(resp.HTTPResponse, 0), nil
}

// Upgrade starts an in-place Kubernetes version upgrade. confirmSingleMaster
// must be true on non-HA clusters (brief API interruption).
func (s *KubernetesServiceOp) Upgrade(ctx context.Context, clusterID string, versionID int, confirmSingleMaster bool) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	body := spec.UpgradeK8SClusterVersionJSONRequestBody{VersionID: versionID}
	if confirmSingleMaster {
		body.ConfirmSingleMaster = &confirmSingleMaster
	}
	resp, err := s.client.spec.UpgradeK8SClusterVersionWithResponse(ctx, clusterID,
		&spec.UpgradeK8SClusterVersionParams{XProjectID: projectID}, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

// SetMaintenance sets the upgrade mode (manual, auto_patch, auto_minor) and
// optionally the weekly 4-hour maintenance window.
func (s *KubernetesServiceOp) SetMaintenance(ctx context.Context, clusterID, mode string, day, startHour *int) (*Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	body := spec.UpdateK8SClusterMaintenanceJSONRequestBody{
		UpgradeMode:      spec.UpdateK8SClusterMaintenanceJSONBodyUpgradeMode(mode),
		MaintenanceDay:   day,
		MaintenanceStart: startHour,
	}
	resp, err := s.client.spec.UpdateK8SClusterMaintenanceWithResponse(ctx, clusterID,
		&spec.UpdateK8SClusterMaintenanceParams{XProjectID: projectID}, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
