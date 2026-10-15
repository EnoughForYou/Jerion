package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/contract"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func setupTestApp(t *testing.T) (*App, *adb.FakeRunner) {
	t.Helper()

	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_app.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	logger := logx.GetDefault()
	app := NewAppWithDeps(fake, st, logger, tmpDir)
	return app, fake
}

func TestAppServicesIntegration(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	// 1. Preflight Service
	checks, err := app.PreflightRun()
	if err != nil {
		t.Fatalf("PreflightRun failed: %v", err)
	}
	if len(checks) == 0 {
		t.Fatalf("expected preflight checks, got 0")
	}

	// 2. Settings Service
	settings, err := app.SettingsGet()
	if err != nil {
		t.Fatalf("SettingsGet failed: %v", err)
	}
	if settings.Theme != "system" {
		t.Fatalf("expected default theme system, got %s", settings.Theme)
	}

	newTheme := "dark"
	updatedSettings, err := app.SettingsUpdate(contract.Settings{
		Theme:              newTheme,
		DeepScanTimeoutSec: 180,
	})
	if err != nil {
		t.Fatalf("SettingsUpdate failed: %v", err)
	}
	if updatedSettings.Theme != "dark" || updatedSettings.DeepScanTimeoutSec != 180 {
		t.Fatalf("unexpected updated settings: %+v", updatedSettings)
	}

	// 3. Allowlist Service
	allowEntry, err := app.AllowlistAdd(contract.AllowlistEntry{
		Type:   "package",
		Target: "com.whitelisted.app",
	})
	if err != nil {
		t.Fatalf("AllowlistAdd failed: %v", err)
	}
	if allowEntry.Target != "com.whitelisted.app" {
		t.Fatalf("expected target com.whitelisted.app, got %s", allowEntry.Target)
	}

	allowList, err := app.AllowlistList()
	if err != nil {
		t.Fatalf("AllowlistList failed: %v", err)
	}
	if len(allowList) != 1 {
		t.Fatalf("expected 1 allowlist entry, got %d", len(allowList))
	}

	err = app.AllowlistRemove(allowEntry.ID)
	if err != nil {
		t.Fatalf("AllowlistRemove failed: %v", err)
	}

	// 4. Device Service
	fake.SetResponse("devices -l", "List of devices attached\nMOCK-0001 device product:Infinix model:Note_11S device:X698 transport_id:1\n", "", nil)
	fake.SetResponse("getprop", "[ro.product.manufacturer]: [Infinix]\n[ro.build.version.release]: [12]\n[ro.build.version.sdk]: [31]\n", "", nil)
	fake.SetResponse("dumpsys battery", "Current Battery Service state:\n  level: 88\n  scale: 100\n  status: 3\n  health: 2\n", "", nil)

	_, _ = app.tracker.PollOnce(context.Background())
	devList, err := app.DeviceList()
	if err != nil {
		t.Fatalf("DeviceList failed: %v", err)
	}
	if len(devList) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devList))
	}
	if devList[0].Serial != "MOCK-0001" {
		t.Fatalf("expected serial MOCK-0001, got %s", devList[0].Serial)
	}

	err = app.DeviceSelect("MOCK-0001")
	if err != nil {
		t.Fatalf("DeviceSelect failed: %v", err)
	}

	connDev, err := app.DeviceGetConnected()
	if err != nil {
		t.Fatalf("DeviceGetConnected failed: %v", err)
	}
	if connDev == nil || connDev.Serial != "MOCK-0001" {
		t.Fatalf("unexpected connected device: %+v", connDev)
	}

	// 5. Inventory Service
	fake.SetResponse("pm list packages -f -u", "package:/data/app/com.example.app/base.apk=com.example.app\n", "", nil)
	fake.SetResponse("dumpsys package com.example.app", "Package [com.example.app] (12345):\n    versionCode=100 minSdk=21 targetSdk=30\n    versionName=1.0.0\n    pkgFlags=[ HAS_CODE ]\n", "", nil)

	inv, err := app.InventoryList(contract.InventoryFilter{Category: "all", Search: ""})
	if err != nil {
		t.Fatalf("InventoryList failed: %v", err)
	}
	if len(inv) != 1 || inv[0].Package != "com.example.app" {
		t.Fatalf("expected package com.example.app, got %+v", inv)
	}

	invDetail, err := app.InventoryGet("com.example.app")
	if err != nil {
		t.Fatalf("InventoryGet failed: %v", err)
	}
	if invDetail.Package != "com.example.app" {
		t.Fatalf("expected com.example.app, got %s", invDetail.Package)
	}

	// 6. Monitor Service
	err = app.MonitorStart()
	if err != nil {
		t.Fatalf("MonitorStart failed: %v", err)
	}
	snap, err := app.MonitorSnapshot()
	if err != nil {
		t.Fatalf("MonitorSnapshot failed: %v", err)
	}
	if !snap.Status.IsActive {
		t.Fatalf("expected active monitor status")
	}
	err = app.MonitorStop()
	if err != nil {
		t.Fatalf("MonitorStop failed: %v", err)
	}

	// 7. Audit Service
	fake.SetResponse("settings get secure accessibility_enabled", "0\n", "", nil)
	fake.SetResponse("settings get secure enabled_accessibility_services", "\n", "", nil)
	fake.SetResponse("settings get secure enabled_notification_listeners", "\n", "", nil)
	fake.SetResponse("dumpsys device_policy", "No active admins\n", "", nil)
	fake.SetResponse("settings get global install_non_market_apps", "0\n", "", nil)
	fake.SetResponse("settings get global adb_wifi_enabled", "0\n", "", nil)
	fake.SetResponse("settings get global adb_enabled", "1\n", "", nil)
	fake.SetResponse("settings get global private_dns_mode", "hostname\n", "", nil)
	fake.SetResponse("settings get global private_dns_specifier", "dns.google\n", "", nil)
	fake.SetResponse("settings get global http_proxy", ":0\n", "", nil)
	fake.SetResponse("settings get global global_http_proxy_host", "\n", "", nil)

	auditItems, err := app.AuditRun()
	if err != nil {
		t.Fatalf("AuditRun failed: %v", err)
	}
	if len(auditItems) == 0 {
		t.Fatalf("expected audit items, got 0")
	}

	latestAudit, err := app.AuditGetLatest()
	if err != nil || len(latestAudit) != len(auditItems) {
		t.Fatalf("AuditGetLatest mismatch: %v", err)
	}

	// 8. Remediate Service
	pkgSample := "com.adware.sample"
	app.mu.Lock()
	app.cachedResults["scan-test"] = &contract.ScanResult{
		ID: "scan-test",
		Findings: []contract.Finding{
			{
				ID:      "find-1",
				RuleID:  "rule-adware-1",
				Title:   "Suspicious Adware Found",
				Package: &pkgSample,
				RecommendedAction: &contract.RemediationAction{
					Type:    "disable",
					Package: pkgSample,
				},
			},
		},
	}
	app.mu.Unlock()

	plan, err := app.RemediateCreatePlan([]string{"find-1"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed: %v", err)
	}
	if len(plan.Actions) != 1 {
		t.Fatalf("expected 1 action in plan, got %d", len(plan.Actions))
	}
	if plan.Actions[0].Package != pkgSample {
		t.Fatalf("expected package %s, got %s", pkgSample, plan.Actions[0].Package)
	}

	// Unknown finding ID must not invent dummy packages
	unknownPlan, err := app.RemediateCreatePlan([]string{"non-existent-finding-id"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed: %v", err)
	}
	if len(unknownPlan.Actions) != 0 {
		t.Fatalf("expected 0 actions for unknown finding, got %d", len(unknownPlan.Actions))
	}

	// Unrecognized action ID in plan must be rejected by backend
	fakePlan := contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{
				ID:      "act-unrecognized-999",
				Type:    "disable",
				Package: "com.malicious.fake",
			},
		},
	}
	_, fakeErr := app.RemediateExecute(fakePlan)
	if fakeErr == nil || !strings.Contains(fakeErr.Error(), "InvalidAction") {
		t.Fatalf("expected InvalidAction error for unrecognized action ID, got: %v", fakeErr)
	}

	// Test permission cleaning
	if p := extractCleanPermission("Declared/Requested: android.permission.SYSTEM_ALERT_WINDOW"); p != "android.permission.SYSTEM_ALERT_WINDOW" {
		t.Fatalf("expected cleaned permission, got '%s'", p)
	}

	fake.SetResponse("pm disable-user --user 0 com.adware.sample", "Package com.adware.sample new state: disabled-user", "", nil)
	execLogs, err := app.RemediateExecute(plan)
	if err != nil {
		t.Fatalf("RemediateExecute failed: %v", err)
	}
	if len(execLogs) != 1 || execLogs[0].Status != "success" {
		t.Fatalf("expected success log, got %+v", execLogs)
	}

	hist, err := app.RemediateGetHistory()
	if err != nil || len(hist) != 1 {
		t.Fatalf("RemediateGetHistory failed: %v", err)
	}

	fake.SetResponse("pm enable com.adware.sample", "Package com.adware.sample new state: enabled", "", nil)
	undoneLog, err := app.RemediateUndo(execLogs[0].ID)
	if err != nil || undoneLog.Status != "undone" {
		t.Fatalf("RemediateUndo failed: %v", err)
	}

	// 9. Report Service & Canonical Comparison
	now := time.Now().UTC()
	scan1 := store.ScanRecord{
		ID:           "scan-comp-1",
		DeviceSerial: "MOCK-0001",
		Mode:         "quick",
		StartedAt:    now.Add(-2 * time.Hour),
		FinishedAt:   now.Add(-1 * time.Hour),
		RulesVersion: "1.0.0",
		AppVersion:   "0.4.0",
		Findings: []store.FindingRecord{
			{
				ID:         "find-old-resolved",
				ScanID:     "scan-comp-1",
				RuleID:     "RULE-A",
				Package:    "com.resolved.adware",
				Category:   "adware",
				Severity:   "high",
				Confidence: "high",
				Weight:     80,
				Title:      "Resolved Adware",
				WhyFlagged: "Old threat",
				Evidence:   []store.EvidenceRecord{{Source: "permission", Snippet: "android.permission.SYSTEM_ALERT_WINDOW"}},
				Actions:    []string{"disable"},
			},
			{
				ID:         "find-shared-unchanged",
				ScanID:     "scan-comp-1",
				RuleID:     "RULE-B",
				Package:    "com.shared.threat",
				Category:   "malware",
				Severity:   "critical",
				Confidence: "high",
				Weight:     95,
				Title:      "Persistent Threat",
				WhyFlagged: "Persisting threat",
				Evidence:   []store.EvidenceRecord{{Source: "hash", Snippet: "abc123"}},
				Actions:    []string{"uninstall"},
			},
		},
	}
	scan2 := store.ScanRecord{
		ID:           "scan-comp-2",
		DeviceSerial: "MOCK-0001",
		Mode:         "quick",
		StartedAt:    now.Add(-30 * time.Minute),
		FinishedAt:   now,
		RulesVersion: "1.0.0",
		AppVersion:   "0.4.0",
		Findings: []store.FindingRecord{
			{
				ID:         "find-shared-scan2",
				ScanID:     "scan-comp-2",
				RuleID:     "RULE-B",
				Package:    "com.shared.threat",
				Category:   "malware",
				Severity:   "critical",
				Confidence: "high",
				Weight:     95,
				Title:      "Persistent Threat",
				WhyFlagged: "Persisting threat",
				Evidence:   []store.EvidenceRecord{{Source: "hash", Snippet: "abc123"}},
				Actions:    []string{"uninstall"},
			},
			{
				ID:         "find-new-threat",
				ScanID:     "scan-comp-2",
				RuleID:     "RULE-C",
				Package:    "com.new.spyware",
				Category:   "spyware",
				Severity:   "medium",
				Confidence: "medium",
				Weight:     50,
				Title:      "New Spyware",
				WhyFlagged: "Newly introduced threat",
				Evidence:   []store.EvidenceRecord{{Source: "manifest", Snippet: "hidden_service"}},
				Actions:    []string{"disable"},
			},
		},
	}
	if err := app.store.SaveScanResult(scan1); err != nil {
		t.Fatalf("SaveScanResult scan1 failed: %v", err)
	}
	if err := app.store.SaveScanResult(scan2); err != nil {
		t.Fatalf("SaveScanResult scan2 failed: %v", err)
	}

	// Test ReportList
	reports, err := app.ReportList()
	if err != nil {
		t.Fatalf("ReportList failed: %v", err)
	}
	if len(reports) < 2 {
		t.Fatalf("expected at least 2 reports, got %d", len(reports))
	}
	if reports[0].Verdict != "high_risk" {
		t.Errorf("expected high_risk verdict for scan with critical finding, got %s", reports[0].Verdict)
	}

	// Test ReportCompare
	comp, err := app.ReportCompare(scan1.ID, scan2.ID)
	if err != nil {
		t.Fatalf("ReportCompare failed: %v", err)
	}
	if len(comp.NewFindings) != 1 {
		t.Errorf("expected 1 new finding, got %d", len(comp.NewFindings))
	}
	if len(comp.ResolvedFindings) != 1 {
		t.Errorf("expected 1 resolved finding, got %d", len(comp.ResolvedFindings))
	}
	if len(comp.UnchangedFindings) != 1 {
		t.Errorf("expected 1 unchanged finding, got %d", len(comp.UnchangedFindings))
	}

	// Test ReportExport (HTML and JSON)
	exportedJSON, err := app.ReportExport(contract.ReportExportOptions{ScanID: scan2.ID, Format: "json"})
	if err != nil {
		t.Fatalf("ReportExport JSON failed: %v", err)
	}
	if _, err := os.Stat(exportedJSON); err != nil {
		t.Errorf("expected exported JSON file to exist: %s", exportedJSON)
	}

	exportedHTML, err := app.ReportExport(contract.ReportExportOptions{ScanID: scan2.ID, Format: "html"})
	if err != nil {
		t.Fatalf("ReportExport HTML failed: %v", err)
	}
	if _, err := os.Stat(exportedHTML); err != nil {
		t.Errorf("expected exported HTML file to exist: %s", exportedHTML)
	}
}

func TestAppEdgeCases(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	ctx := context.Background()

	// 1. No device connected
	fake.SetResponse("devices -l", "List of devices attached\n\n", "", nil)
	_, _ = app.tracker.PollOnce(ctx)

	_, err := app.InventoryList(contract.InventoryFilter{Category: "all"})
	if err == nil || !strings.Contains(err.Error(), "DeviceOffline") {
		t.Fatalf("expected DeviceOffline when no device connected, got %v", err)
	}

	// 2. Unauthorized device
	fake.SetResponse("devices -l", "List of devices attached\nMOCK-UNAUTH unauthorized transport_id:2\n", "", nil)
	_, _ = app.tracker.PollOnce(ctx)

	_, err = app.InventoryList(contract.InventoryFilter{Category: "all"})
	if err == nil || !strings.Contains(err.Error(), "DeviceUnauthorized") {
		t.Fatalf("expected DeviceUnauthorized for unauthorized device, got %v", err)
	}
	if !strings.Contains(err.Error(), "RSA security fingerprint") {
		t.Fatalf("expected RSA guidance in error remedy, got %v", err)
	}

	// 3. Abrupt disconnect during session
	fake.SetResponse("devices -l", "List of devices attached\nMOCK-READY device transport_id:3\n", "", nil)
	_, _ = app.tracker.PollOnce(ctx)
	if app.tracker.ActiveDevice() == nil || app.tracker.ActiveDevice().Serial != "MOCK-READY" {
		t.Fatalf("expected active device MOCK-READY")
	}

	// Device unplugged!
	fake.SetResponse("devices -l", "List of devices attached\n\n", "", nil)
	_, _ = app.tracker.PollOnce(ctx)
	if app.tracker.ActiveDevice() != nil {
		t.Fatalf("expected active device to become nil after unplug, got %+v", app.tracker.ActiveDevice())
	}
}

func TestWrapADBError(t *testing.T) {
	errMissing := adb.NewADBError(adb.ErrCodeAdbMissing, "adb not found", nil)
	wErr := wrapADBError(errMissing)
	if !strings.Contains(wErr.Error(), "AdbMissing") {
		t.Fatalf("expected AdbMissing code, got: %v", wErr)
	}

	errUnauth := adb.NewADBError(adb.ErrCodeDeviceUnauthorized, "unauthorized", nil)
	wErr2 := wrapADBError(errUnauth)
	if !strings.Contains(wErr2.Error(), "DeviceUnauthorized") {
		t.Fatalf("expected DeviceUnauthorized code, got: %v", wErr2)
	}

	errTimeout := adb.NewADBError(adb.ErrCodeTimeout, "timeout", nil)
	wErr3 := wrapADBError(errTimeout)
	if !strings.Contains(wErr3.Error(), "Timeout") {
		t.Fatalf("expected Timeout code, got: %v", wErr3)
	}
}

func TestClamAVAppIntegration(t *testing.T) {
	app, _ := setupTestApp(t)
	defer app.store.Close()

	// 1. Initial config
	cfg, err := app.ClamAVGetConfig()
	if err != nil {
		t.Fatalf("ClamAVGetConfig failed: %v", err)
	}
	if cfg.Enabled {
		t.Fatalf("expected ClamAV disabled initially")
	}
	if cfg.Host != "127.0.0.1:3310" {
		t.Fatalf("expected default host 127.0.0.1:3310, got %s", cfg.Host)
	}

	// 2. Save config
	if err := app.ClamAVSaveConfig(true, "10.0.0.5:3310"); err != nil {
		t.Fatalf("ClamAVSaveConfig failed: %v", err)
	}

	cfgUpdated, err := app.ClamAVGetConfig()
	if err != nil {
		t.Fatalf("ClamAVGetConfig failed: %v", err)
	}
	if !cfgUpdated.Enabled || cfgUpdated.Host != "10.0.0.5:3310" {
		t.Fatalf("expected enabled=true host=10.0.0.5:3310, got %+v", cfgUpdated)
	}

	// 3. Test connection to unreachable endpoint
	status, err := app.ClamAVTestConnection("127.0.0.1:54321")
	if err != nil {
		t.Fatalf("unexpected error from TestConnection: %v", err)
	}
	if status.Connected {
		t.Fatalf("expected Connected false for non-listening host")
	}
	if status.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestAppMonitorWorkflow(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("dumpsys window", `
  mCurrentFocus=Window{12345 u0 com.popup.spammer/com.popup.spammer.AdActivity}
  Window #1 Window{67890 u0 com.popup.spammer/com.popup.spammer.OverlayView}:
    ty=TYPE_APPLICATION_OVERLAY
`, "", nil)

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	// Start monitor
	if err := app.MonitorStart(); err != nil {
		t.Fatalf("MonitorStart failed: %v", err)
	}

	snap, err := app.MonitorSnapshot()
	if err != nil {
		t.Fatalf("MonitorSnapshot failed: %v", err)
	}
	if !snap.Status.IsActive {
		t.Fatalf("expected monitor status active")
	}
	if len(snap.Status.SourcesActive) != 1 || snap.Status.SourcesActive[0] != "dumpsys window" {
		t.Fatalf("expected SourcesActive ['dumpsys window'], got %v", snap.Status.SourcesActive)
	}

	// Trigger MarkUserPopup
	if err := app.MonitorMarkPopup(); err != nil {
		t.Fatalf("MonitorMarkPopup failed: %v", err)
	}

	// Stop monitor
	if err := app.MonitorStop(); err != nil {
		t.Fatalf("MonitorStop failed: %v", err)
	}

	snapStopped, err := app.MonitorSnapshot()
	if err != nil {
		t.Fatalf("MonitorSnapshot failed: %v", err)
	}
	if snapStopped.Status.IsActive {
		t.Fatalf("expected monitor status inactive after stop")
	}
}

func TestRemediationDeviceMismatchAndSingleUse(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkgTarget := "com.adware.locked"
	nowStr := time.Now().UTC().Format(time.RFC3339)
	app.cachedResults["scan-h3"] = &contract.ScanResult{
		ID:        "scan-h3",
		DeviceID:  "dev-MOCK-0001",
		Mode:      "quick",
		StartedAt: nowStr,
		Status:    "completed",
		Findings: []contract.Finding{
			{
				ID:       "find-h3",
				RuleID:   "RULE-H3",
				Title:    "Flagged Threat",
				Severity: "high",
				Package:  &pkgTarget,
				RecommendedAction: &contract.RemediationAction{
					ID:       "rec-h3",
					Type:     "disable",
					Package:  pkgTarget,
					Title:    "Disable Threat",
					Impact:   "Freeze package",
					Undoable: true,
				},
			},
		},
	}

	plan, err := app.RemediateCreatePlan([]string{"find-h3"})
	if err != nil || len(plan.Actions) != 1 {
		t.Fatalf("RemediateCreatePlan failed: %v, plan: %+v", err, plan)
	}

	// 1. Switch active device to MOCK-0002 -> execution must be rejected with DeviceMismatch
	app.activeSerial = "MOCK-0002"
	_, errMismatch := app.RemediateExecute(plan)
	if errMismatch == nil || !strings.Contains(errMismatch.Error(), "DeviceMismatch") {
		t.Fatalf("expected DeviceMismatch error, got: %v", errMismatch)
	}

	// 2. Switch back to original planned device MOCK-0001 -> execution succeeds
	app.activeSerial = "MOCK-0001"
	fake.SetResponse("pm disable-user --user 0 com.adware.locked", "Package com.adware.locked new state: disabled-user", "", nil)
	logs, err := app.RemediateExecute(plan)
	if err != nil || len(logs) != 1 || logs[0].Status != "success" {
		t.Fatalf("RemediateExecute failed on matching device: %v, logs: %+v", err, logs)
	}

	// 3. Re-execution of the same executed plan must be rejected with InvalidAction (single-use)
	_, errReplay := app.RemediateExecute(plan)
	if errReplay == nil || !strings.Contains(errReplay.Error(), "InvalidAction") {
		t.Fatalf("expected InvalidAction error on replaying already executed plan, got: %v", errReplay)
	}
}

func TestRemediationStructuredFailureReporting(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkgTarget := "com.malware.uninstallfail"
	nowStr := time.Now().UTC().Format(time.RFC3339)
	app.cachedResults["scan-fail"] = &contract.ScanResult{
		ID:        "scan-fail",
		DeviceID:  "dev-MOCK-0001",
		Mode:      "quick",
		StartedAt: nowStr,
		Status:    "completed",
		Findings: []contract.Finding{
			{
				ID:       "find-fail",
				RuleID:   "RULE-FAIL",
				Title:    "Malware to Uninstall",
				Severity: "critical",
				Package:  &pkgTarget,
				RecommendedAction: &contract.RemediationAction{
					ID:       "rec-fail",
					Type:     "uninstall",
					Package:  pkgTarget,
					Title:    "Uninstall Malware",
					Impact:   "Remove malware",
					Undoable: true,
				},
			},
		},
	}

	plan, err := app.RemediateCreatePlan([]string{"find-fail"})
	if err != nil || len(plan.Actions) != 1 {
		t.Fatalf("RemediateCreatePlan failed: %v", err)
	}

	// pm path fails to simulate APK backup failure
	fake.SetResponse("pm path com.malware.uninstallfail", "", "package not found", os.ErrNotExist)

	// RemediateExecute should return structured failed log entry instead of hiding failure
	results, execErr := app.RemediateExecute(plan)
	if execErr != nil {
		t.Fatalf("RemediateExecute should return structured results, got unexpected top-level error: %v", execErr)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 structured log entry for failed action, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected status failed, got %s", results[0].Status)
	}
	if results[0].Error == nil || !strings.Contains(*results[0].Error, "BackupFailed") {
		t.Fatalf("expected BackupFailed in error details, got: %v", results[0].Error)
	}
}

func TestAppsListAppOpsAndAdminAccuracy(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkg := "com.test.threat"
	fake.SetResponse("pm list packages", fmt.Sprintf("package:/data/app/test/base.apk=%s\n", pkg), "", nil)
	fake.SetResponse(fmt.Sprintf("dumpsys package %s", pkg), fmt.Sprintf(`Package [%s]
userId=10199
versionName=1.0.0
versionCode=10
targetSdk=31
`, pkg), "", nil)

	// Mock AppOps overlay
	fake.SetResponse("dumpsys appops", `Current AppOps Service state:
  Package com.test.threat:
    SYSTEM_ALERT_WINDOW: allow; time=+5s ago
`, "", nil)

	// Mock Device Admin via dumpsys device_policy
	fake.SetResponse("dumpsys device_policy", `Current Device Policy Manager state:
  Enabled Device Admins (user 0):
    admin=com.test.threat/.AdminReceiver
`, "", nil)

	// Mock failure on battery whitelist
	fake.SetResponse("dumpsys deviceidle whitelist", "", "command not supported", os.ErrInvalid)

	apps, err := app.InventoryList(contract.InventoryFilter{Category: "all"})
	if err != nil || len(apps) != 1 {
		t.Fatalf("InventoryList failed: %v, len=%d", err, len(apps))
	}

	appInfo := apps[0]
	if !appInfo.HasOverlayPermission {
		t.Errorf("expected HasOverlayPermission=true from AppOps allow")
	}
	if !appInfo.IsDeviceAdmin {
		t.Errorf("expected IsDeviceAdmin=true from dumpsys device_policy")
	}
	if appInfo.IsBatteryOptimized {
		t.Errorf("expected IsBatteryOptimized=false when whitelist is unreadable")
	}
	if !containsString(appInfo.SuspiciousFlags, "battery_status_unknown") {
		t.Errorf("expected battery_status_unknown in SuspiciousFlags when whitelist fails, got: %v", appInfo.SuspiciousFlags)
	}
	if !containsString(appInfo.SuspiciousFlags, "overlay_granted") {
		t.Errorf("expected overlay_granted in SuspiciousFlags, got: %v", appInfo.SuspiciousFlags)
	}
	if !containsString(appInfo.SuspiciousFlags, "device_admin_active") {
		t.Errorf("expected device_admin_active in SuspiciousFlags, got: %v", appInfo.SuspiciousFlags)
	}
}

func TestScanEvidenceTypeMapping(t *testing.T) {
	if mapEvidenceType("declared permission") != "permission" {
		t.Errorf("expected permission evidence type")
	}
	if mapEvidenceType("intel_hash_match") != "hash" {
		t.Errorf("expected hash evidence type")
	}
	if mapEvidenceType("dex_class_scan") != "class_prefix" {
		t.Errorf("expected class_prefix evidence type")
	}
	if mapEvidenceType("logcat_stream") != "log" {
		t.Errorf("expected log evidence type")
	}
	if mapEvidenceType("secure_setting") != "setting" {
		t.Errorf("expected setting evidence type")
	}
	if mapEvidenceType("manifest_node") != "manifest" {
		t.Errorf("expected manifest evidence type")
	}
}



