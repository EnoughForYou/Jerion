package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

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
	plan, err := app.RemediateCreatePlan([]string{"find-1"})
	if err != nil {
		t.Fatalf("RemediateCreatePlan failed: %v", err)
	}
	if len(plan.Actions) != 1 {
		t.Fatalf("expected 1 action in plan, got %d", len(plan.Actions))
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

