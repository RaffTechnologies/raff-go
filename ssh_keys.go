package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// SSHKey represents a registered SSH public key.
type SSHKey = spec.SSHKey

// CreateSSHKeyRequest is the request body for registering an SSH key.
type CreateSSHKeyRequest = spec.CreateSSHKeyRequest

// UpdateSSHKeyRequest is the request body for renaming an SSH key.
type UpdateSSHKeyRequest = spec.UpdateSSHKeyRequest

// SSHKeyService handles communication with the SSH key endpoints.
//
// SSH keys are scoped to the account, not to a project — list returns every
// key the API key's account can see.
type SSHKeyService interface {
	List(ctx context.Context) ([]SSHKey, *Response, error)
	Get(ctx context.Context, keyID string) (*SSHKey, *Response, error)
	Create(ctx context.Context, req *CreateSSHKeyRequest) (*SSHKey, *Response, error)
	Update(ctx context.Context, keyID string, req *UpdateSSHKeyRequest) (*SSHKey, *Response, error)
	Delete(ctx context.Context, keyID string) (*Response, error)
}

// SSHKeyServiceOp implements SSHKeyService.
type SSHKeyServiceOp struct {
	client *Client
}

var _ SSHKeyService = &SSHKeyServiceOp{}

func (s *SSHKeyServiceOp) List(ctx context.Context) ([]SSHKey, *Response, error) {
	resp, err := s.client.spec.ListSSHKeysWithResponse(ctx)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var keys []SSHKey
	if resp.JSON200.Data != nil {
		keys = *resp.JSON200.Data
	}
	return keys, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SSHKeyServiceOp) Get(ctx context.Context, keyID string) (*SSHKey, *Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetSSHKeyWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SSHKeyServiceOp) Create(ctx context.Context, req *CreateSSHKeyRequest) (*SSHKey, *Response, error) {
	resp, err := s.client.spec.CreateSSHKeyWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SSHKeyServiceOp) Update(ctx context.Context, keyID string, req *UpdateSSHKeyRequest) (*SSHKey, *Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateSSHKeyWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *SSHKeyServiceOp) Delete(ctx context.Context, keyID string) (*Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.DeleteSSHKeyWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}

