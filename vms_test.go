package raff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
