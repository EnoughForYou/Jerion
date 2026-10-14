package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/apk"
	"github.com/EnoughForYou/Jerion/internal/inventory"
	"github.com/EnoughForYou/Jerion/internal/rules"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func TestRuleEvaluatorPositiveAndNegative(t *testing.T) {
	loadedRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		t.Fatalf("LoadEmbeddedRules failed: %v", err)
	}
	evaluator := NewEvaluator(loadedRules)

	storeInstaller := "com.android.vending"

	// 1. ADW-001 (Overlay without launcher)
	// Positive: non-system, no launcher, has SYSTEM_ALERT_WINDOW
	posADW1 := inventory.AppInfo{
		Package:              "com.stealth.overlay",
		IsSystem:             false,
		HasLauncher:          false,
		RequestedPermissions: []string{"android.permission.SYSTEM_ALERT_WINDOW"},
	}
	findings := evaluator.EvaluateApp(posADW1, nil)
	if !hasFinding(findings, "ADW-001") {
		t.Errorf("expected ADW-001 positive match, got findings: %+v", findings)
	}

	// Negative: has launcher icon (legitimate utility)
	negADW1 := inventory.AppInfo{
		Package:              "com.legit.screenrecorder",
		IsSystem:             false,
		HasLauncher:          true,
		RequestedPermissions: []string{"android.permission.SYSTEM_ALERT_WINDOW"},
	}
	findings = evaluator.EvaluateApp(negADW1, nil)
	if hasFinding(findings, "ADW-001") {
		t.Errorf("expected ADW-001 negative (no match), but was flagged")
	}

	// 2. ADW-002 (>= 2 Ad SDKs + WAKE_LOCK)
	posADW2 := inventory.AppInfo{
		Package:              "com.intrusive.cleaner",
		IsSystem:             false,
		RequestedPermissions: []string{"android.permission.WAKE_LOCK"},
	}
	apkDetail2 := &apk.APKDetail{
		DetectedAdSDKs: map[string]string{
			"com/unity3d/ads": "Unity Ads",
			"com/applovin":    "AppLovin MAX",
		},
	}
	findings = evaluator.EvaluateApp(posADW2, apkDetail2)
	if !hasFinding(findings, "ADW-002") {
		t.Errorf("expected ADW-002 positive match")
	}

	// Negative: only 1 ad SDK
	apkDetail2Single := &apk.APKDetail{
		DetectedAdSDKs: map[string]string{
			"com/unity3d/ads": "Unity Ads",
		},
	}
	findings = evaluator.EvaluateApp(posADW2, apkDetail2Single)
	if hasFinding(findings, "ADW-002") {
		t.Errorf("expected ADW-002 negative for single ad SDK")
	}

	// 3. STK-001 (Hidden launcher + RECORD_AUDIO + FINE_LOCATION)
	posSTK1 := inventory.AppInfo{
		Package:              "com.spy.tracker",
		IsSystem:             false,
		HasLauncher:          false,
		RequestedPermissions: []string{"android.permission.RECORD_AUDIO", "android.permission.ACCESS_FINE_LOCATION"},
	}
	findings = evaluator.EvaluateApp(posSTK1, nil)
	if !hasFinding(findings, "STK-001") {
		t.Errorf("expected STK-001 positive match")
	}

	// Negative: has launcher icon (e.g. voice recorder)
	negSTK1 := inventory.AppInfo{
		Package:              "com.recorder.voice",
		IsSystem:             false,
		HasLauncher:          true,
		RequestedPermissions: []string{"android.permission.RECORD_AUDIO", "android.permission.ACCESS_FINE_LOCATION"},
	}
	findings = evaluator.EvaluateApp(negSTK1, nil)
	if hasFinding(findings, "STK-001") {
		t.Errorf("expected STK-001 negative for app with launcher icon")
	}

	// 4. DRO-001 (Sideloaded + REQUEST_INSTALL_PACKAGES)
	posDRO1 := inventory.AppInfo{
		Package:              "com.untrusted.dropper",
		IsSystem:             false,
		Installer:            nil, // sideloaded
		RequestedPermissions: []string{"android.permission.REQUEST_INSTALL_PACKAGES"},
	}
	findings = evaluator.EvaluateApp(posDRO1, nil)
	if !hasFinding(findings, "DRO-001") {
		t.Errorf("expected DRO-001 positive match")
	}

	// Negative: installed from official Google Play Store
	negDRO1 := inventory.AppInfo{
		Package:              "com.official.storeapp",
		IsSystem:             false,
		Installer:            &storeInstaller,
		RequestedPermissions: []string{"android.permission.REQUEST_INSTALL_PACKAGES"},
	}
	findings = evaluator.EvaluateApp(negDRO1, nil)
	if hasFinding(findings, "DRO-001") {
		t.Errorf("expected DRO-001 negative for Play Store installer")
	}

	// 5. ADM-001 (Sideloaded + BIND_DEVICE_ADMIN)
	posADM1 := inventory.AppInfo{
		Package:              "com.ransomware.admin",
		IsSystem:             false,
		Installer:            nil, // sideloaded
		RequestedPermissions: []string{"android.permission.BIND_DEVICE_ADMIN"},
	}
	findings = evaluator.EvaluateApp(posADM1, nil)
	if !hasFinding(findings, "ADM-001") {
		t.Errorf("expected ADM-001 positive match")
	}

	// Negative: official store install
	negADM1 := inventory.AppInfo{
		Package:              "com.enterprise.mdm",
		IsSystem:             false,
		Installer:            &storeInstaller,
		RequestedPermissions: []string{"android.permission.BIND_DEVICE_ADMIN"},
	}
	findings = evaluator.EvaluateApp(negADM1, nil)
	if hasFinding(findings, "ADM-001") {
		t.Errorf("expected ADM-001 negative for Play Store mdm app")
	}

	// 6. ADW-003 (Boot-Triggered Overlay Adware)
	posADW3 := inventory.AppInfo{
		Package:              "com.popup.bootspammer",
		IsSystem:             false,
		RequestedPermissions: []string{"android.permission.SYSTEM_ALERT_WINDOW", "android.permission.RECEIVE_BOOT_COMPLETED"},
	}
	findings = evaluator.EvaluateApp(posADW3, nil)
	if !hasFinding(findings, "ADW-003") {
		t.Errorf("expected ADW-003 positive match")
	}

	negADW3 := inventory.AppInfo{
		Package:              "com.popup.noboot",
		IsSystem:             false,
		RequestedPermissions: []string{"android.permission.SYSTEM_ALERT_WINDOW"},
	}
	findings = evaluator.EvaluateApp(negADW3, nil)
	if hasFinding(findings, "ADW-003") {
		t.Errorf("expected ADW-003 negative when missing RECEIVE_BOOT_COMPLETED")
	}

	// 7. BNK-001 (Banking Trojan Accessibility Overlay)
	posBNK1 := inventory.AppInfo{
		Package:              "com.malicious.bankhijack",
		IsSystem:             false,
		RequestedPermissions: []string{"android.permission.BIND_ACCESSIBILITY_SERVICE", "android.permission.SYSTEM_ALERT_WINDOW"},
	}
	findings = evaluator.EvaluateApp(posBNK1, nil)
	if !hasFinding(findings, "BNK-001") {
		t.Errorf("expected BNK-001 positive match")
	}

	negBNK1 := inventory.AppInfo{
		Package:              "com.accessibility.reader",
		IsSystem:             false,
		RequestedPermissions: []string{"android.permission.BIND_ACCESSIBILITY_SERVICE"},
	}
	findings = evaluator.EvaluateApp(negBNK1, nil)
	if hasFinding(findings, "BNK-001") {
		t.Errorf("expected BNK-001 negative when missing SYSTEM_ALERT_WINDOW")
	}

	// 8. PUA-001 (Sideloaded Hidden Notification Spammer)
	posPUA1 := inventory.AppInfo{
		Package:              "com.spam.hiddennotify",
		IsSystem:             false,
		HasLauncher:          false,
		Installer:            nil, // sideload
		RequestedPermissions: []string{"android.permission.POST_NOTIFICATIONS"},
	}
	findings = evaluator.EvaluateApp(posPUA1, nil)
	if !hasFinding(findings, "PUA-001") {
		t.Errorf("expected PUA-001 positive match")
	}

	negPUA1 := inventory.AppInfo{
		Package:              "com.store.legitnotify",
		IsSystem:             false,
		HasLauncher:          false,
		Installer:            &storeInstaller,
		RequestedPermissions: []string{"android.permission.POST_NOTIFICATIONS"},
	}
	findings = evaluator.EvaluateApp(negPUA1, nil)
	if hasFinding(findings, "PUA-001") {
		t.Errorf("expected PUA-001 negative when installed from Play Store")
	}
}

func TestQuickScanAndDeepScan(t *testing.T) {
	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()

	// 1. Setup mock device outputs
	fake.SetResponse("pm list packages", "package:/data/app/~~t/base.apk=com.test.adware\n", "", nil)
	fake.SetResponse("dumpsys package com.test.adware", `Package [com.test.adware]
userId=10200
versionName=1.0.0
versionCode=100
requested permissions:
  android.permission.SYSTEM_ALERT_WINDOW
`, "", nil)

	// Create fake puller output
	fake.SetResponse("pm path com.test.adware", "package:/data/app/~~t/base.apk\n", "", nil)

	// Build synthetic APK on disk for deep scan test
	apkPath := filepath.Join(tmpDir, "pulled_apks", "com.test.adware", "base.apk")
	_, _, err := apk.CreateSyntheticAPK(
		apkPath, "com.test.adware", "1.0.0", 100,
		[]string{"android.permission.SYSTEM_ALERT_WINDOW"},
		[]string{"com/unity3d/ads/UnityAds", "com/applovin/AppLovin"},
		"Jerion Test",
	)
	if err != nil {
		t.Fatalf("CreateSyntheticAPK failed: %v", err)
	}

	fake.SetHandler(
		func(serial string, args []string) bool {
			return len(args) >= 3 && args[0] == "pull"
		},
		func(serial string, args []string) ([]byte, []byte, error) {
			dest := args[2]
			_ = os.MkdirAll(filepath.Dir(dest), 0o750)
			data, _ := os.ReadFile(apkPath)
			_ = os.WriteFile(dest, data, 0o644)
			return []byte("1 file pulled"), nil, nil
		},
	)

	invSvc := inventory.NewService(fake, nil)
	puller := inventory.NewPuller(fake, nil)
	s, _ := store.OpenMemory()
	defer s.Close()

	scanner, err := NewScanner(invSvc, puller, nil, nil, s, tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	// Execute Quick Scan with canonical ScanID
	canonicalID := "scan-canonical-42"
	var progressRecords []ScanProgress
	quickResult, err := scanner.Scan(context.Background(), "DEV-TEST-01", ScanOptions{ScanID: canonicalID, Mode: "quick"}, func(p ScanProgress) {
		progressRecords = append(progressRecords, p)
	})
	if err != nil {
		t.Fatalf("QuickScan failed: %v", err)
	}

	if quickResult.ID != canonicalID {
		t.Errorf("expected canonical scan ID %s, got %s", canonicalID, quickResult.ID)
	}
	if quickResult.TotalApps != 1 {
		t.Errorf("expected TotalApps=1, got %d", quickResult.TotalApps)
	}

	if quickResult.Mode != "quick" || len(quickResult.Findings) == 0 {
		t.Errorf("expected quick scan findings, got: %+v", quickResult)
	}
	if len(progressRecords) == 0 {
		t.Errorf("expected progress callbacks")
	}

	// Execute Deep Scan
	deepResult, err := scanner.Scan(context.Background(), "DEV-TEST-01", ScanOptions{Mode: "deep"}, nil)
	if err != nil {
		t.Fatalf("DeepScan failed: %v", err)
	}
	if deepResult.Mode != "deep" {
		t.Errorf("expected deep scan mode, got %s", deepResult.Mode)
	}
}

func hasFinding(findings []Finding, ruleID string) bool {
	for _, f := range findings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}
