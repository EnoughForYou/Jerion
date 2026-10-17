package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/contract"
	"github.com/EnoughForYou/Jerion/internal/engine"
	"github.com/EnoughForYou/Jerion/internal/intel"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/preflight"
	"github.com/EnoughForYou/Jerion/internal/remediate"
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
		ID:       "scan-test",
		DeviceID: "dev-MOCK-0001",
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

	// 1. Initial default config
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

	// 2. Reject non-loopback address
	err = app.ClamAVSaveConfig(true, "10.0.0.5:3310")
	if err == nil {
		t.Fatalf("expected ClamAVSaveConfig to reject non-loopback address 10.0.0.5:3310")
	}
	if !errors.Is(err, intel.ErrClamAVNonLoopback) {
		t.Fatalf("expected ErrClamAVNonLoopback, got %v", err)
	}

	// 3. Reject invalid port
	err = app.ClamAVSaveConfig(true, "127.0.0.1:99999")
	if err == nil {
		t.Fatalf("expected ClamAVSaveConfig to reject port 99999")
	}
	if !errors.Is(err, intel.ErrClamAVInvalidAddress) {
		t.Fatalf("expected ErrClamAVInvalidAddress, got %v", err)
	}

	// 4. Save valid loopback address
	if err := app.ClamAVSaveConfig(true, "127.0.0.2:3310"); err != nil {
		t.Fatalf("ClamAVSaveConfig failed for valid loopback: %v", err)
	}

	cfgUpdated, err := app.ClamAVGetConfig()
	if err != nil {
		t.Fatalf("ClamAVGetConfig failed: %v", err)
	}
	if !cfgUpdated.Enabled || cfgUpdated.Host != "127.0.0.2:3310" {
		t.Fatalf("expected enabled=true host=127.0.0.2:3310, got %+v", cfgUpdated)
	}

	// 5. Test connection with non-loopback host returns error status without panicking
	statusNonLoopback, err := app.ClamAVTestConnection("192.168.1.1:3310")
	if err != nil {
		t.Fatalf("unexpected error from TestConnection: %v", err)
	}
	if statusNonLoopback.Connected {
		t.Fatalf("expected Connected false for non-loopback host")
	}
	if !strings.Contains(statusNonLoopback.Error, "loopback") {
		t.Fatalf("expected error mentioning loopback, got %q", statusNonLoopback.Error)
	}

	// 6. Test connection to unreachable loopback endpoint
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

func TestRemediateFailClosedDeviceBinding(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	// 1. Without active device, RemediateCreatePlan must fail closed
	app.activeSerial = ""
	_, err := app.RemediateCreatePlan([]string{"find-001"})
	if err == nil || !strings.Contains(err.Error(), "DeviceOffline") {
		t.Fatalf("expected DeviceOffline error when no device is connected, got: %v", err)
	}

	// 2. Without active device, RemediateExecute must fail closed
	_, err = app.RemediateExecute(contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: "act-1", Type: "disable", Package: "com.test.app"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "DeviceOffline") {
		t.Fatalf("expected DeviceOffline error in execute when no device is connected, got: %v", err)
	}

	// 3. Finding belonging to a different device must be ignored
	// Insert scan result and finding for Device A
	devAScanID := "scan-dev-a"
	now := time.Now().UTC()
	_ = app.store.SaveScanResult(store.ScanRecord{
		ID:           devAScanID,
		DeviceSerial: "DEV-A",
		Mode:         "quick",
		StartedAt:    now,
		FinishedAt:   now,
		Findings: []store.FindingRecord{
			{
				ID:       "find-dev-a",
				ScanID:   devAScanID,
				Package:  "com.other.device.app",
				Severity: "high",
				Actions:  []string{"disable"},
			},
		},
	})

	// Set active device to Device B in tracker
	fake.SetResponse("devices -l", "DEV-B device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "DEV-B"

	plan, err := app.RemediateCreatePlan([]string{"find-dev-a"})
	if err != nil {
		t.Fatalf("unexpected error creating plan: %v", err)
	}
	if len(plan.Actions) != 0 {
		t.Fatalf("expected 0 actions because finding belongs to DEV-A, not DEV-B, got: %d", len(plan.Actions))
	}
}

func TestRemediatePrevalidationAndCanonicalDTOs(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	activeDev := "DEV-VALID"
	fake.SetResponse("devices -l", "DEV-VALID device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = activeDev

	// Prepopulate backend cache with a valid action
	cachedAct := app.remediate.CreateAction(remediate.ActionDisable, "com.trusted.app", "", "", false)
	cachedAct.DeviceSerial = activeDev
	app.cachedActions[cachedAct.ID] = cachedAct

	// Test 1: Plan with valid action + invalid/expired action
	// Pre-validation must fail immediately before executing the valid action
	fake.SetResponse("pm disable-user --user 0 com.trusted.app", "Package com.trusted.app new state: disabled-user", "", nil)

	bogusAct := contract.RemediationAction{
		ID:      "act-bogus-999",
		Type:    "disable",
		Package: "com.bogus.app",
	}

	_, err := app.RemediateExecute(contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: cachedAct.ID, Type: cachedAct.Type, Package: cachedAct.Package},
			bogusAct,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "InvalidAction") {
		t.Fatalf("expected InvalidAction pre-validation failure, got: %v", err)
	}

	// Verify the valid action was NOT executed and remains intact in cache
	if _, exists := app.cachedActions[cachedAct.ID]; !exists {
		t.Fatalf("expected cached action to still exist because execution was aborted during pre-validation")
	}

	// Test 2: Canonical DTO verification: caller sends tampered title/package in actDto
	tamperedActDto := contract.RemediationAction{
		ID:      cachedAct.ID,
		Type:    cachedAct.Type,
		Package: "com.tampered.client.package", // Tampered by client
		Title:   "Tampered Title",
	}

	results, err := app.RemediateExecute(contract.RemediationPlan{
		Actions: []contract.RemediationAction{tamperedActDto},
	})
	if err != nil {
		t.Fatalf("RemediateExecute failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Result action must reflect canonical backend values, NOT tampered client DTO
	if results[0].Action.Package != "com.trusted.app" {
		t.Fatalf("expected canonical package %q, got: %q", "com.trusted.app", results[0].Action.Package)
	}
	if results[0].Action.Title != cachedAct.Title {
		t.Fatalf("expected canonical title %q, got: %q", cachedAct.Title, results[0].Action.Title)
	}
}

func TestAppsListThreeStateUnknowns(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkg := "com.test.overlayapp"
	fake.SetResponse("pm list packages", fmt.Sprintf("package:/data/app/test/base.apk=%s\n", pkg), "", nil)
	fake.SetResponse(fmt.Sprintf("dumpsys package %s", pkg), fmt.Sprintf(`Package [%s]
userId=10199
versionName=1.0.0
versionCode=10
targetSdk=31
requested permissions:
  android.permission.SYSTEM_ALERT_WINDOW
`, pkg), "", nil)

	// AppOps fails -> returns error
	fake.SetResponse("dumpsys appops", "", "could not find service: appops", os.ErrNotExist)
	fake.SetResponse(fmt.Sprintf("cmd appops get %s SYSTEM_ALERT_WINDOW", pkg), "", "could not find service: appops", os.ErrNotExist)
	fake.SetResponse(fmt.Sprintf("appops get %s SYSTEM_ALERT_WINDOW", pkg), "", "could not find service: appops", os.ErrNotExist)

	// Device policy fails -> returns error
	fake.SetResponse("dumpsys device_policy", "", "could not find service: device_policy", os.ErrNotExist)
	fake.SetResponse("dpm list-owners", "", "could not find command", os.ErrNotExist)

	// Battery whitelist fails -> returns error
	fake.SetResponse("dumpsys deviceidle whitelist", "", "could not find service: deviceidle", os.ErrNotExist)
	fake.SetResponse(fmt.Sprintf("dumpsys deviceidle is-whitelisted %s", pkg), "", "could not find service: deviceidle", os.ErrNotExist)

	// Test InventoryGet with detailed permissions
	appInfo, err := app.InventoryGet(pkg)
	if err != nil || appInfo == nil {
		t.Fatalf("InventoryGet failed: %v", err)
	}

	if !containsString(appInfo.SuspiciousFlags, "overlay_status_unknown") {
		t.Errorf("expected overlay_status_unknown when AppOps is unavailable, got: %v", appInfo.SuspiciousFlags)
	}
	if !containsString(appInfo.SuspiciousFlags, "device_admin_unknown") {
		t.Errorf("expected device_admin_unknown when device policy is unavailable, got: %v", appInfo.SuspiciousFlags)
	}
	if !containsString(appInfo.SuspiciousFlags, "battery_status_unknown") {
		t.Errorf("expected battery_status_unknown when battery whitelist is unavailable, got: %v", appInfo.SuspiciousFlags)
	}
	if appInfo.IsBatteryOptimized {
		t.Errorf("expected IsBatteryOptimized=false when whitelist query failed")
	}
}

func TestRemediateCreatePlan_RejectEmptyDeviceID(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkgTarget := "com.malware.emptydev"
	app.mu.Lock()
	// Scan result with empty DeviceID
	app.cachedResults["scan-empty-dev"] = &contract.ScanResult{
		ID:        "scan-empty-dev",
		DeviceID:  "", // empty device ID
		Mode:      "quick",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Status:    "completed",
		Findings: []contract.Finding{
			{
				ID:       "find-empty-dev",
				RuleID:   "RULE-EMPTY-DEV",
				Title:    "Malware from Unknown Device",
				Severity: "high",
				Package:  &pkgTarget,
				RecommendedAction: &contract.RemediationAction{
					ID:      "rec-empty-dev",
					Type:    "disable",
					Package: pkgTarget,
					Title:   "Disable App",
				},
			},
		},
	}
	app.mu.Unlock()

	// Finding from unknown/empty device ID must be rejected, yielding 0 actions
	plan, err := app.RemediateCreatePlan([]string{"find-empty-dev"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan returned unexpected error: %v", err)
	}
	if len(plan.Actions) != 0 {
		t.Fatalf("expected 0 actions from finding with empty DeviceID, got %d", len(plan.Actions))
	}
}

func TestRemediateExecute_DuplicateActionID(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	actID := "act-dup-1"
	app.mu.Lock()
	app.cachedActions[actID] = remediate.Action{
		ID:           actID,
		Type:         remediate.ActionDisable,
		Package:      "com.test.duplicate",
		DeviceSerial: "MOCK-0001",
		Title:        "Disable duplicate test",
	}
	app.mu.Unlock()

	plan := contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: actID, Type: "disable", Package: "com.test.duplicate"},
			{ID: actID, Type: "disable", Package: "com.test.duplicate"}, // duplicate!
		},
	}

	_, err := app.RemediateExecute(plan)
	if err == nil {
		t.Fatalf("expected error for duplicate action IDs, got nil")
	}
	if !strings.Contains(err.Error(), "Duplicate action ID") {
		t.Fatalf("expected 'Duplicate action ID' in error message, got: %v", err)
	}

	// Action should NOT have been consumed from cache since validation failed before reservation
	app.mu.Lock()
	_, exists := app.cachedActions[actID]
	app.mu.Unlock()
	if !exists {
		t.Fatalf("expected action to remain in cache when plan is rejected for duplicate IDs")
	}
}

func TestRemediateExecute_AtomicReservationConcurrency(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	fake.SetResponse("pm disable-user --user 0 com.test.concurrent", "Package com.test.concurrent new state: disabled-user\n", "", nil)

	actID := "act-concurrent-1"
	app.mu.Lock()
	app.cachedActions[actID] = remediate.Action{
		ID:           actID,
		Type:         remediate.ActionDisable,
		Package:      "com.test.concurrent",
		DeviceSerial: "MOCK-0001",
		Title:        "Disable concurrent test",
	}
	app.mu.Unlock()

	plan := contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: actID, Type: "disable", Package: "com.test.concurrent"},
		},
	}

	var wg sync.WaitGroup
	var successCount int
	var failCount int
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := app.RemediateExecute(plan)
			mu.Lock()
			defer mu.Unlock()
			if err == nil && len(results) == 1 && results[0].Status == "success" {
				successCount++
			} else if err != nil && strings.Contains(err.Error(), "InvalidAction") {
				failCount++
			}
		}()
	}

	wg.Wait()

	if successCount != 1 || failCount != 1 {
		t.Fatalf("expected exactly 1 success and 1 fail with InvalidAction, got successes=%d, fails=%d", successCount, failCount)
	}
}

func TestAppsListUnparseableAppOps(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkg := "com.test.unparseableops"
	fake.SetResponse("pm list packages", fmt.Sprintf("package:/data/app/test/base.apk=%s\n", pkg), "", nil)
	fake.SetResponse(fmt.Sprintf("dumpsys package %s", pkg), fmt.Sprintf(`Package [%s]
userId=10199
versionName=1.0.0
versionCode=10
targetSdk=31
requested permissions:
  android.permission.SYSTEM_ALERT_WINDOW
`, pkg), "", nil)

	fake.SetResponse("dumpsys device_policy", "No active admins\n", "", nil)
	fake.SetResponse("dumpsys deviceidle whitelist", "com.android.systemui\n", "", nil)

	// 1. Unparseable stdout with exit 0 (e.g. error banner or unexpected format)
	fake.SetResponse(fmt.Sprintf("cmd appops get %s SYSTEM_ALERT_WINDOW", pkg), "Segmentation fault (core dumped)\n", "", nil)
	fake.SetResponse(fmt.Sprintf("appops get %s SYSTEM_ALERT_WINDOW", pkg), "Segmentation fault (core dumped)\n", "", nil)

	appInfo, err := app.InventoryGet(pkg)
	if err != nil || appInfo == nil {
		t.Fatalf("InventoryGet failed: %v", err)
	}

	if !containsString(appInfo.SuspiciousFlags, "overlay_status_unknown") {
		t.Errorf("expected overlay_status_unknown when AppOps output is unparseable, got: %v", appInfo.SuspiciousFlags)
	}

	// 2. Valid recognized AppOps output stating permission is not granted
	fake.SetResponse(fmt.Sprintf("cmd appops get %s SYSTEM_ALERT_WINDOW", pkg), "No operations.\n", "", nil)
	fake.SetResponse(fmt.Sprintf("appops get %s SYSTEM_ALERT_WINDOW", pkg), "No operations.\n", "", nil)
	appInfo2, err2 := app.InventoryGet(pkg)
	if err2 != nil || appInfo2 == nil {
		t.Fatalf("InventoryGet second call failed: %v", err2)
	}
	if containsString(appInfo2.SuspiciousFlags, "overlay_status_unknown") {
		t.Errorf("overlay_status_unknown should NOT be present when AppOps recognized 'No operations.'")
	}
	if containsString(appInfo2.SuspiciousFlags, "overlay_granted") {
		t.Errorf("overlay_granted should NOT be present")
	}
}

func TestRemediateHistory_AuditPersistenceFields(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	fake.SetResponse("pm disable-user --user 0 com.test.hist", "Package com.test.hist new state: disabled-user\n", "", nil)

	actID := "act-hist-1"
	app.mu.Lock()
	app.cachedActions[actID] = remediate.Action{
		ID:           actID,
		Type:         remediate.ActionDisable,
		Package:      "com.test.hist",
		DeviceSerial: "MOCK-0001",
		Title:        "Disable History App",
	}
	app.mu.Unlock()

	plan := contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: actID, Type: "disable", Package: "com.test.hist"},
		},
	}

	results, err := app.RemediateExecute(plan)
	if err != nil || len(results) != 1 {
		t.Fatalf("RemediateExecute failed: %v", err)
	}

	history, err := app.RemediateGetHistory()
	if err != nil || len(history) != 1 {
		t.Fatalf("RemediateGetHistory failed: %v", err)
	}

	entry := history[0]
	if entry.AuditSaved == nil || !*entry.AuditSaved {
		t.Errorf("expected AuditSaved to be true in history entry")
	}
	if entry.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", entry.Status)
	}
}

func TestConvertEngineScanResult_PartialVerdictReview(t *testing.T) {
	app, _ := setupTestApp(t)
	defer app.store.Close()

	now := time.Now().UTC()
	partialResult := &engine.ScanResult{
		ID:           "scan-partial-001",
		DeviceSerial: "MOCK-0001",
		Mode:         "deep",
		StartedAt:    now.Add(-10 * time.Second),
		FinishedAt:   now,
		Findings:     []engine.Finding{}, // 0 findings!
		Partial:      true,
		TotalApps:    5,
	}

	contractRes := app.convertEngineScanResult(partialResult)
	if contractRes.Status != "partial" {
		t.Errorf("expected status 'partial', got '%s'", contractRes.Status)
	}
	if contractRes.Summary.Verdict != "review" {
		t.Errorf("expected verdict 'review' for partial scan without findings, got '%s'", contractRes.Summary.Verdict)
	}
}

func TestConvertStoreScanRecord_PartialVerdictReview(t *testing.T) {
	app, _ := setupTestApp(t)
	defer app.store.Close()

	now := time.Now().UTC()
	storeRecord := store.ScanRecord{
		ID:           "scan-store-partial-001",
		DeviceSerial: "MOCK-0001",
		Mode:         "deep",
		StartedAt:    now.Add(-10 * time.Second),
		FinishedAt:   now,
		Findings:     []store.FindingRecord{}, // 0 findings!
		Partial:      true,
		TotalApps:    10,
	}

	if err := app.store.SaveScanResult(storeRecord); err != nil {
		t.Fatalf("SaveScanResult failed: %v", err)
	}

	contractRes, err := app.ScanGet("scan-store-partial-001")
	if err != nil || contractRes == nil {
		t.Fatalf("ScanGet failed: %v", err)
	}
	if contractRes.Status != "partial" {
		t.Errorf("expected status 'partial', got '%s'", contractRes.Status)
	}
	if contractRes.Summary.Verdict != "review" {
		t.Errorf("expected verdict 'review' for stored partial scan without findings, got '%s'", contractRes.Summary.Verdict)
	}
}

func TestInventoryGet_AppOpsFallback(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	pkg := "com.test.fallbackops"
	fake.SetResponse(fmt.Sprintf("dumpsys package %s", pkg), fmt.Sprintf(`Package [%s]
userId=10200
versionName=1.0.0
versionCode=10
targetSdk=31
requested permissions:
  android.permission.SYSTEM_ALERT_WINDOW
`, pkg), "", nil)

	// cmd appops returns exit 0 but unrecognized output (e.g. unknown service error banner)
	fake.SetResponse(fmt.Sprintf("cmd appops get %s SYSTEM_ALERT_WINDOW", pkg), "cmd: Can't find service: appops\n", "", nil)
	// fallback appops get returns valid recognized allow
	fake.SetResponse(fmt.Sprintf("appops get %s SYSTEM_ALERT_WINDOW", pkg), "SYSTEM_ALERT_WINDOW: allow\n", "", nil)

	appInfo, err := app.InventoryGet(pkg)
	if err != nil || appInfo == nil {
		t.Fatalf("InventoryGet failed: %v", err)
	}

	if !appInfo.HasOverlayPermission {
		t.Errorf("expected HasOverlayPermission=true via appops fallback")
	}
	if !containsString(appInfo.SuspiciousFlags, "overlay_granted") {
		t.Errorf("expected overlay_granted in SuspiciousFlags, got: %v", appInfo.SuspiciousFlags)
	}
}

func TestRemediateCreatePlan_NilCachedResult(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	fake.SetResponse("devices -l", "MOCK-0001 device product:mock model:MockPhone device:mock_dev\n", "", nil)
	_, _ = app.tracker.PollOnce(context.Background())
	app.activeSerial = "MOCK-0001"

	app.mu.Lock()
	// Insert nil entry in cachedResults to ensure defensive iteration
	app.cachedResults["scan-nil"] = nil
	app.mu.Unlock()

	plan, err := app.RemediateCreatePlan([]string{"find-nonexistent"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed: %v", err)
	}
	if len(plan.Actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(plan.Actions))
	}
}

func TestConvertPreflightCheck_SourceURLPopulation(t *testing.T) {
	app, _ := setupTestApp(t)
	defer app.store.Close()

	// Missing P1 should populate SourceURL with official Google platform-tools download URL
	missingP1 := preflight.PreflightCheck{
		ID:          "P1",
		Title:       "ADB Binary",
		Description: "Missing",
		Status:      preflight.StatusMissing,
	}
	dto1 := app.convertPreflightCheck(missingP1)
	if dto1.SourceURL == nil || *dto1.SourceURL == "" {
		t.Fatalf("expected non-empty SourceURL for missing P1 check, got nil")
	}
	if !strings.Contains(*dto1.SourceURL, "dl.google.com/android/repository/platform-tools") {
		t.Errorf("expected official Google download URL, got %s", *dto1.SourceURL)
	}

	// OK P1 should have nil SourceURL
	okP1 := preflight.PreflightCheck{
		ID:          "P1",
		Title:       "ADB Binary",
		Description: "Ready",
		Status:      preflight.StatusOk,
	}
	dto2 := app.convertPreflightCheck(okP1)
	if dto2.SourceURL != nil {
		t.Errorf("expected nil SourceURL for ok P1 check, got %v", *dto2.SourceURL)
	}
}

func TestPreflightSelectAdbPath_EmptyContextError(t *testing.T) {
	app, _ := setupTestApp(t)
	defer app.store.Close()

	// Empty path with nil a.ctx should return window context not ready error
	app.ctx = nil
	err := app.PreflightSelectAdbPath("")
	if err == nil {
		t.Fatalf("expected error when choosing path with nil window context, got nil")
	}
	if !strings.Contains(err.Error(), "Window context not ready") {
		t.Errorf("unexpected error string: %v", err)
	}
}

func TestConvertDevice_UnauthorizedAndModelBrandParsing(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()

	// Case 1: Unauthorized device
	unauthDev := adb.DeviceInfo{
		Serial: "UNAUTH-123",
		State:  adb.StateUnauthorized,
	}
	resUnauth := app.convertDevice(unauthDev)
	if resUnauth == nil {
		t.Fatalf("expected non-nil device for unauthorized state")
	}
	if resUnauth.ConnectionState != "unauthorized" {
		t.Errorf("expected ConnectionState 'unauthorized', got '%s'", resUnauth.ConnectionState)
	}
	if resUnauth.Manufacturer != "Pending Authorization" {
		t.Errorf("expected Manufacturer 'Pending Authorization', got '%s'", resUnauth.Manufacturer)
	}
	if resUnauth.AndroidVersion != "Pending Authorization" {
		t.Errorf("expected AndroidVersion 'Pending Authorization', got '%s'", resUnauth.AndroidVersion)
	}
	if resUnauth.Model != "Unauthorized Android Device" {
		t.Errorf("expected default Model 'Unauthorized Android Device', got '%s'", resUnauth.Model)
	}
	if resUnauth.BatteryLevel != 0 || resUnauth.IsCharging {
		t.Errorf("expected 0%% battery and not charging for unauthorized, got %d%% charging=%v", resUnauth.BatteryLevel, resUnauth.IsCharging)
	}

	// Case 2: Ready device with brand & marketname fallback
	fake.SetResponse("getprop", "[ro.product.brand]: [Infinix]\n[ro.product.marketname]: [NOTE 11S]\n[ro.build.version.release]: [12]\n[ro.build.version.sdk]: [31]\n", "", nil)
	fake.SetResponse("dumpsys battery", "Current Battery Service state:\n  level: 95\n  scale: 100\n  status: 2\n", "", nil)

	readyDev := adb.DeviceInfo{
		Serial: "READY-456",
		State:  adb.StateReady,
	}
	resReady := app.convertDevice(readyDev)
	if resReady == nil {
		t.Fatalf("expected non-nil device for ready state")
	}
	if resReady.ConnectionState != "ready" {
		t.Errorf("expected ConnectionState 'ready', got '%s'", resReady.ConnectionState)
	}
	if resReady.Manufacturer != "Infinix" {
		t.Errorf("expected Manufacturer 'Infinix' from ro.product.brand fallback, got '%s'", resReady.Manufacturer)
	}
	if resReady.Model != "NOTE 11S" {
		t.Errorf("expected Model 'NOTE 11S' from ro.product.marketname, got '%s'", resReady.Model)
	}
	if resReady.AndroidVersion != "12" {
		t.Errorf("expected AndroidVersion '12', got '%s'", resReady.AndroidVersion)
	}
	if resReady.SDKInt != 31 {
		t.Errorf("expected SDKInt 31, got %d", resReady.SDKInt)
	}
	if resReady.BatteryLevel != 95 || !resReady.IsCharging {
		t.Errorf("expected 95%% battery charging=true, got %d%% charging=%v", resReady.BatteryLevel, resReady.IsCharging)
	}

	// Case 3: Ready device with ro.product.model taking precedence over marketname
	fake.SetResponse("getprop", "[ro.product.manufacturer]: [Transsion]\n[ro.product.brand]: [Infinix]\n[ro.product.model]: [Infinix X698]\n[ro.product.marketname]: [NOTE 11S]\n", "", nil)
	resReady2 := app.convertDevice(readyDev)
	if resReady2.Model != "Infinix X698" {
		t.Errorf("expected ro.product.model to take precedence, got '%s'", resReady2.Model)
	}
	if resReady2.Manufacturer != "Transsion" {
		t.Errorf("expected ro.product.manufacturer 'Transsion' to take precedence over brand, got '%s'", resReady2.Manufacturer)
	}

	// Case 4: Offline device
	offlineDev := adb.DeviceInfo{
		Serial: "OFFLINE-789",
		State:  adb.StateOffline,
	}
	resOffline := app.convertDevice(offlineDev)
	if resOffline == nil {
		t.Fatalf("expected non-nil device for offline state")
	}
	if resOffline.ConnectionState != "offline" {
		t.Errorf("expected ConnectionState 'offline', got '%s'", resOffline.ConnectionState)
	}
	if resOffline.Manufacturer != "Offline" {
		t.Errorf("expected Manufacturer 'Offline', got '%s'", resOffline.Manufacturer)
	}
	if resOffline.AndroidVersion != "Offline" {
		t.Errorf("expected AndroidVersion 'Offline', got '%s'", resOffline.AndroidVersion)
	}
	if resOffline.Model != "Offline Android Device" {
		t.Errorf("expected default Model 'Offline Android Device', got '%s'", resOffline.Model)
	}
}

func TestRemediateDeviceUnauthorizedForwarding(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()
	ctx := context.Background()

	// Simulate connected device in unauthorized state (USB RSA debugging prompt not accepted)
	fake.SetResponse("devices -l", "List of devices attached\nMOCK-UNAUTH unauthorized transport_id:2\n", "", nil)
	_, _ = app.tracker.PollOnce(ctx)

	// 1. RemediateCreatePlan must forward DeviceUnauthorized, NOT DeviceOffline
	_, err := app.RemediateCreatePlan([]string{"find-unauth"})
	if err == nil {
		t.Fatalf("expected error from RemediateCreatePlan on unauthorized device, got nil")
	}
	if !strings.Contains(err.Error(), "DeviceUnauthorized") {
		t.Errorf("expected DeviceUnauthorized error code, got: %v", err)
	}
	if strings.Contains(err.Error(), "DeviceOffline") {
		t.Errorf("error should not be masked as DeviceOffline: %v", err)
	}

	// 2. RemediateExecute must forward DeviceUnauthorized, NOT DeviceOffline
	_, err = app.RemediateExecute(contract.RemediationPlan{
		Actions: []contract.RemediationAction{
			{ID: "act-unauth", Type: "disable", Package: "com.test.app"},
		},
	})
	if err == nil {
		t.Fatalf("expected error from RemediateExecute on unauthorized device, got nil")
	}
	if !strings.Contains(err.Error(), "DeviceUnauthorized") {
		t.Errorf("expected DeviceUnauthorized error code in execute, got: %v", err)
	}
	if strings.Contains(err.Error(), "DeviceOffline") {
		t.Errorf("error should not be masked as DeviceOffline in execute: %v", err)
	}
}

func TestRemediateUpdatedSystemApp(t *testing.T) {
	app, fake := setupTestApp(t)
	defer app.store.Close()
	ctx := context.Background()

	serial := "SYS-APP-001"
	fake.SetResponse("devices -l", fmt.Sprintf("List of devices attached\n%s device transport_id:1\n", serial), "", nil)
	_, _ = app.tracker.PollOnce(ctx)
	app.activeSerial = serial

	pkgUpdated := "com.google.android.youtube"
	pkgUser := "com.cleaner.junk"

	// Mock ADB responses for updated system app
	// Note: Active APK path is in /data/app/ due to Play Store update
	fake.SetResponse("pm path com.google.android.youtube", "package:/data/app/~~xyz/com.google.android.youtube-1/base.apk\n", "", nil)
	// But Package Manager reports it as a system package
	fake.SetResponse("pm list packages -s com.google.android.youtube", "package:com.google.android.youtube\n", "", nil)

	// Mock ADB responses for regular user app
	fake.SetResponse("pm path com.cleaner.junk", "package:/data/app/~~abc/com.cleaner.junk-1/base.apk\n", "", nil)
	fake.SetResponse("pm list packages -s com.cleaner.junk", "", "", nil)
	fake.SetResponse("dumpsys package com.cleaner.junk", "Package [com.cleaner.junk]:\n  pkgFlags=[ HAS_CODE ]\n", "", nil)

	// Cache scan result with findings for both apps
	app.mu.Lock()
	app.cachedResults["scan-test"] = &contract.ScanResult{
		ID:       "scan-test",
		DeviceID: fmt.Sprintf("dev-%s", serial),
		Findings: []contract.Finding{
			{
				ID:      "find-updated-sys",
				Package: &pkgUpdated,
				RecommendedAction: &contract.RemediationAction{
					Type: remediate.ActionUninstall,
				},
			},
			{
				ID:      "find-user-app",
				Package: &pkgUser,
				RecommendedAction: &contract.RemediationAction{
					Type: remediate.ActionUninstall,
				},
			},
		},
	}
	app.mu.Unlock()

	// 1. RemediateCreatePlan for updated system app
	planSys, err := app.RemediateCreatePlan([]string{"find-updated-sys"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed for updated system app: %v", err)
	}
	if len(planSys.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(planSys.Actions))
	}
	sysAct := planSys.Actions[0]
	if !strings.Contains(sysAct.Title, "Uninstall system package") || !strings.Contains(sysAct.Title, "for user 0") {
		t.Errorf("expected system uninstall title for updated system app, got: %s", sysAct.Title)
	}
	if planSys.RequiresBackup {
		t.Errorf("expected RequiresBackup=false for system package uninstall")
	}

	// 2. RemediateExecute for updated system app should execute pm uninstall --user 0
	fake.SetResponse("pm uninstall --user 0 com.google.android.youtube", "Success\n", "", nil)
	resSys, err := app.RemediateExecute(planSys)
	if err != nil {
		t.Fatalf("RemediateExecute failed for system app: %v", err)
	}
	if len(resSys) != 1 || resSys[0].Status != "success" {
		t.Fatalf("expected successful execution: %+v", resSys)
	}

	// 3. RemediateCreatePlan for user app must require backup
	planUser, err := app.RemediateCreatePlan([]string{"find-user-app"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed for user app: %v", err)
	}
	if len(planUser.Actions) != 1 {
		t.Fatalf("expected 1 action for user app, got %d", len(planUser.Actions))
	}
	userAct := planUser.Actions[0]
	if !strings.Contains(userAct.Title, "Uninstall application") || strings.Contains(userAct.Title, "for user 0") {
		t.Errorf("expected standard user uninstall title, got: %s", userAct.Title)
	}
	if !planUser.RequiresBackup {
		t.Errorf("expected RequiresBackup=true for user app uninstall")
	}
}

func TestScanQuickTerminalStates(t *testing.T) {
	terminalStatuses := []struct {
		name           string
		terminalStatus string
		mockScanErr    bool
	}{
		{name: "Terminal Completed", terminalStatus: "completed", mockScanErr: false},
		{name: "Terminal Failed", terminalStatus: "failed", mockScanErr: true},
		{name: "Terminal Partial", terminalStatus: "partial", mockScanErr: false},
		{name: "Terminal Cancelled", terminalStatus: "cancelled", mockScanErr: false},
	}

	for _, tc := range terminalStatuses {
		t.Run(tc.name, func(t *testing.T) {
			app, fake := setupTestApp(t)
			defer app.store.Close()
			ctx := context.Background()

			serial := "DEV-SCAN-01"
			fake.SetResponse("devices -l", fmt.Sprintf("List of devices attached\n%s device transport_id:1\n", serial), "", nil)
			_, _ = app.tracker.PollOnce(ctx)
			app.activeSerial = serial

			if tc.mockScanErr {
				fake.SetResponse("pm list packages -f -3", "", "pm failed", os.ErrPermission)
			} else if tc.terminalStatus == "completed" {
				fake.SetResponse("pm list packages -f -3", "", "", nil)
			} else {
				// For partial or cancelled, intercept the active cached scan result
				fake.SetResponse("pm list packages -f -3", "", "", nil)
				go func() {
					for i := 0; i < 20; i++ {
						time.Sleep(50 * time.Millisecond)
						app.mu.Lock()
						for id, res := range app.cachedResults {
							if res != nil {
								res.Status = tc.terminalStatus
								app.cachedResults[id] = res
								app.mu.Unlock()
								return
							}
						}
						app.mu.Unlock()
					}
				}()
			}

			start := time.Now()
			res, err := app.ScanQuick()
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("expected ScanQuick to return terminal result without error, got: %v", err)
			}
			if res == nil {
				t.Fatalf("expected non-nil ScanResult")
			}
			if res.Status != tc.terminalStatus {
				t.Errorf("expected status %s, got %s", tc.terminalStatus, res.Status)
			}
			if elapsed > 10*time.Second {
				t.Errorf("ScanQuick took too long (%v), failed to terminate promptly on status %s", elapsed, tc.terminalStatus)
			}
		})
	}
}


