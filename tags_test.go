package raff

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testProjectID = "a1b2c3d4-0000-4000-8000-000000000000"
	testTagID     = "6f0d2a4e-3b1c-4d5e-9f60-7a8b9c0d1e2f"
	testSGID      = "3f1c2b7a-9d4e-4c61-8a2f-5b6e7d8c9a01"
	testVMID      = "8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a"
	testTagJSON   = `{"id":"` + testTagID + `","name":"web","color":"blue","priority":null,"server_count":3,` +
		`"firewalls":[{"tag_id":"` + testTagID + `","tag_name":"web","project_id":"` + testProjectID + `",` +
		`"security_group_id":"` + testSGID + `","security_group_name":"web-server"}],` +
		`"created_at":"2026-11-10T09:00:00Z","updated_at":"2026-11-10T09:00:00Z"}`
	testFirewallTagsJSON = `{"success":true,"data":{"links":[{"tag_id":"` + testTagID + `","tag_name":"web","project_id":"` +
		testProjectID + `","security_group_id":"` + testSGID + `","security_group_name":"web-server"}],"job_counts":{"pending":2,"done":1}}}`
)

// request is what the test server saw.
type request struct {
	method, path, query, project, body string
}

func tagServer(t *testing.T, status int, reply string) (*httptest.Server, *request) {
	t.Helper()
	got := &request{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got = request{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("X-Project-ID"), string(b)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

// List reads data (the account tags), not the old tags/counts, and scopes the counts to the configured project.
func TestTagsList(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, `{"success":true,"tags":[],"counts":{"web":3},"data":[`+testTagJSON+`]}`)
	client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID(testProjectID))
	tags, _, err := client.Tags.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodGet || got.path != "/api/v1/tags" || got.project != testProjectID {
		t.Errorf("request %+v", got)
	}
	if len(tags) != 1 || tags[0].Name != "web" || tags[0].Color != "blue" || tags[0].Priority != nil || tags[0].ServerCount != 3 {
		t.Fatalf("tags %+v", tags)
	}
	if len(tags[0].Firewalls) != 1 || tags[0].Firewalls[0].SecurityGroupName != "web-server" {
		t.Errorf("firewalls %+v", tags[0].Firewalls)
	}
}

func TestTagsCreate(t *testing.T) {
	srv, got := tagServer(t, http.StatusCreated, `{"success":true,"data":`+testTagJSON+`}`)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	color := TagColor("blue")
	priority := 2
	tag, _, err := client.Tags.Create(context.Background(), &CreateTagRequest{Name: "web", Color: &color, Priority: &priority})
	if err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPost || got.path != "/api/v1/tags" {
		t.Errorf("request %+v", got)
	}
	for _, want := range []string{`"name":"web"`, `"color":"blue"`, `"priority":2`} {
		if !strings.Contains(got.body, want) {
			t.Errorf("body %s, want %s", got.body, want)
		}
	}
	if tag.ID.String() != testTagID {
		t.Errorf("tag %+v", tag)
	}
}

// A conflict comes back as an error carrying the server's message.
func TestTagsCreateConflict(t *testing.T) {
	srv, _ := tagServer(t, http.StatusConflict, `{"success":false,"error":"a tag with this name already exists"}`)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	_, _, err := client.Tags.Create(context.Background(), &CreateTagRequest{Name: "web"})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %v", err)
	}
}

// Priority 0 is sent, so the server removes the priority; fields left nil are not sent.
func TestTagsUpdateClearsPriority(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, `{"success":true,"data":`+testTagJSON+`}`)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	zero := 0
	if _, _, err := client.Tags.Update(context.Background(), testTagID, &UpdateTagRequest{Priority: &zero}); err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPatch || got.path != "/api/v1/tags/"+testTagID {
		t.Errorf("request %+v", got)
	}
	if !strings.Contains(got.body, `"priority":0`) || strings.Contains(got.body, "name") || strings.Contains(got.body, "color") {
		t.Errorf("body %s", got.body)
	}
}

func TestTagsDelete(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, `{"success":true,"message":"Tag deleted"}`)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	if _, err := client.Tags.Delete(context.Background(), testTagID); err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodDelete || got.path != "/api/v1/tags/"+testTagID {
		t.Errorf("request %+v", got)
	}
}

func TestSecurityGroupSetTags(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, testFirewallTagsJSON)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	ft, _, err := client.SecurityGroups.SetTags(context.Background(), testSGID, []string{testTagID})
	if err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPut || got.path != "/api/v1/security-groups/"+testSGID+"/tags" {
		t.Errorf("request %+v", got)
	}
	if !strings.Contains(got.body, `"tag_ids":["`+testTagID+`"]`) {
		t.Errorf("body %s", got.body)
	}
	if len(ft.Links) != 1 || ft.JobCounts.Pending == nil || *ft.JobCounts.Pending != 2 {
		t.Errorf("firewall tags %+v", ft)
	}
}

// An empty list is sent as [] (unlink every tag), not null.
func TestSecurityGroupSetTagsEmpty(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, `{"success":true,"data":{"links":[],"job_counts":{}}}`)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	if _, _, err := client.SecurityGroups.SetTags(context.Background(), testSGID, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.body, `"tag_ids":[]`) {
		t.Errorf("body %s", got.body)
	}
}

func TestSecurityGroupGetTags(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, testFirewallTagsJSON)
	client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID(testProjectID))
	ft, _, err := client.SecurityGroups.GetTags(context.Background(), testSGID)
	if err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodGet || got.path != "/api/v1/security-groups/"+testSGID+"/tags" || got.project != testProjectID {
		t.Errorf("request %+v", got)
	}
	if len(ft.Links) != 1 || ft.Links[0].TagName != "web" {
		t.Errorf("firewall tags %+v", ft)
	}
}

func TestVMTagFirewall(t *testing.T) {
	srv, got := tagServer(t, http.StatusOK, `{"success":true,"data":{"firewalls":[],"conflict":true,"job_status":"conflict","job_error":""}}`)
	client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID(testProjectID))
	tf, _, err := client.VMs.TagFirewall(context.Background(), testVMID)
	if err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodGet || got.path != "/api/v1/vms/"+testVMID+"/tag-firewall" || got.project != testProjectID {
		t.Errorf("request %+v", got)
	}
	if !tf.Conflict || tf.JobStatus == nil || *tf.JobStatus != "conflict" {
		t.Errorf("tag firewall %+v", tf)
	}
}

// RemoveTag sends reset_firewall=true only when asked.
func TestVMRemoveTagResetFirewall(t *testing.T) {
	cases := []struct {
		name  string
		opts  []*RemoveVMTagOptions
		query string
	}{
		{"default", nil, ""},
		{"reset", []*RemoveVMTagOptions{{ResetFirewall: true}}, "reset_firewall=true"},
	}
	for _, c := range cases {
		srv, got := tagServer(t, http.StatusOK, `{"success":true,"tags":[]}`)
		client := NewFromToken("test", SetBaseURL(srv.URL), SetProjectID(testProjectID))
		if _, _, err := client.VMs.RemoveTag(context.Background(), testVMID, testTagID, c.opts...); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got.method != http.MethodDelete || got.path != "/api/v1/vms/"+testVMID+"/tags/"+testTagID || got.query != c.query {
			t.Errorf("%s: request %+v", c.name, got)
		}
	}
}
