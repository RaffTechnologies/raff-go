package raff

import (
	"encoding/json"
	"testing"
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
