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
	overlays, ok := ParseAppOpsOverlay(fixture)
	if !ok {
		t.Fatalf("expected valid fixture to be recognized as valid AppOps output")
	}
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

	// Recognized AppOps with no active overlays
	emptyValidFixture := `Current AppOps Service state:
  No operations.
`
	emptyOverlays, emptyOk := ParseAppOpsOverlay(emptyValidFixture)
	if !emptyOk {
		t.Errorf("expected empty valid fixture to be recognized")
	}
	if len(emptyOverlays) != 0 {
		t.Errorf("expected 0 active overlays, got %d", len(emptyOverlays))
	}

	// Unrecognized / Error outputs
	if _, unrecOk := ParseAppOpsOverlay("Segmentation fault"); unrecOk {
		t.Errorf("expected segmentation fault to be unrecognized")
	}
	if _, unrecOk := ParseAppOpsOverlay("Can't find service: appops"); unrecOk {
		t.Errorf("expected service error to be unrecognized")
	}
	if _, unrecOk := ParseAppOpsOverlay("   \n  \t "); unrecOk {
		t.Errorf("expected empty string to be unrecognized")
	}
	if _, unrecOk := ParseAppOpsOverlay("Some random unexpected text with no AppOps signatures"); unrecOk {
		t.Errorf("expected random text to be unrecognized")
	}

	// Single check: positive (allowed, recognized)
	if allowed, rec := ParseSingleAppOpOverlay("SYSTEM_ALERT_WINDOW: allow"); !allowed || !rec {
		t.Errorf("expected single allow to be allowed=true, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay("allow"); !allowed || !rec {
		t.Errorf("expected single allow word to be allowed=true, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay("mode=0"); !allowed || !rec {
		t.Errorf("expected mode=0 to be allowed=true, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}

	// Single check: valid negative (not allowed, recognized)
	if allowed, rec := ParseSingleAppOpOverlay("SYSTEM_ALERT_WINDOW: ignore"); allowed || !rec {
		t.Errorf("expected single ignore to be allowed=false, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay("No operations."); allowed || !rec {
		t.Errorf("expected No operations to be allowed=false, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay("mode=default"); allowed || !rec {
		t.Errorf("expected mode=default to be allowed=false, recognized=true; got allowed=%v, rec=%v", allowed, rec)
	}

	// Single check: unparseable / error (not allowed, unrecognized)
	if allowed, rec := ParseSingleAppOpOverlay("Error: could not contact AppOps"); allowed || rec {
		t.Errorf("expected error message to be recognized=false; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay("unexpected random output"); allowed || rec {
		t.Errorf("expected random text to be recognized=false; got allowed=%v, rec=%v", allowed, rec)
	}
	if allowed, rec := ParseSingleAppOpOverlay(""); allowed || rec {
		t.Errorf("expected empty string to be recognized=false; got allowed=%v, rec=%v", allowed, rec)
	}

	// Dumpsys appops with bracketed UID format (e.g. Package com.test.threat [10001]:)
	bracketOutput := `Current AppOps Service state:
  Package com.test.bracketed [10001]:
    SYSTEM_ALERT_WINDOW: allow; time=+2s ago
  Package com.test.parenthesis (uid 10002):
    SYSTEM_ALERT_WINDOW: allow
`
	bMap, bOk := ParseAppOpsOverlay(bracketOutput)
	if !bOk {
		t.Errorf("expected bracketed dumpsys appops output to be recognized")
	}
	if !bMap["com.test.bracketed"] {
		t.Errorf("expected com.test.bracketed to be parsed as allowed=true")
	}
	if !bMap["com.test.parenthesis"] {
		t.Errorf("expected com.test.parenthesis to be parsed as allowed=true")
	}

	// Lowercase package header format (package com.test.lowercase:)
	lowerOutput := `appops service state:
  package com.test.lowercase:
    SYSTEM_ALERT_WINDOW: allow
`
	lMap, lOk := ParseAppOpsOverlay(lowerOutput)
	if !lOk {
		t.Errorf("expected lowercase dumpsys appops output to be recognized")
	}
	if !lMap["com.test.lowercase"] {
		t.Errorf("expected com.test.lowercase to be parsed as allowed=true")
	}
}

func TestParseDumpsysPackage_UpdatedSystemApp(t *testing.T) {
	fixture := `
Packages:
  Package [com.google.android.youtube] (7a2c10b):
    userId=10045
    pkg=Package{1b2c3d com.google.android.youtube}
    codePath=/data/app/~~a8-vQ==/com.google.android.youtube-1/base.apk
    resourcePath=/data/app/~~a8-vQ==/com.google.android.youtube-1/base.apk
    versionCode=1520032800 minSdk=26 targetSdk=34
    versionName=19.05.36
    pkgFlags=[ HAS_CODE ALLOW_CLEAR_USER_DATA ]
    pkgPrivateFlags=[ UPDATED_SYSTEM_APP ]
    firstInstallTime=2021-01-01 12:00:00
    lastUpdateTime=2024-02-15 10:30:00
    enabled=0
`
	info := ParseDumpsysPackage(fixture)
	if info == nil {
		t.Fatal("expected non-nil AppInfo")
	}
	if !info.IsSystem {
		t.Errorf("expected IsSystem=true for updated system app with UPDATED_SYSTEM_APP flag")
	}
	if info.Package != "com.google.android.youtube" {
		t.Errorf("expected com.google.android.youtube, got %s", info.Package)
	}
}

func TestParseDumpsysPackage_FlagSubstrings(t *testing.T) {
	// 1. User app with REQUIRED_FOR_SYSTEM_USER in private flags must NOT be marked as system app
	userFixture := `
Packages:
  Package [com.user.taskmanager]:
    userId=10190
    pkgFlags=[ HAS_CODE ALLOW_CLEAR_USER_DATA ]
    pkgPrivateFlags=[ REQUIRED_FOR_SYSTEM_USER ]
`
	userInfo := ParseDumpsysPackage(userFixture)
	if userInfo == nil {
		t.Fatal("expected non-nil AppInfo")
	}
	if userInfo.IsSystem {
		t.Errorf("expected IsSystem=false for user app with REQUIRED_FOR_SYSTEM_USER flag, got true")
	}

	// 2. System app with SYSTEM in flags must be marked as system app
	sysFixture := `
Packages:
  Package [com.android.systemui]:
    userId=1000
    pkgFlags=[ SYSTEM HAS_CODE ]
`
	sysInfo := ParseDumpsysPackage(sysFixture)
	if sysInfo == nil {
		t.Fatal("expected non-nil AppInfo")
	}
	if !sysInfo.IsSystem {
		t.Errorf("expected IsSystem=true for system app with SYSTEM flag")
	}
}

func TestIsSystemPackage(t *testing.T) {
	fake := adb.NewFakeRunner()
	svc := NewService(fake, nil)
	ctx := context.Background()

	// 1. Updated system app with /data/app path but listed in pm list packages -s
	fake.SetResponse("pm list packages -s com.google.android.youtube", "package:com.google.android.youtube\n", "", nil)
	fake.SetResponse("pm path com.google.android.youtube", "package:/data/app/~~abc/com.google.android.youtube-1/base.apk\n", "", nil)

	if !svc.IsSystemPackage(ctx, "dev-123", "com.google.android.youtube") {
		t.Errorf("expected updated system app with pm list -s match to be recognized as system app")
	}

	// 2. System app detected via dumpsys package flags (pm list -s returned other package or empty)
	fake.SetResponse("pm list packages -s com.android.settings", "package:com.other.app\n", "", nil)
	dumpsysSettings := `
Packages:
  Package [com.android.settings]:
    pkgFlags=[ SYSTEM HAS_CODE ]
    codePath=/system/priv-app/Settings/Settings.apk
`
	fake.SetResponse("dumpsys package com.android.settings", dumpsysSettings, "", nil)
	if !svc.IsSystemPackage(ctx, "dev-123", "com.android.settings") {
		t.Errorf("expected com.android.settings to be recognized as system app via dumpsys flags")
	}

	// 3. Fallback: recognized via /system/ codePath
	fake.SetResponse("pm list packages -s com.oem.telephony", "", "", nil)
	fake.SetResponse("dumpsys package com.oem.telephony", "Package [com.oem.telephony]:\n  userId=1000\n", "", nil)
	fake.SetResponse("pm path com.oem.telephony", "package:/system/priv-app/TelephonyProvider.apk\n", "", nil)
	if !svc.IsSystemPackage(ctx, "dev-123", "com.oem.telephony") {
		t.Errorf("expected com.oem.telephony to be recognized as system app via fallback path")
	}

	// 4. Fallback: recognized via /system_ext/ or /apex/ codePath
	fake.SetResponse("pm list packages -s com.android.apex.service", "", "", nil)
	fake.SetResponse("dumpsys package com.android.apex.service", "Package [com.android.apex.service]:\n  userId=1000\n", "", nil)
	fake.SetResponse("pm path com.android.apex.service", "package:/system_ext/app/Service.apk\n", "", nil)
	if !svc.IsSystemPackage(ctx, "dev-123", "com.android.apex.service") {
		t.Errorf("expected com.android.apex.service to be recognized as system app via system_ext path")
	}

	// 5. Regular 3rd-party user app with /data/app/ path even if package name has "system"
	fake.SetResponse("pm list packages -s com.system.cleaner", "", "", nil)
	dumpsysCleaner := `
Packages:
  Package [com.system.cleaner]:
    pkgFlags=[ HAS_CODE ]
    codePath=/data/app/com.system.cleaner/base.apk
`
	fake.SetResponse("dumpsys package com.system.cleaner", dumpsysCleaner, "", nil)
	fake.SetResponse("pm path com.system.cleaner", "package:/data/app/~~xyz/com.system.cleaner-1/base.apk\n", "", nil)
	if svc.IsSystemPackage(ctx, "dev-123", "com.system.cleaner") {
		t.Errorf("expected user app com.system.cleaner to NOT be recognized as system app")
	}

	// 6. Regular 3rd-party user app with /data/app/ path
	fake.SetResponse("pm list packages -s com.example.adware", "", "", nil)
	dumpsysAdware := `
Packages:
  Package [com.example.adware]:
    pkgFlags=[ HAS_CODE ]
    codePath=/data/app/com.example.adware/base.apk
`
	fake.SetResponse("dumpsys package com.example.adware", dumpsysAdware, "", nil)
	fake.SetResponse("pm path com.example.adware", "package:/data/app/com.example.adware/base.apk\n", "", nil)
	if svc.IsSystemPackage(ctx, "dev-123", "com.example.adware") {
		t.Errorf("expected user app com.example.adware to NOT be recognized as system app")
	}

	// 7. Edge cases: empty/nil/invalid package names
	if svc.IsSystemPackage(ctx, "", "com.example.adware") {
		t.Errorf("expected false for empty serial")
	}
	if svc.IsSystemPackage(ctx, "dev-123", "") {
		t.Errorf("expected false for empty package")
	}
	if svc.IsSystemPackage(ctx, "dev-123", "com.evil;rm -rf /") {
		t.Errorf("expected false for invalid package name with shell characters")
	}
	if svc.IsSystemPackage(ctx, "dev-123", "bad package name with spaces") {
		t.Errorf("expected false for invalid package name with spaces")
	}
	var nilSvc *Service
	if nilSvc.IsSystemPackage(ctx, "dev-123", "com.example.adware") {
		t.Errorf("expected false for nil service")
	}
}


