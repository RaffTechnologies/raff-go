package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// VM represents a virtual machine. Aliased to the generated spec type so new
// fields propagate automatically when the OpenAPI spec changes.
type VM = spec.VM

// VMTag is a custom tag attached to a VM.
type VMTag = spec.VMTag

// CreateVMRequest is the request body for creating a VM.
type CreateVMRequest = spec.CreateVMRequest

// DeleteVMRequest is the optional request body for deleting a VM.
type DeleteVMRequest = spec.DeleteVMRequest

// RenameVMRequest is the request body for renaming a VM.
type RenameVMRequest = spec.RenameVMRequest

// ReinstallVMRequest is the request body for reinstalling a VM.
type ReinstallVMRequest = spec.ReinstallVMRequest

// ResizeVMRequest is the request body for resizing a VM.
type ResizeVMRequest = spec.ResizeVMRequest

// ResizeVMDiskRequest is the request body for resizing a VM's disk.
type ResizeVMDiskRequest = spec.ResizeVMDiskRequest

// ResizeResponse is the response from a resize operation, including billing.
type ResizeResponse = spec.ResizeResponse

// BulkDeleteVMsRequest is the request body for bulk-deleting VMs.
type BulkDeleteVMsRequest = spec.DeleteVMsBulkRequest

// BulkDeleteVMsResult is the per-VM result envelope from bulk delete.
type BulkDeleteVMsResult = spec.BulkDeleteVMsResult

// BulkDeleteVMItemResult is the result for a single VM in bulk delete.
type BulkDeleteVMItemResult = spec.BulkDeleteVMItemResult

// VMListOptions are the query parameters for listing VMs.
type VMListOptions = spec.ListVMsParams

// AddVMTagRequest is the request body for adding a tag to a VM.
type AddVMTagRequest = spec.AddVMTagRequest

// UpdateVMTagRequest is the request body for updating a tag.
type UpdateVMTagRequest = spec.UpdateVMTagRequest

// VMTagsResponse is the response shape for tag mutations — returns the full
// updated tag list for the VM.
type VMTagsResponse = spec.VMTagsResponse

// VMNote is a free-form note attached to a VM.
type VMNote = spec.VMNote

// VMNotesResponse holds personal and account notes for a VM.
type VMNotesResponse = spec.VMNotesResponse

// UpsertVMNoteRequest is the request body for creating or updating a note.
type UpsertVMNoteRequest = spec.UpsertVMNoteRequest

// VMNoteType is "personal" or "account".
type VMNoteType = spec.UpsertVMNoteParamsType

// GetVMNotesOptions are the optional filters for listing notes.
type GetVMNotesOptions = spec.GetVMNotesParams

// VMNotesFilterType is "personal" or "account" used by GetVMNotesOptions.Type.
type VMNotesFilterType = spec.GetVMNotesParamsType

// VMNetwork is a network interface attached to a VM.
type VMNetwork = spec.VMNetwork

// ListVMNetworksOptions are the optional filters for listing VM networks.
type ListVMNetworksOptions = spec.ListVMNetworksParams

// VMNetworkType is "public", "vpc", or "ipv6", used by ListVMNetworksOptions.Type.
type VMNetworkType = spec.ListVMNetworksParamsType

// AttachVPCRequest is the request body for attaching a VM to a VPC.
type AttachVPCRequest = spec.AttachVPCRequest

// AttachIPRequest is the request body for attaching a floating IP to a VM.
type AttachIPRequest = spec.AttachIPRequest

// AttachIPResponse is the response from attaching a floating IP.
type AttachIPResponse = spec.AttachIPResponse

// AttachSecurityGroupRequest is the request body for attaching a security group.
type AttachSecurityGroupRequest = spec.AttachSecurityGroupRequest

// SaveImageRequest is the request body for saving a VM disk as a custom image.
type SaveImageRequest = spec.SaveImageRequest

// VMImage is a custom OS image saved from a VM disk.
type VMImage = spec.VMImage

// VMService handles communication with the VM endpoints.
type VMService interface {
	List(ctx context.Context, opts *VMListOptions) ([]VM, *Response, error)
	Get(ctx context.Context, vmID string) (*VM, *Response, error)
	Create(ctx context.Context, req *CreateVMRequest) (*VM, *Response, error)
	Delete(ctx context.Context, vmID string, req *DeleteVMRequest) (*Response, error)
	BulkDelete(ctx context.Context, req *BulkDeleteVMsRequest) (*BulkDeleteVMsResult, *Response, error)
	Start(ctx context.Context, vmID string) (*Response, error)
	Stop(ctx context.Context, vmID string) (*Response, error)
	Reboot(ctx context.Context, vmID string) (*Response, error)
	Rename(ctx context.Context, vmID string, req *RenameVMRequest) (*Response, error)
	ResetPassword(ctx context.Context, vmID string) (*Response, error)
	Reinstall(ctx context.Context, vmID string, req *ReinstallVMRequest) (*Response, error)
	FactoryReset(ctx context.Context, vmID string) (*Response, error)
	Resize(ctx context.Context, vmID string, req *ResizeVMRequest) (*ResizeResponse, *Response, error)
	ResizeDisk(ctx context.Context, vmID string, req *ResizeVMDiskRequest) (*ResizeResponse, *Response, error)
	HardReboot(ctx context.Context, vmID string) (*Response, error)
	SaveImage(ctx context.Context, vmID string, req *SaveImageRequest) (*VMImage, *Response, error)
	ListNetworks(ctx context.Context, vmID string, opts *ListVMNetworksOptions) ([]VMNetwork, *Response, error)
	AttachVPC(ctx context.Context, vmID string, req *AttachVPCRequest) (*Response, error)
	DetachVPC(ctx context.Context, vmID string, nicID int) (*Response, error)
	AttachIP(ctx context.Context, vmID string, req *AttachIPRequest) (*AttachIPResponse, *Response, error)
	DetachIP(ctx context.Context, vmID string, nicID int) (*Response, error)
	AttachSecurityGroup(ctx context.Context, vmID string, req *AttachSecurityGroupRequest) (*Response, error)
	DetachSecurityGroup(ctx context.Context, vmID, securityGroupID string, nicID int) (*Response, error)
	AddTag(ctx context.Context, vmID string, req *AddVMTagRequest) ([]VMTag, *Response, error)
	UpdateTag(ctx context.Context, vmID, tagID string, req *UpdateVMTagRequest) ([]VMTag, *Response, error)
	RemoveTag(ctx context.Context, vmID, tagID string) ([]VMTag, *Response, error)
	GetNotes(ctx context.Context, vmID string, opts *GetVMNotesOptions) (*VMNotesResponse, *Response, error)
	UpsertNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error)
	UpdateNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error)
	AppendNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error)
}

// VMServiceOp implements VMService.
type VMServiceOp struct {
	client *Client
}

var _ VMService = &VMServiceOp{}

func (s *VMServiceOp) List(ctx context.Context, opts *VMListOptions) ([]VM, *Response, error) {
	resp, err := s.client.spec.ListVMsWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var vms []VM
	if resp.JSON200.Data != nil {
		vms = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return vms, responseFrom(resp.HTTPResponse, total), nil
}

func (s *VMServiceOp) Get(ctx context.Context, vmID string) (*VM, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetVMWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) Create(ctx context.Context, req *CreateVMRequest) (*VM, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.CreateVMWithResponse(ctx, &spec.CreateVMParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) Delete(ctx context.Context, vmID string, req *DeleteVMRequest) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	body := DeleteVMRequest{}
	if req != nil {
		body = *req
	}
	resp, err := s.client.spec.DeleteVMWithResponse(ctx, id, &spec.DeleteVMParams{XProjectID: projectID}, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) BulkDelete(ctx context.Context, req *BulkDeleteVMsRequest) (*BulkDeleteVMsResult, *Response, error) {
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.DeleteVMsBulkWithResponse(ctx, &spec.DeleteVMsBulkParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) Start(ctx context.Context, vmID string) (*Response, error) {
	return s.action(ctx, vmID, actionStart)
}

func (s *VMServiceOp) Stop(ctx context.Context, vmID string) (*Response, error) {
	return s.action(ctx, vmID, actionStop)
}

func (s *VMServiceOp) Reboot(ctx context.Context, vmID string) (*Response, error) {
	return s.action(ctx, vmID, actionReboot)
}

func (s *VMServiceOp) ResetPassword(ctx context.Context, vmID string) (*Response, error) {
	return s.action(ctx, vmID, actionResetPassword)
}

func (s *VMServiceOp) FactoryReset(ctx context.Context, vmID string) (*Response, error) {
	return s.action(ctx, vmID, actionFactoryReset)
}

func (s *VMServiceOp) Rename(ctx context.Context, vmID string, req *RenameVMRequest) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RenameVMWithResponse(ctx, id, &spec.RenameVMParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) Reinstall(ctx context.Context, vmID string, req *ReinstallVMRequest) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.ReinstallVMWithResponse(ctx, id, &spec.ReinstallVMParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) Resize(ctx context.Context, vmID string, req *ResizeVMRequest) (*ResizeResponse, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ResizeVMWithResponse(ctx, id, &spec.ResizeVMParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) ResizeDisk(ctx context.Context, vmID string, req *ResizeVMDiskRequest) (*ResizeResponse, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ResizeVMDiskWithResponse(ctx, id, &spec.ResizeVMDiskParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

type vmAction int

const (
	actionStart vmAction = iota
	actionStop
	actionReboot
	actionResetPassword
	actionFactoryReset
)

func (s *VMServiceOp) action(ctx context.Context, vmID string, a vmAction) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}

	switch a {
	case actionStart:
		resp, err := s.client.spec.StartVMWithResponse(ctx, id, &spec.StartVMParams{XProjectID: projectID})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
		}
		return responseFrom(resp.HTTPResponse, 0), nil
	case actionStop:
		resp, err := s.client.spec.StopVMWithResponse(ctx, id, &spec.StopVMParams{XProjectID: projectID})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
		}
		return responseFrom(resp.HTTPResponse, 0), nil
	case actionReboot:
		resp, err := s.client.spec.RebootVMWithResponse(ctx, id, &spec.RebootVMParams{XProjectID: projectID})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
		}
		return responseFrom(resp.HTTPResponse, 0), nil
	case actionResetPassword:
		resp, err := s.client.spec.ResetVMPasswordWithResponse(ctx, id, &spec.ResetVMPasswordParams{XProjectID: projectID})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
		}
		return responseFrom(resp.HTTPResponse, 0), nil
	case actionFactoryReset:
		resp, err := s.client.spec.FactoryResetVMWithResponse(ctx, id, &spec.FactoryResetVMParams{XProjectID: projectID})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
		}
		return responseFrom(resp.HTTPResponse, 0), nil
	}
	return nil, nil
}

func (s *VMServiceOp) HardReboot(ctx context.Context, vmID string) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.HardRebootVMWithResponse(ctx, id, &spec.HardRebootVMParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) SaveImage(ctx context.Context, vmID string, req *SaveImageRequest) (*VMImage, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.SaveVMDiskAsImageWithResponse(ctx, id, &spec.SaveVMDiskAsImageParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) ListNetworks(ctx context.Context, vmID string, opts *ListVMNetworksOptions) ([]VMNetwork, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListVMNetworksWithResponse(ctx, id, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var nets []VMNetwork
	if resp.JSON200.Data != nil {
		nets = *resp.JSON200.Data
	}
	return nets, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) AttachVPC(ctx context.Context, vmID string, req *AttachVPCRequest) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.AttachVMToVPCWithResponse(ctx, id, &spec.AttachVMToVPCParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) DetachVPC(ctx context.Context, vmID string, nicID int) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DetachVMFromVPCWithResponse(ctx, id, nicID, &spec.DetachVMFromVPCParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) AttachIP(ctx context.Context, vmID string, req *AttachIPRequest) (*AttachIPResponse, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AttachVMIPWithResponse(ctx, id, &spec.AttachVMIPParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) DetachIP(ctx context.Context, vmID string, nicID int) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DetachVMIPWithResponse(ctx, id, nicID, &spec.DetachVMIPParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) AttachSecurityGroup(ctx context.Context, vmID string, req *AttachSecurityGroupRequest) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.AttachVMSecurityGroupWithResponse(ctx, id, &spec.AttachVMSecurityGroupParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) DetachSecurityGroup(ctx context.Context, vmID, securityGroupID string, nicID int) (*Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, err
	}
	sgID, err := parseUUID(securityGroupID)
	if err != nil {
		return nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DetachVMSecurityGroupWithResponse(ctx, id, sgID, nicID, &spec.DetachVMSecurityGroupParams{XProjectID: projectID})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) AddTag(ctx context.Context, vmID string, req *AddVMTagRequest) ([]VMTag, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AddVMTagWithResponse(ctx, id, &spec.AddVMTagParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Tags, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) UpdateTag(ctx context.Context, vmID, tagID string, req *UpdateVMTagRequest) ([]VMTag, *Response, error) {
	vid, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	tid, err := parseUUID(tagID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateVMTagWithResponse(ctx, vid, tid, &spec.UpdateVMTagParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Tags, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) RemoveTag(ctx context.Context, vmID, tagID string) ([]VMTag, *Response, error) {
	vid, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	tid, err := parseUUID(tagID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.RemoveVMTagWithResponse(ctx, vid, tid, &spec.RemoveVMTagParams{XProjectID: projectID})
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Tags, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) GetNotes(ctx context.Context, vmID string, opts *GetVMNotesOptions) (*VMNotesResponse, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetVMNotesWithResponse(ctx, id, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) UpsertNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpsertVMNoteWithResponse(ctx, id, noteType, &spec.UpsertVMNoteParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Note == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Note, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) UpdateNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	t := spec.UpdateVMNoteParamsType(noteType)
	resp, err := s.client.spec.UpdateVMNoteWithResponse(ctx, id, t, &spec.UpdateVMNoteParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Note == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Note, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *VMServiceOp) AppendNote(ctx context.Context, vmID string, noteType VMNoteType, req *UpsertVMNoteRequest) (*VMNote, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := s.client.requireProjectID()
	if err != nil {
		return nil, nil, err
	}
	t := spec.AppendVMNoteParamsType(noteType)
	resp, err := s.client.spec.AppendVMNoteWithResponse(ctx, id, t, &spec.AppendVMNoteParams{XProjectID: projectID}, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Note == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Note, responseFrom(resp.HTTPResponse, 0), nil
}
