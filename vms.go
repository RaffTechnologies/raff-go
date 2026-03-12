package raff

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const vmsBasePath = "/api/v1/vms"

// VMService handles communication with the VM related methods of the Raff API.
type VMService interface {
	List(ctx context.Context, opts *VMListOptions) ([]VM, *Response, error)
	Get(ctx context.Context, vmID string) (*VM, *Response, error)
	Create(ctx context.Context, req *VMCreateRequest) (*VM, *Response, error)
	Delete(ctx context.Context, vmID string) (*Response, error)
	Start(ctx context.Context, vmID string) (*Response, error)
	Stop(ctx context.Context, vmID string) (*Response, error)
	Reboot(ctx context.Context, vmID string) (*Response, error)
	Resize(ctx context.Context, vmID string, req *VMResizeRequest) (*Response, error)
}

// VMServiceOp implements VMService.
type VMServiceOp struct {
	client *Client
}

var _ VMService = &VMServiceOp{}

// VM represents a virtual machine (public response fields).
type VM struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	CPU                int       `json:"cpu"`
	RAM                int       `json:"ram"`
	Storage            int       `json:"storage"`
	AddedStorage       int       `json:"added_storage"`
	TotalStorage       int       `json:"total_storage"`
	TemplateID         string    `json:"template_id"`
	TemplateName       string    `json:"template_name"`
	TemplateVersion    string    `json:"template_version"`
	VMID               int       `json:"vm_id"`
	Version            string    `json:"version,omitempty"`
	PricePerHour       string    `json:"price_per_hour"`
	PricingID          int       `json:"pricing_id"`
	CreatedBy          string    `json:"created_by,omitempty"`
	BillingType        string    `json:"billing_type,omitempty"`
	SubscriptionID     string    `json:"subscription_id,omitempty"`
	BackupType         *string   `json:"backup_type,omitempty"`
	Region             string    `json:"region"`
	ProjectID          string    `json:"project_id,omitempty"`
	ReservedIP         *string   `json:"reserved_ip,omitempty"`
	PublicIPv4Address  string    `json:"public_ipv4_address,omitempty"`
	PublicIPv6Address  *string   `json:"public_ipv6_address,omitempty"`
	PrivateIPv4Address string    `json:"private_ipv4_address,omitempty"`
	PrivateIPv6Address *string   `json:"private_ipv6_address,omitempty"`
	Tags               []string  `json:"tags,omitempty"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// VMCreateRequest represents a request to create a VM.
type VMCreateRequest struct {
	Name             string   `json:"name"`
	TemplateID       string   `json:"template_id"`
	PricingID        int      `json:"pricing_id"`
	Region           string   `json:"region"`
	ProjectID        string   `json:"project_id,omitempty"`
	SSHKeys          []string `json:"ssh_keys,omitempty"`
	Password         string   `json:"password,omitempty"`
	ExtraStorage     int      `json:"extra_storage,omitempty"`
	ExtraStorageType string   `json:"extra_storage_type,omitempty"`
	BackupType       string   `json:"backup_type,omitempty"`
	BackupTime       string   `json:"backup_time,omitempty"`
	BackupDate       string   `json:"backup_date,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	VPCID            string   `json:"vpc_id,omitempty"`
	VPCName          string   `json:"vpc_name,omitempty"`
	VPCCIDR          string   `json:"vpc_cidr,omitempty"`
}

// VMResizeRequest represents a request to resize a VM.
type VMResizeRequest struct {
	CPU int `json:"cpu"`
	RAM int `json:"ram"`
}

// VMListOptions specifies the optional parameters to the VM List method.
type VMListOptions struct {
	ProjectID string
	Region    string
	Status    string
	Limit     int
	Offset    int
}

// List returns all VMs for the authenticated account.
func (s *VMServiceOp) List(ctx context.Context, opts *VMListOptions) ([]VM, *Response, error) {
	path := vmsBasePath
	if opts != nil {
		params := url.Values{}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.Region != "" {
			params.Set("region", opts.Region)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var vms []VM
	resp, err := s.client.Do(ctx, req, &vms)
	if err != nil {
		return nil, resp, err
	}

	return vms, resp, nil
}

// Get returns a single VM by ID.
func (s *VMServiceOp) Get(ctx context.Context, vmID string) (*VM, *Response, error) {
	path := fmt.Sprintf("%s/%s", vmsBasePath, vmID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var vm VM
	resp, err := s.client.Do(ctx, req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return &vm, resp, nil
}

// Create creates a new VM.
func (s *VMServiceOp) Create(ctx context.Context, createReq *VMCreateRequest) (*VM, *Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, vmsBasePath, createReq)
	if err != nil {
		return nil, nil, err
	}

	var vm VM
	resp, err := s.client.Do(ctx, req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return &vm, resp, nil
}

// Delete deletes a VM.
func (s *VMServiceOp) Delete(ctx context.Context, vmID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", vmsBasePath, vmID)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Start starts a stopped VM.
func (s *VMServiceOp) Start(ctx context.Context, vmID string) (*Response, error) {
	return s.vmAction(ctx, vmID, "start")
}

// Stop stops a running VM.
func (s *VMServiceOp) Stop(ctx context.Context, vmID string) (*Response, error) {
	return s.vmAction(ctx, vmID, "stop")
}

// Reboot reboots a running VM.
func (s *VMServiceOp) Reboot(ctx context.Context, vmID string) (*Response, error) {
	return s.vmAction(ctx, vmID, "reboot")
}

// Resize resizes a VM's CPU and RAM.
func (s *VMServiceOp) Resize(ctx context.Context, vmID string, resizeReq *VMResizeRequest) (*Response, error) {
	path := fmt.Sprintf("%s/%s/resize", vmsBasePath, vmID)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, resizeReq)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *VMServiceOp) vmAction(ctx context.Context, vmID, action string) (*Response, error) {
	path := fmt.Sprintf("%s/%s/%s", vmsBasePath, vmID, action)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
