package inventory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

func TestParsePMList(t *testing.T) {
	fixture := `package:/data/app/~~a8-vQ==/com.infinix.xos.launcher-1/base.apk=com.infinix.xos.launcher
package:/data/app/~~k9-pW==/com.example.adware-2/base.apk=com.example.adware
package:/system/priv-app/Settings/Settings.apk=com.android.settings
`
	entries := ParsePMList(fixture)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].Package != "com.infinix.xos.launcher" || entries[0].CodePath != "/data/app/~~a8-vQ==/com.infinix.xos.launcher-1/base.apk" {
		t.Errorf("unexpected entry 0: %+v", entries[0])
	}
	if entries[1].Package != "com.example.adware" {
		t.Errorf("unexpected entry 1: %+v", entries[1])
	}
	if entries[2].Package != "com.android.settings" {
		t.Errorf("unexpected entry 2: %+v", entries[2])
	}
}

func TestParseDumpsysPackage(t *testing.T) {
	fixture := `
Packages:
  Package [com.cleaner.master] (8d1b32e):
    userId=10185
    sharedUser=null
    pkg=Package{2f1590 com.cleaner.master}
    codePath=/data/app/~~xY12==/com.cleaner.master-ABC==
    resourcePath=/data/app/~~xY12==/com.cleaner.master-ABC==
    versionCode=20401 minSdk=24 targetSdk=31
    versionName=2.4.1
    splits=[config.arm64_v8a, config.xxhdpi]
    installerPackageName=com.android.vending
    signatures=PackageSignatures{7b1892 [Signer #1 certificate SHA-256 digest: A1B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0]}
    installPermissionsFixed=true
    pkgFlags=[ HAS_CODE ALLOW_CLEAR_USER_DATA ]
    declared permissions:
    requested permissions:
      android.permission.INTERNET
      android.permission.ACCESS_NETWORK_STATE
      android.permission.SYSTEM_ALERT_WINDOW
      android.permission.WAKE_LOCK
    install permissions:
      android.permission.INTERNET: granted=true
      android.permission.ACCESS_NETWORK_STATE: granted=true
    runtime permissions:
      android.permission.SYSTEM_ALERT_WINDOW: granted=true
    firstInstallTime=2022-04-15 14:22:10
    lastUpdateTime=2022-05-20 09:15:30
    enabled=0
    Activity Resolver Table:
      Non-Data Actions:
        android.intent.action.MAIN:
          8d1976a com.cleaner.master/.MainActivity filter 3a28bc
            Action: "android.intent.action.MAIN"
            Category: "android.intent.category.LAUNCHER"
`
	info := ParseDumpsysPackage(fixture)
	if info == nil {
		t.Fatal("expected non-nil AppInfo")
	}

	if info.Package != "com.cleaner.master" {
		t.Errorf("expected com.cleaner.master, got %s", info.Package)
	}
	if info.UID != 10185 {
		t.Errorf("expected UID 10185, got %d", info.UID)
	}
	if info.VersionCode != 20401 {
		t.Errorf("expected VersionCode 20401, got %d", info.VersionCode)
	}
	if info.VersionName != "2.4.1" {
		t.Errorf("expected VersionName 2.4.1, got %s", info.VersionName)
	}
	if info.TargetSDK != 31 {
		t.Errorf("expected TargetSDK 31, got %d", info.TargetSDK)
	}
	if info.MinSDK != 24 {
		t.Errorf("expected MinSDK 24, got %d", info.MinSDK)
	}
	if info.Installer == nil || *info.Installer != "com.android.vending" {
		t.Errorf("expected installer com.android.vending, got %v", info.Installer)
	}
	if !info.HasLauncher {
		t.Errorf("expected HasLauncher=true")
	}
	if len(info.RequestedPermissions) != 4 {
		t.Errorf("expected 4 requested perms, got %d", len(info.RequestedPermissions))
	}
	if len(info.GrantedPermissions) != 3 {
		t.Errorf("expected 3 granted perms, got %d: %v", len(info.GrantedPermissions), info.GrantedPermissions)
	}
	if len(info.CertSHA256) != 1 || info.CertSHA256[0] != "A1B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0" {
		t.Errorf("unexpected cert: %v", info.CertSHA256)
	}
}

func TestPuller(t *testing.T) {
	fake := adb.NewFakeRunner()
	tmpDir := t.TempDir()

	// Simulate pm path output
	fake.SetResponse("pm path", "package:/data/app/~~pkg/base.apk\n", "", nil)

	// Create a dummy file that adb pull would produce
	fake.SetHandler(
		func(serial string, args []string) bool {
			return len(args) >= 3 && args[0] == "pull"
		},
		func(serial string, args []string) ([]byte, []byte, error) {
			dest := args[2]
			_ = os.WriteFile(dest, []byte("fake-apk-content"), 0o644)
			return []byte("1 file pulled"), nil, nil
		},
	)

	puller := NewPuller(fake, nil)
	basePath, allPaths, shaHex, err := puller.Pull(context.Background(), "DEV-01", "com.cleaner.master", tmpDir)
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	if basePath != filepath.Join(tmpDir, "com.cleaner.master", "base.apk") {
		t.Errorf("unexpected basePath: %s", basePath)
	}
	if len(allPaths) != 1 {
		t.Errorf("expected 1 pulled path, got %d", len(allPaths))
	}
	if shaHex == "" {
		t.Errorf("expected non-empty sha256")
	}
}

func TestInventoryService(t *testing.T) {
	fake := adb.NewFakeRunner()
	fake.SetResponse("pm list packages", "package:/data/app/test/base.apk=com.test.app\n", "", nil)
	fake.SetResponse("dumpsys package", `Package [com.test.app]
userId=10099
versionName=1.0.0
versionCode=10
`, "", nil)

	svc := NewService(fake, nil)

	apps, err := svc.ListPackages(context.Background(), "DEV-01", true)
	if err != nil || len(apps) != 1 {
		t.Fatalf("ListPackages failed: len=%d, err=%v", len(apps), err)
	}

	detail, err := svc.GetPackageDetail(context.Background(), "DEV-01", "com.test.app")
	if err != nil || detail.UID != 10099 {
		t.Fatalf("GetPackageDetail failed: %+v, err=%v", detail, err)
	}
}

func TestParseDeviceAdmins(t *testing.T) {
	fixture := `Current Device Policy Manager state:
  Enabled Device Admins (user 0):
    com.cleaner.admin/.AdminReceiver:
      uid=10185
    admin=com.mdm.enterprise/.EnterpriseAdmin
  User 0: admin=com.google.android.apps.work.clouddpc/.receivers.CloudDeviceAdminReceiver,name=Work Profile
`
	admins := ParseDeviceAdmins(fixture)
	if !admins["com.cleaner.admin"] {
		t.Errorf("expected com.cleaner.admin to be detected as device admin")
	}
	if !admins["com.mdm.enterprise"] {
		t.Errorf("expected com.mdm.enterprise to be detected as device admin")
	}
	if !admins["com.google.android.apps.work.clouddpc"] {
		t.Errorf("expected com.google.android.apps.work.clouddpc to be detected as device admin")
	}
	if admins["com.unrelated.package"] {
		t.Errorf("unexpected package detected as admin")
	}
}

func TestParseAppOpsOverlay(t *testing.T) {
	fixture := `Current AppOps Service state:
  Package com.overlay.active:
    SYSTEM_ALERT_WINDOW: allow; time=+12s ago; duration=+5s
  Package com.overlay.denied:
    SYSTEM_ALERT_WINDOW: ignore; time=+2m ago
  Package com.overlay.default:
    SYSTEM_ALERT_WINDOW: default
  Package com.overlay.mode0:
    SYSTEM_ALERT_WINDOW (allow): mode=0
`
	overlays := ParseAppOpsOverlay(fixture)
	if !overlays["com.overlay.active"] {
		t.Errorf("expected com.overlay.active to be detected as active overlay")
	}
	if !overlays["com.overlay.mode0"] {
		t.Errorf("expected com.overlay.mode0 to be detected as active overlay")
	}
	if overlays["com.overlay.denied"] {
		t.Errorf("com.overlay.denied should NOT be detected as active overlay")
	}
	if overlays["com.overlay.default"] {
		t.Errorf("com.overlay.default should NOT be detected as active overlay")
	}

	// Single check
	if !ParseSingleAppOpOverlay("SYSTEM_ALERT_WINDOW: allow") {
		t.Errorf("expected single allow to be parsed as true")
	}
	if !ParseSingleAppOpOverlay("allow") {
		t.Errorf("expected single allow word to be parsed as true")
	}
	if ParseSingleAppOpOverlay("SYSTEM_ALERT_WINDOW: ignore") {
		t.Errorf("expected single ignore to be parsed as false")
	}
	if ParseSingleAppOpOverlay("No operations.") {
		t.Errorf("expected No operations to be parsed as false")
	}
}

