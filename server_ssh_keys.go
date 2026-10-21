package raff

import (
	"context"
	"net/http"

	"github.com/rafftechnologies/raff-go/spec"
)

// ServerSSHKey is one key line of root's authorized_keys on a running Linux server.
type ServerSSHKey = spec.ServerSSHKey

// ServerSSHKeys is a server's keys after a read, add or remove.
type ServerSSHKeys = spec.ServerSSHKeys

// AddServerSSHKeyRequest adds a key to a server: SSHKey is a saved key's ID or a public key; SaveAs also saves a public
// key to the account under that name.
type AddServerSSHKeyRequest = spec.AddServerSSHKeyRequest

// RemoveServerSSHKeyOptions are the options of removing a key from a server.
type RemoveServerSSHKeyOptions struct {
	// ConfirmLast removes the key even when it is root's last key (a server created with only an SSH key has password
	// sign-in turned off).
	ConfirmLast bool
}

// Why a server's keys could not be read or changed: the ErrorResponse's Reason.
const (
	SSHKeysNotLinux        = "not_linux"
	SSHKeysInvalidKey      = "invalid_key"
	SSHKeysNotOnServer     = "not_on_server"
	SSHKeysNotRunning      = "not_running"
	SSHKeysNoGuestAgent    = "no_guest_agent"
	SSHKeysAgentBlocked    = "agent_blocked"
	SSHKeysSELinuxBlocked  = "selinux_blocked"
	SSHKeysLastKey         = "last_key"
	SSHKeysRefused         = "refused"
	SSHKeysHostUnavailable = "host_unavailable"
	SSHKeysContextFailed   = "context_failed"
	SSHKeysNoAnswer        = "no_answer"
)

// ListSSHKeys reads the keys in root's authorized_keys on a running Linux server, from the server itself (a few
// seconds). Each key is named after the saved key with the same fingerprint.
func (s *VMServiceOp) ListSSHKeys(ctx context.Context, vmID string) (*ServerSSHKeys, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.ListServerSSHKeysWithResponse(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// AddSSHKey adds a key to root's authorized_keys on a running Linux server, at once and without a restart. A key the
// server already has changes nothing (Changed 0).
func (s *VMServiceOp) AddSSHKey(ctx context.Context, vmID string, req *AddServerSSHKeyRequest) (*ServerSSHKeys, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.spec.AddServerSSHKeyWithResponse(ctx, id, *req)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// RemoveSSHKey removes the key with this fingerprint ("SHA256:...", as ssh-keygen -l shows it) from root's
// authorized_keys on a running Linux server, and from its settings so it does not come back at the next restart.
func (s *VMServiceOp) RemoveSSHKey(ctx context.Context, vmID, fingerprint string, opts *RemoveServerSSHKeyOptions) (*ServerSSHKeys, *Response, error) {
	id, err := parseUUID(vmID)
	if err != nil {
		return nil, nil, err
	}
	params := &spec.RemoveServerSSHKeyParams{}
	if opts != nil && opts.ConfirmLast {
		params.ConfirmLast = Bool(true)
	}
	resp, err := s.client.spec.RemoveServerSSHKeyWithResponse(ctx, id, fingerprint, params, literalPath)
	if err != nil {
		return nil, nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, responseFrom(resp.HTTPResponse, 0), errorFromResponse(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Data, responseFrom(resp.HTTPResponse, 0), nil
}

// literalPath sends the request path as it is, a fingerprint's slashes included: some proxies refuse an encoded
// slash (%2F), and the API takes the fingerprint's slashes literally.
func literalPath(_ context.Context, req *http.Request) error {
	req.URL.RawPath = ""
	return nil
}
