package contract_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/contract"
)

func verifyFixture[T any](t *testing.T, fixturePath string) {
	t.Helper()

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixturePath, err)
	}

	// 1. Strict Unmarshal with DisallowUnknownFields
	var target T
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&target); err != nil {
		t.Fatalf("strict decode failed for %s: %v", filepath.Base(fixturePath), err)
	}

	// 2. Round-trip Marshal
	outBytes, err := json.Marshal(target)
	if err != nil {
		t.Fatalf("marshal failed for %s: %v", filepath.Base(fixturePath), err)
	}

	// 3. Re-decode to ensure structural round-trip equivalence
	var rawOriginal map[string]any
	var rawOut map[string]any
	if err := json.Unmarshal(data, &rawOriginal); err != nil {
		t.Fatalf("unmarshal original json to map failed: %v", err)
	}
	if err := json.Unmarshal(outBytes, &rawOut); err != nil {
		t.Fatalf("unmarshal output json to map failed: %v", err)
	}

	// Compare top-level keys
	for k := range rawOriginal {
		if _, exists := rawOut[k]; !exists {
			t.Errorf("round-trip missing key %q in %s", k, filepath.Base(fixturePath))
		}
	}
}

func TestContractFixturesValidation(t *testing.T) {
	baseDir := filepath.Join("..", "testdata", "contract")

	t.Run("ActionLogEntry", func(t *testing.T) {
		verifyFixture[contract.ActionLogEntry](t, filepath.Join(baseDir, "action_log_entry.json"))
	})

	t.Run("AllowlistEntry", func(t *testing.T) {
		verifyFixture[contract.AllowlistEntry](t, filepath.Join(baseDir, "allowlist_entry.json"))
	})

	t.Run("AppInfo", func(t *testing.T) {
		verifyFixture[contract.AppInfo](t, filepath.Join(baseDir, "app_info.json"))
	})

	t.Run("AuditItem", func(t *testing.T) {
		verifyFixture[contract.AuditItem](t, filepath.Join(baseDir, "audit_item.json"))
	})

	t.Run("Device", func(t *testing.T) {
		verifyFixture[contract.Device](t, filepath.Join(baseDir, "device.json"))
	})

	t.Run("Evidence", func(t *testing.T) {
		verifyFixture[contract.Evidence](t, filepath.Join(baseDir, "evidence.json"))
	})

	t.Run("Finding", func(t *testing.T) {
		verifyFixture[contract.Finding](t, filepath.Join(baseDir, "finding.json"))
	})

	t.Run("MonitorCandidate", func(t *testing.T) {
		verifyFixture[contract.MonitorCandidate](t, filepath.Join(baseDir, "monitor_candidate.json"))
	})

	t.Run("MonitorEvent", func(t *testing.T) {
		verifyFixture[contract.MonitorEvent](t, filepath.Join(baseDir, "monitor_event.json"))
	})

	t.Run("MonitorSnapshot", func(t *testing.T) {
		verifyFixture[contract.MonitorSnapshot](t, filepath.Join(baseDir, "monitor_snapshot.json"))
	})

	t.Run("PreflightCheck", func(t *testing.T) {
		verifyFixture[contract.PreflightCheck](t, filepath.Join(baseDir, "preflight_check.json"))
	})

	t.Run("RemediationAction", func(t *testing.T) {
		verifyFixture[contract.RemediationAction](t, filepath.Join(baseDir, "remediation_action.json"))
	})

	t.Run("RemediationPlan", func(t *testing.T) {
		verifyFixture[contract.RemediationPlan](t, filepath.Join(baseDir, "remediation_plan.json"))
	})

	t.Run("ReportExportOptions", func(t *testing.T) {
		verifyFixture[contract.ReportExportOptions](t, filepath.Join(baseDir, "report_export_options.json"))
	})

	t.Run("ScanComparison", func(t *testing.T) {
		verifyFixture[contract.ScanComparison](t, filepath.Join(baseDir, "scan_comparison.json"))
	})

	t.Run("ScanOptions", func(t *testing.T) {
		verifyFixture[contract.ScanOptions](t, filepath.Join(baseDir, "scan_options.json"))
	})

	t.Run("ScanProgress", func(t *testing.T) {
		verifyFixture[contract.ScanProgress](t, filepath.Join(baseDir, "scan_progress.json"))
	})

	t.Run("ScanResult", func(t *testing.T) {
		verifyFixture[contract.ScanResult](t, filepath.Join(baseDir, "scan_result.json"))
	})

	t.Run("ScanSummaryItem", func(t *testing.T) {
		verifyFixture[contract.ScanSummaryItem](t, filepath.Join(baseDir, "scan_summary_item.json"))
	})

	t.Run("SecuritySummary", func(t *testing.T) {
		verifyFixture[contract.SecuritySummary](t, filepath.Join(baseDir, "security_summary.json"))
	})

	t.Run("Settings", func(t *testing.T) {
		verifyFixture[contract.Settings](t, filepath.Join(baseDir, "settings.json"))
	})
}

func init() {
	_ = reflect.TypeOf(contract.Device{})
}
