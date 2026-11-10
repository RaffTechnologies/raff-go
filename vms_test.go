package raff

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// A reset the server refused comes back as 200 with success false: ResetPassword returns the server's reason as an
// error. A reset that went through returns no error.
func TestResetPasswordRefused(t *testing.T) {
	const vmID = "8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a"
	cases := []struct {
		name, body string
		wantErr    string
	}{
		{"set", `{"success":true,"message":"Password reset successfully.","result":"set"}`, ""},
		{"refused", `{"success":false,"error":"Your server refused the new password: The password does not meet the password policy requirements."}`, "password policy"},
	}
	for _, c := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/vms/"+vmID+"/reset-password" {
				t.Errorf("%s: request %s %s", c.name, r.Method, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(c.body))
		}))
		client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID("a1b2c3d4-0000-4000-8000-000000000000"))
		_, err := client.VMs.ResetPassword(context.Background(), vmID)
		srv.Close()
		switch {
		case c.wantErr == "" && err != nil:
			t.Errorf("%s: error %v", c.name, err)
		case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
			t.Errorf("%s: error %v, want %q", c.name, err, c.wantErr)
		}
	}
}

// Create sends security_group_id, the firewall the server's public cards carry from the first boot, and leaves it
// out when it is not set.
func TestCreateSecurityGroupID(t *testing.T) {
	const sgID = "3f1c2b7a-9d4e-4c61-8a2f-5b6e7d8c9a01"
	sg, err := parseUUID(sgID)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		sg   *openapi_types.UUID
		want string
	}{
		{"set", &sg, `"security_group_id":"` + sgID + `"`},
		{"unset", nil, ""},
	}
	for _, c := range cases {
		var body string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a","name":"web-01"}}`))
		}))
		client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID("a1b2c3d4-0000-4000-8000-000000000000"))
		_, _, err := client.VMs.Create(context.Background(), &CreateVMRequest{Name: "web-01", Region: "us-east", PricingID: 78, SecurityGroupID: c.sg})
		srv.Close()
		if err != nil {
			t.Fatalf("%s: error %v", c.name, err)
		}
		if c.want != "" && !strings.Contains(body, c.want) {
			t.Errorf("%s: body %s, want %s", c.name, body, c.want)
		}
		if c.want == "" && strings.Contains(body, "security_group_id") {
			t.Errorf("%s: body %s sends security_group_id", c.name, body)
		}
	}
}
