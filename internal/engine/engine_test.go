package engine

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/apk"
	"github.com/EnoughForYou/Jerion/internal/intel"
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

func startMockClamd(t *testing.T, handler func(net.Conn)) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock clamd listener: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			go handler(conn)
		}
	}()

	cleanup := func() {
		close(done)
		_ = listener.Close()
	}

	return listener.Addr().String(), cleanup
}

func TestDeepScan_ClamAV_MalwareDetected(t *testing.T) {
	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()

	fake.SetResponse("pm list packages", "package:/data/app/~~t/base.apk=com.test.malware\n", "", nil)
	fake.SetResponse("dumpsys package com.test.malware", `Package [com.test.malware]
userId=10201
versionName=1.0.0
versionCode=100
`, "", nil)
	fake.SetResponse("pm path com.test.malware", "package:/data/app/~~t/base.apk\n", "", nil)

	apkPath := filepath.Join(tmpDir, "pulled_apks", "com.test.malware", "base.apk")
	_, _, err := apk.CreateSyntheticAPK(
		apkPath, "com.test.malware", "1.0.0", 100,
		[]string{"android.permission.INTERNET"},
		nil,
		"Malware Test",
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

	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			// When terminal chunk (4 zero bytes) is received, respond FOUND
			if n >= 4 && buf[n-4] == 0 && buf[n-3] == 0 && buf[n-2] == 0 && buf[n-1] == 0 {
				_, _ = conn.Write([]byte("stream: Win.Test.EICAR_HDB-1 FOUND\x00"))
				return
			}
		}
	})
	defer cleanup()

	s, _ := store.OpenMemory()
	defer s.Close()

	agg := intel.NewAggregator(s, nil)
	clamClient := intel.NewClamAVClient(addr)
	agg.SetClamAVClient(clamClient)

	invSvc := inventory.NewService(fake, nil)
	puller := inventory.NewPuller(fake, nil)

	scanner, err := NewScanner(invSvc, puller, nil, agg, s, tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	res, err := scanner.Scan(context.Background(), "DEV-TEST-CLAM", ScanOptions{Mode: "deep"}, nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !hasFinding(res.Findings, "CLAMAV-SIGNATURE-001") {
		t.Fatalf("expected CLAMAV-SIGNATURE-001 finding, got: %+v", res.Findings)
	}

	var clamFinding *Finding
	for i := range res.Findings {
		if res.Findings[i].RuleID == "CLAMAV-SIGNATURE-001" {
			clamFinding = &res.Findings[i]
			break
		}
	}
	if clamFinding == nil {
		t.Fatalf("missing clamav finding")
	}
	if clamFinding.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", clamFinding.Severity)
	}
	if clamFinding.Category != "malware" {
		t.Errorf("expected category malware, got %s", clamFinding.Category)
	}
	if len(clamFinding.Actions) == 0 || clamFinding.Actions[0] != "uninstall" {
		t.Errorf("expected uninstall action, got %v", clamFinding.Actions)
	}
}

func TestDeepScan_ClamAV_Error_MarksPartial(t *testing.T) {
	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()

	fake.SetResponse("pm list packages", "package:/data/app/~~t/base.apk=com.test.broken1\npackage:/data/app/~~t/base.apk=com.test.broken2\n", "", nil)
	fake.SetResponse("dumpsys package com.test.broken1", `Package [com.test.broken1]
userId=10202
versionName=1.0.0
versionCode=100
`, "", nil)
	fake.SetResponse("dumpsys package com.test.broken2", `Package [com.test.broken2]
userId=10203
versionName=1.0.0
versionCode=101
`, "", nil)
	fake.SetResponse("pm path com.test.broken1", "package:/data/app/~~t/base.apk\n", "", nil)
	fake.SetResponse("pm path com.test.broken2", "package:/data/app/~~t/base.apk\n", "", nil)

	apkPath1 := filepath.Join(tmpDir, "pulled_apks", "com.test.broken1", "base.apk")
	_, _, err := apk.CreateSyntheticAPK(
		apkPath1, "com.test.broken1", "1.0.0", 100,
		[]string{"android.permission.INTERNET"},
		nil,
		"Broken Test 1",
	)
	if err != nil {
		t.Fatalf("CreateSyntheticAPK 1 failed: %v", err)
	}

	apkPath2 := filepath.Join(tmpDir, "pulled_apks", "com.test.broken2", "base.apk")
	_, _, err = apk.CreateSyntheticAPK(
		apkPath2, "com.test.broken2", "1.0.0", 101,
		[]string{"android.permission.INTERNET"},
		nil,
		"Broken Test 2",
	)
	if err != nil {
		t.Fatalf("CreateSyntheticAPK 2 failed: %v", err)
	}

	fake.SetHandler(
		func(serial string, args []string) bool {
			return len(args) >= 3 && args[0] == "pull"
		},
		func(serial string, args []string) ([]byte, []byte, error) {
			dest := args[2]
			_ = os.MkdirAll(filepath.Dir(dest), 0o750)
			srcPath := apkPath1
			if strings.Contains(dest, "com.test.broken2") {
				srcPath = apkPath2
			}
			data, _ := os.ReadFile(srcPath)
			_ = os.WriteFile(dest, data, 0o644)
			return []byte("1 file pulled"), nil, nil
		},
	)

	// Mock server closes immediately on connection (triggering error)
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		_ = conn.Close()
	})
	defer cleanup()

	s, _ := store.OpenMemory()
	defer s.Close()

	agg := intel.NewAggregator(s, nil)
	clamClient := intel.NewClamAVClient(addr)
	agg.SetClamAVClient(clamClient)

	invSvc := inventory.NewService(fake, nil)
	puller := inventory.NewPuller(fake, nil)

	scanner, err := NewScanner(invSvc, puller, nil, agg, s, tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	res, err := scanner.Scan(context.Background(), "DEV-TEST-BROKEN", ScanOptions{Mode: "deep"}, nil)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	clamErrCount := 0
	for _, f := range res.Findings {
		if f.RuleID == "CLAMAV-SCAN-ERR" {
			clamErrCount++
			if f.Package != "" {
				t.Errorf("expected scan-level finding with empty package, got %q", f.Package)
			}
		}
	}
	if clamErrCount != 1 {
		t.Fatalf("expected exactly 1 deduplicated CLAMAV-SCAN-ERR finding, got %d", clamErrCount)
	}
	if !res.Partial {
		t.Fatalf("expected scan result Partial to be true when ClamAV fails")
	}
}

