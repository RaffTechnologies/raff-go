package raff

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sshTestVM = "8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a"

// The three calls reach their routes; the fingerprint goes in the path as it is, slashes included, and the last key
// is removed only when asked.
func TestServerSSHKeys(t *testing.T) {
	const fp = "SHA256:KH7zm92O0/dbuPCSkvuS9zdDvBneuFaHt+qwZbp8OYU"
	keys := `{"success":true,"data":{"keys":[{"fingerprint":"` + fp + `","type":"ssh-ed25519","comment":"me","has_options":false,"saved_key_id":"","saved_key_name":"","public_key":"ssh-ed25519 AAAA me"}],"changed":1,"read_at":"2026-10-20T14:00:00Z"}}`
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = append(got, r.Method+" "+r.RequestURI+" "+string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(keys))
	}))
	defer srv.Close()
	client := NewFromToken("test", SetBaseURL(srv.URL))
	ctx := context.Background()

	list, _, err := client.VMs.ListSSHKeys(ctx, sshTestVM)
	if err != nil || len(*list.Keys) != 1 || (*list.Keys)[0].Fingerprint == nil || *(*list.Keys)[0].Fingerprint != fp {
		t.Fatalf("list: %+v %v", list, err)
	}
	if _, _, err := client.VMs.AddSSHKey(ctx, sshTestVM, &AddServerSSHKeyRequest{SSHKey: "ssh-ed25519 AAAA me", SaveAs: String("laptop")}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.VMs.RemoveSSHKey(ctx, sshTestVM, fp, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.VMs.RemoveSSHKey(ctx, sshTestVM, fp, &RemoveServerSSHKeyOptions{ConfirmLast: true}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /api/v1/vms/" + sshTestVM + "/ssh-keys ",
		"POST /api/v1/vms/" + sshTestVM + `/ssh-keys {"save_as":"laptop","ssh_key":"ssh-ed25519 AAAA me"}`,
		"DELETE /api/v1/vms/" + sshTestVM + "/ssh-keys/" + fp + " ",
		"DELETE /api/v1/vms/" + sshTestVM + "/ssh-keys/" + fp + "?confirm_last=true ",
	}
	for i := range want {
		if i >= len(got) || strings.TrimSpace(got[i]) != strings.TrimSpace(want[i]) {
			t.Errorf("request %d:\n got %q\nwant %q", i, got, want[i])
		}
	}
}

// A refusal's code is the error's Reason.
func TestServerSSHKeysRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"success":false,"error":"This is the last SSH key on web-01.","reason":"last_key"}`))
	}))
	defer srv.Close()
	_, _, err := NewFromToken("test", SetBaseURL(srv.URL)).VMs.RemoveSSHKey(context.Background(), sshTestVM, "SHA256:x", nil)
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Reason != SSHKeysLastKey || apiErr.StatusCode != http.StatusConflict || !strings.Contains(apiErr.Message, "last SSH key") {
		t.Fatalf("error %v", err)
	}
}
