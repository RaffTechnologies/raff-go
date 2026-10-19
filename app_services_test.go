package raff

import (
	"encoding/json"
	"testing"
)

// An update sends only the fields that are set, so a description change does
// not reset the port or turn previews off.
func TestUpdateAppServiceRequestSendsOnlySetFields(t *testing.T) {
	body, err := json.Marshal(&UpdateAppServiceRequest{Description: String("api"), PreviewsEnabled: Bool(false)})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), `{"description":"api","previews_enabled":false}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

// A create carries the GitHub connection, the variables and the database link
// under the names the API reads.
func TestCreateAppServiceRequestNewFields(t *testing.T) {
	body, err := json.Marshal(&CreateAppServiceRequest{Name: "web", ServiceType: "web", SourceType: "git",
		GitHubInstallationID: "8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a", LinkDatabaseID: "a1b2c3d4",
		EnvVars: []AppEnvVarInput{{Key: "API_KEY", Value: "x", IsSecret: true}}})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["github_installation_id"] != "8d0c5f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a" || got["link_database_id"] != "a1b2c3d4" {
		t.Errorf("body = %s", body)
	}
	vars, ok := got["env_vars"].([]any)
	if !ok || len(vars) != 1 || vars[0].(map[string]any)["is_secret"] != true {
		t.Errorf("env_vars = %v", got["env_vars"])
	}
}
