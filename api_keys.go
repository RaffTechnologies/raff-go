package raff

import (
	"context"

	"github.com/rafftechnologies/raff-go/spec"
)

// APIKey represents an API key (without the secret).
type APIKey = spec.APIKey

// APIKeyWithSecret is the create/regenerate response — includes the
// plaintext secret, which is only returned once.
type APIKeyWithSecret = spec.APIKeyWithSecret

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest = spec.CreateAPIKeyRequest

// UpdateAPIKeyRequest is the request body for updating an API key.
type UpdateAPIKeyRequest = spec.UpdateAPIKeyRequest

// APIKeyListOptions are the query parameters for listing API keys.
type APIKeyListOptions = spec.ListAPIKeysParams

// APIKeyService handles communication with the API key endpoints.
//
// API keys are scoped to the account.
type APIKeyService interface {
	List(ctx context.Context, opts *APIKeyListOptions) ([]APIKey, *Response, error)
	Get(ctx context.Context, keyID string) (*APIKey, *Response, error)
	Create(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithSecret, *Response, error)
	Update(ctx context.Context, keyID string, req *UpdateAPIKeyRequest) (*APIKey, *Response, error)
	Regenerate(ctx context.Context, keyID string) (*APIKeyWithSecret, *Response, error)
	Revoke(ctx context.Context, keyID string) (*Response, error)
}

// APIKeyServiceOp implements APIKeyService.
type APIKeyServiceOp struct {
	client *Client
}

var _ APIKeyService = &APIKeyServiceOp{}

func (s *APIKeyServiceOp) List(ctx context.Context, opts *APIKeyListOptions) ([]APIKey, *Response, error) {
	resp, err := s.client.spec.ListAPIKeysWithResponse(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	var keys []APIKey
	if resp.JSON200.Data != nil {
		keys = *resp.JSON200.Data
	}
	total := 0
	if resp.JSON200.Total != nil {
		total = *resp.JSON200.Total
	}
	return keys, responseFrom(resp.HTTPResponse, total), nil
}

func (s *APIKeyServiceOp) Get(ctx context.Context, keyID string) (*APIKey, *Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.GetAPIKeyWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *APIKeyServiceOp) Create(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithSecret, *Response, error) {
	resp, err := s.client.spec.CreateAPIKeyWithResponse(ctx, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON201 == nil || resp.JSON201.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON201.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *APIKeyServiceOp) Update(ctx context.Context, keyID string, req *UpdateAPIKeyRequest) (*APIKey, *Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.UpdateAPIKeyWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *APIKeyServiceOp) Regenerate(ctx context.Context, keyID string) (*APIKeyWithSecret, *Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.RegenerateAPIKeyWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

func (s *APIKeyServiceOp) Revoke(ctx context.Context, keyID string) (*Response, error) {
	id, err := parseUUID(keyID)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.spec.RevokeAPIKeyWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return responseFrom(resp.HTTPResponse, 0), nil
}
