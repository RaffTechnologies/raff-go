package raff

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The gateway sends "" for unset optional fields (no VPC, payg billing, not
// paused, a backup still running). These must decode, or every list with
// such a database fails.
func TestDatabaseDecodesEmptyOptionalFields(t *testing.T) {
	body := `{"id":"5b0d7b8a-7c3e-4f53-9a55-3a8f0d2f6b11","database_id":"a1b2c3d4","project_id":"0c7e3f3e-1f1d-4b7a-8a59-0e1f7d6c2b3a",
		"name":"orders","engine":"postgres","engine_version":"18","status":"running","status_message":"","region":"us-east",
		"vpc_id":"","subscription_id":"","suspended_at":"","public_access":false,"public_port":0,
		"created_at":"2026-10-05T10:00:00Z","updated_at":"2026-10-05T10:00:05Z"}`
	var d Database
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		t.Fatalf("decode database: %v", err)
	}
	if StringValue(d.DatabaseID) != "a1b2c3d4" || d.Status == nil || *d.Status != DatabaseStatusRunning {
		t.Fatalf("unexpected database: %+v", d)
	}

	backup := `{"id":"9e1f5a7e-5f39-4f0e-9d43-0b8f4a6f7c21","database_id":"5b0d7b8a-7c3e-4f53-9a55-3a8f0d2f6b11","backup_type":"on_demand",
		"status":"running","size_bytes":0,"backup_timestamp":"","created_at":"2026-10-05T10:00:00Z","completed_at":"","expires_at":"",
		"restorable":false,"restore_window_start":""}`
	var b DatabaseBackup
	if err := json.Unmarshal([]byte(backup), &b); err != nil {
		t.Fatalf("decode backup: %v", err)
	}

	metrics := `{"success":true,"cpu_used_millicores":0,"metrics_updated_at":""}`
	if _, _, err := decodeFlat[DatabaseMetrics](200, nil, []byte(metrics)); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
}

const upgradeDBID = "a1b2c3d4"

// upgradeDB is a database body in the shape the API answers with.
func upgradeDB(status, engineVersion, upgrade string) string {
	return `{"success":true,"database":{"database_id":"` + upgradeDBID + `","name":"orders","engine":"postgres","engine_version":"` + engineVersion +
		`","status":"` + status + `","status_message":"","last_version_upgrade":` + upgrade + `}}`
}

func upgradeRecord(state, message string) string {
	finished := ""
	if state == "done" || state == "analyzing" || state == "failed" {
		finished = "2026-11-25T10:01:10Z"
	}
	return `{"from_version":"18","to_version":"19","state":"` + state + `","message":"` + message +
		`","started_at":"2026-11-25T10:00:00Z","finished_at":"` + finished + `"}`
}

// Upgrade posts the version to the upgrade path and decodes the database it
// answers with; a 412 refusal comes back with the API's message unchanged.
func TestDatabaseUpgrade(t *testing.T) {
	const refusal = "Column public.events.payload uses type abstime, which PostgreSQL 19 cannot carry over."
	var gotMethod, gotPath, gotBody string
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotMethod, gotPath, gotBody = r.Method, r.URL.Path, string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusPreconditionFailed {
			_, _ = w.Write([]byte(`{"success":false,"error":"` + refusal + `"}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"database":{"database_id":"` + upgradeDBID + `","engine_version":"18","status":"upgrading",` +
			`"status_message":"Upgrading to PostgreSQL 19: taking a backup first","last_version_upgrade":` + upgradeRecord("backing_up", "") + `}}`))
	}))
	defer srv.Close()
	client := NewFromToken("test", SetBaseURL(srv.URL))

	db, _, err := client.Databases.Upgrade(context.Background(), upgradeDBID, "19")
	if err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v1/databases/"+upgradeDBID+"/upgrade" {
		t.Errorf("request %s %s", gotMethod, gotPath)
	}
	if strings.TrimSpace(gotBody) != `{"version":"19"}` {
		t.Errorf("body %s", gotBody)
	}
	if db.Status == nil || *db.Status != DatabaseStatusUpgrading || StringValue(db.StatusMessage) != "Upgrading to PostgreSQL 19: taking a backup first" {
		t.Errorf("unexpected database: %+v", db)
	}
	if u := db.LastVersionUpgrade; u == nil || u.State == nil || *u.State != DatabaseUpgradeBackingUp || StringValue(u.FinishedAt) != "" {
		t.Errorf("unexpected upgrade: %+v", db.LastVersionUpgrade)
	}

	status = http.StatusPreconditionFailed
	_, resp, err := client.Databases.Upgrade(context.Background(), upgradeDBID, "19")
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusPreconditionFailed || apiErr.Message != refusal {
		t.Fatalf("refusal: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("refusal response %+v", resp)
	}
}

// upgradeServer answers each GET with the next body; the last one repeats.
func upgradeServer(t *testing.T, bodies ...string) (*httptest.Server, *int) {
	t.Helper()
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/databases/"+upgradeDBID {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		body := bodies[min(polls, len(bodies)-1)]
		polls++
		w.Header().Set("Content-Type", "application/json")
		if body == "404" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success":false,"error":"Database not found"}`))
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &polls
}

// WaitForUpgrade keeps polling through the upgrade's steps and returns the
// database once it runs on the new version.
func TestDatabaseWaitForUpgradeSucceeds(t *testing.T) {
	srv, polls := upgradeServer(t,
		upgradeDB("upgrading", "18", upgradeRecord("backing_up", "")),
		upgradeDB("upgrading", "19", upgradeRecord("finishing", "")),
		upgradeDB("running", "19", upgradeRecord("analyzing", "")),
	)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := client.Databases.WaitForUpgrade(ctx, upgradeDBID, "19")
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if StringValue(db.EngineVersion) != "19" || *polls != 3 {
		t.Errorf("engine_version %q after %d polls", StringValue(db.EngineVersion), *polls)
	}
}

// A failed upgrade ends the wait with the API's own sentence; an earlier
// upgrade to another version does not count.
func TestDatabaseWaitForUpgradeFailed(t *testing.T) {
	const reason = "The upgrade to PostgreSQL 19 could not be completed, so your database is back on PostgreSQL 18, unchanged. What PostgreSQL reported: could not load library."
	srv, _ := upgradeServer(t,
		upgradeDB("running", "18", `{"from_version":"17","to_version":"18","state":"failed","message":"old","started_at":"","finished_at":""}`),
		upgradeDB("running", "18", upgradeRecord("failed", reason)),
	)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := client.Databases.WaitForUpgrade(ctx, upgradeDBID, "19")
	var upErr *UpgradeError
	if !errors.As(err, &upErr) || upErr.Message != reason || upErr.Version != "19" {
		t.Fatalf("error %v", err)
	}
	if err.Error() != "upgrade to PostgreSQL 19 did not complete: "+reason {
		t.Errorf("error text %q", err.Error())
	}
	if db == nil || StringValue(db.EngineVersion) != "18" {
		t.Errorf("database %+v", db)
	}
}

// A database that is gone or entered failed ends the wait with an error.
func TestDatabaseWaitForUpgradeGoneOrFailed(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"gone", "404", "not found"},
		{"failed", upgradeDB("failed", "18", upgradeRecord("upgrading", "")), "entered status failed"},
	}
	for _, c := range cases {
		srv, _ := upgradeServer(t, c.body)
		client := NewFromToken("test", SetBaseURL(srv.URL))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := client.Databases.WaitForUpgrade(ctx, upgradeDBID, "19")
		cancel()
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want %q", c.name, err, c.want)
		}
	}
}

// While the upgrade runs (or is undone), WaitForUpgrade waits until ctx ends.
func TestDatabaseWaitForUpgradeContextEnds(t *testing.T) {
	for _, state := range []string{"upgrading", "reverting", "undone"} {
		srv, _ := upgradeServer(t, upgradeDB("upgrading", "18", upgradeRecord(state, "")))
		client := NewFromToken("test", SetBaseURL(srv.URL))
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		_, err := client.Databases.WaitForUpgrade(ctx, upgradeDBID, "19")
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("%s: error %v, want deadline exceeded", state, err)
		}
	}
}

// WaitForStatus does not stop on the upgrading status.
func TestDatabaseWaitForStatusWaitsThroughUpgrading(t *testing.T) {
	srv, polls := upgradeServer(t,
		upgradeDB("upgrading", "18", upgradeRecord("upgrading", "")),
		upgradeDB("running", "19", upgradeRecord("done", "")),
	)
	client := NewFromToken("test", SetBaseURL(srv.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := client.Databases.WaitForStatus(ctx, upgradeDBID, DatabaseStatusRunning); err != nil || *polls != 2 {
		t.Fatalf("wait: %v after %d polls", err, *polls)
	}
}
