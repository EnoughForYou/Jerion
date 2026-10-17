package adb

import (
	"context"
	"strings"
	"testing"
)

func TestInputValidation(t *testing.T) {
	validPackages := []string{
		"com.android.settings",
		"com.infinix.xos.launcher",
		"org.example_app.test_123",
	}
	for _, pkg := range validPackages {
		if err := ValidatePackageName(pkg); err != nil {
			t.Errorf("expected valid package %q, got error: %v", pkg, err)
		}
	}

	invalidPackages := []string{
		"com.example; rm -rf /",
		"com.example | cat",
		"com.example`id`",
		"com.example$(whoami)",
		"com/example/path",
		"",
	}
	for _, pkg := range invalidPackages {
		if err := ValidatePackageName(pkg); err == nil {
			t.Errorf("expected rejection for injection package %q, got nil", pkg)
		}
	}

	validPermissions := []string{
		"android.permission.INTERNET",
		"android.permission.SYSTEM_ALERT_WINDOW",
		"com.transsion.permission.RECEIVE",
	}
	for _, p := range validPermissions {
		if err := ValidatePermissionName(p); err != nil {
			t.Errorf("expected valid permission %q, got: %v", p, err)
		}
	}

	invalidPermissions := []string{
		"android.permission; reboot",
		"android.permission/test",
		"",
	}
	for _, p := range invalidPermissions {
		if err := ValidatePermissionName(p); err == nil {
			t.Errorf("expected rejection for injection permission %q, got nil", p)
		}
	}
}

func TestFakeRunner(t *testing.T) {
	fake := NewFakeRunner()
	fake.SetResponse("getprop", "[ro.product.model]: [Infinix X698]\n", "", nil)

	ctx := context.Background()
	stdout, _, err := fake.Run(ctx, "07982241C0006769", "shell", "getprop")
	if err != nil {
		t.Fatalf("FakeRunner Run failed: %v", err)
	}
	if !strings.Contains(string(stdout), "Infinix X698") {
		t.Fatalf("unexpected stdout: %s", string(stdout))
	}

	calls := fake.Calls()
	if len(calls) != 1 || calls[0].Serial != "07982241C0006769" {
		t.Fatalf("unexpected calls record: %+v", calls)
	}
}

func TestParseDevices(t *testing.T) {
	fixture := `List of devices attached
07982241C0006769       device product:X698-GL model:Infinix_X698 device:Infinix-X698 transport_id:1
emulator-5554          unauthorized transport_id:2
192.168.1.50:5555      offline transport_id:3
08123456789            authorizing transport_id:4
`
	devs := ParseDevices(fixture)
	if len(devs) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(devs))
	}

	if devs[0].Serial != "07982241C0006769" || devs[0].State != StateReady || devs[0].Model != "Infinix_X698" {
		t.Errorf("unexpected dev 0: %+v", devs[0])
	}
	if devs[1].Serial != "emulator-5554" || devs[1].State != StateUnauthorized {
		t.Errorf("unexpected dev 1: %+v", devs[1])
	}
	if devs[2].Serial != "192.168.1.50:5555" || devs[2].State != StateOffline {
		t.Errorf("unexpected dev 2: %+v", devs[2])
	}
	if devs[3].Serial != "08123456789" || devs[3].State != StateUnauthorized {
		t.Errorf("unexpected dev 3 for authorizing state: %+v", devs[3])
	}
}

func TestParseGetprop(t *testing.T) {
	fixture := `
[ro.product.model]: [Infinix X698]
[ro.product.manufacturer]: [Infinix]
[ro.build.version.release]: [12]
[ro.build.version.sdk]: [31]
[ro.build.version.security_patch]: [2022-04-05]
[ro.build.fingerprint]: [Infinix/X698/Infinix-X698:12/SP1A.210812.016/220415V437:user/release-keys]
[ro.boot.verifiedbootstate]: [green]
`
	props := ParseGetprop(fixture)
	if props["ro.product.model"] != "Infinix X698" {
		t.Errorf("expected Infinix X698, got %s", props["ro.product.model"])
	}
	if props["ro.build.version.sdk"] != "31" {
		t.Errorf("expected 31, got %s", props["ro.build.version.sdk"])
	}
	if props["ro.boot.verifiedbootstate"] != "green" {
		t.Errorf("expected green, got %s", props["ro.boot.verifiedbootstate"])
	}
}

func TestParseBattery(t *testing.T) {
	fixture := `
Current Battery Service state:
  AC powered: false
  USB powered: true
  Wireless powered: false
  status: 2
  health: 2
  present: true
  level: 85
  scale: 100
  voltage: 4150
  temperature: 285
  technology: Li-poly
`
	b := ParseBattery(fixture)
	if b.Level != 85 || b.Percent != 85 {
		t.Errorf("expected level 85, got %d (percent %d)", b.Level, b.Percent)
	}
	if !b.IsCharging || b.Status != "Charging" {
		t.Errorf("expected charging, got status=%s, isCharging=%v", b.Status, b.IsCharging)
	}
	if b.Health != "Good" {
		t.Errorf("expected Good health, got %s", b.Health)
	}
	if b.Temperature != 28.5 {
		t.Errorf("expected 28.5 C, got %f", b.Temperature)
	}
}

func TestParseStorage(t *testing.T) {
	fixture := `
Filesystem     1K-blocks     Used Available Use% Mounted on
/dev/block/dm-5 116245084 45689124  70555960  40% /data
`
	s := ParseStorage(fixture)
	if s.UsePercent != 40 {
		t.Errorf("expected 40%% use, got %d", s.UsePercent)
	}
	expectedTotal := int64(116245084) * 1024
	if s.TotalBytes != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, s.TotalBytes)
	}
}

func TestParseWindowFocus(t *testing.T) {
	fixture := `
  mCurrentFocus=Window{8d1976a u0 com.example.adware/com.example.adware.MainActivity}
  mFocusedApp=ActivityRecord{3a498b u0 com.example.adware/com.example.adware.MainActivity t12}
`
	f := ParseWindowFocus(fixture)
	if f.Package != "com.example.adware" || f.Activity != "com.example.adware.MainActivity" {
		t.Errorf("unexpected window focus: %+v", f)
	}
}

func TestManagerVersionAndServer(t *testing.T) {
	fake := NewFakeRunner()
	fake.SetResponse("version", "Android Debug Bridge version 1.0.41\nVersion 34.0.5-10900870\n", "", nil)
	fake.SetResponse("start-server", "* daemon started successfully *\n", "", nil)

	mgr := NewManager(fake, "", nil)

	ver, err := mgr.CheckVersion(context.Background())
	if err != nil || ver != "1.0.41" {
		t.Fatalf("unexpected CheckVersion: ver=%s, err=%v", ver, err)
	}

	if err := mgr.StartServer(context.Background()); err != nil {
		t.Fatalf("StartServer failed: %v", err)
	}
}

func TestTrackerPollAndEvents(t *testing.T) {
	fake := NewFakeRunner()
	fake.SetResponse("devices -l", `List of devices attached
07982241C0006769       device model:Infinix_X698 transport_id:1
`, "", nil)

	tracker := NewTracker(fake, nil)
	var changeCount int
	tracker.AddListener(func(devices []DeviceInfo, active *DeviceInfo) {
		changeCount++
	})

	devs, err := tracker.PollOnce(context.Background())
	if err != nil || len(devs) != 1 {
		t.Fatalf("PollOnce failed: len=%d, err=%v", len(devs), err)
	}
	if changeCount != 1 {
		t.Fatalf("expected 1 change notification, got %d", changeCount)
	}

	active := tracker.ActiveDevice()
	if active == nil || active.Serial != "07982241C0006769" {
		t.Fatalf("unexpected active device: %+v", active)
	}

	// Poll again with same output -> no change
	_, _ = tracker.PollOnce(context.Background())
	if changeCount != 1 {
		t.Fatalf("expected no notification for identical devices state, got %d", changeCount)
	}
}

func TestManager_SyncRunnerPath(t *testing.T) {
	fake := NewFakeRunner()
	mgr := NewManager(fake, "", nil)

	customPath := "C:\\custom\\platform-tools\\adb.exe"
	mgr.SetCustomPath(customPath)

	if fake.ADBPath() != customPath {
		t.Fatalf("expected fake.ADBPath() to be %q, got %q", customPath, fake.ADBPath())
	}
}

func TestManager_WirelessDebugging(t *testing.T) {
	ctx := context.Background()
	fake := NewFakeRunner()
	mgr := NewManager(fake, "", nil)

	// 1. Pair Wireless validation
	if err := mgr.PairWireless(ctx, "invalid host space", 37123, "123456"); err == nil {
		t.Errorf("expected error on invalid host")
	}
	if err := mgr.PairWireless(ctx, "192.168.1.50", 0, "123456"); err == nil {
		t.Errorf("expected error on port 0")
	}
	if err := mgr.PairWireless(ctx, "192.168.1.50", 70000, "123456"); err == nil {
		t.Errorf("expected error on port > 65535")
	}
	if err := mgr.PairWireless(ctx, "192.168.1.50", 37123, "12345"); err == nil {
		t.Errorf("expected error on 5-digit code")
	}
	if err := mgr.PairWireless(ctx, "192.168.1.50", 37123, "12345a"); err == nil {
		t.Errorf("expected error on non-numeric code")
	}

	// 2. Pair Wireless Success
	fake.SetResponse("pair 192.168.1.50:37123 654321", "Successfully paired to 192.168.1.50:37123 [guid: ...]\n", "", nil)
	if err := mgr.PairWireless(ctx, "192.168.1.50", 37123, "654321"); err != nil {
		t.Fatalf("PairWireless failed: %v", err)
	}

	// 3. Pair Wireless Rejection
	fake.SetResponse("pair 192.168.1.50:37124 111111", "Failed: Wrong password or connection was reset\n", "", nil)
	if err := mgr.PairWireless(ctx, "192.168.1.50", 37124, "111111"); err == nil {
		t.Fatalf("expected error on rejected pairing")
	}

	// 4. Connect Wireless Validation
	if _, err := mgr.ConnectWireless(ctx, "bad host!", 5555); err == nil {
		t.Errorf("expected error on invalid host")
	}
	if _, err := mgr.ConnectWireless(ctx, "192.168.1.50", 80000); err == nil {
		t.Errorf("expected error on invalid port")
	}

	// 5. Connect Wireless Success
	fake.SetResponse("connect 192.168.1.50:5555", "connected to 192.168.1.50:5555\n", "", nil)
	endpoint, err := mgr.ConnectWireless(ctx, "192.168.1.50", 5555)
	if err != nil || endpoint != "192.168.1.50:5555" {
		t.Fatalf("ConnectWireless failed: endpoint=%s, err=%v", endpoint, err)
	}

	// 6. Connect Wireless Rejection
	fake.SetResponse("connect 192.168.1.50:5556", "cannot connect to 192.168.1.50:5556: No connection could be made\n", "", nil)
	if _, err := mgr.ConnectWireless(ctx, "192.168.1.50", 5556); err == nil {
		t.Fatalf("expected error on rejected connection")
	}

	// 7. Disconnect Wireless
	fake.SetResponse("disconnect 192.168.1.50:5555", "disconnected 192.168.1.50:5555\n", "", nil)
	if err := mgr.DisconnectWireless(ctx, "192.168.1.50", 5555); err != nil {
		t.Fatalf("DisconnectWireless failed: %v", err)
	}
}

func TestManager_SoftReconnect(t *testing.T) {
	ctx := context.Background()
	fake := NewFakeRunner()
	mgr := NewManager(fake, "", nil)

	// Invalid serial
	if err := mgr.SoftReconnect(ctx, "invalid serial with spaces!!"); err == nil {
		t.Errorf("expected error on invalid serial")
	}

	// Reconnect global
	fake.SetResponse("reconnect", "reconnecting\n", "", nil)
	if err := mgr.SoftReconnect(ctx, ""); err != nil {
		t.Fatalf("SoftReconnect failed: %v", err)
	}

	// Reconnect specific serial
	fake.SetResponse("reconnect", "reconnecting 07982241C0006769\n", "", nil)
	if err := mgr.SoftReconnect(ctx, "07982241C0006769"); err != nil {
		t.Fatalf("SoftReconnect with serial failed: %v", err)
	}
}

func TestSanitizeArgsForLog(t *testing.T) {
	// Pair command with pairing code
	args := []string{"pair", "192.168.1.50:37123", "654321"}
	sanitized := sanitizeArgsForLog(args)
	if sanitized[2] != "[REDACTED]" {
		t.Errorf("expected pairing code to be [REDACTED], got %s", sanitized[2])
	}
	if args[2] != "654321" {
		t.Errorf("original args should not be mutated in-place")
	}

	// Normal command
	normal := []string{"devices", "-l"}
	sanNormal := sanitizeArgsForLog(normal)
	if len(sanNormal) != 2 || sanNormal[0] != "devices" || sanNormal[1] != "-l" {
		t.Errorf("unexpected normal command sanitization: %v", sanNormal)
	}
}
