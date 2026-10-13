package apk

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestAXMLParser(t *testing.T) {
	pkg := "com.synthetic.adware"
	versionName := "3.2.1"
	versionCode := int64(321)
	permissions := []string{
		"android.permission.INTERNET",
		"android.permission.SYSTEM_ALERT_WINDOW",
	}

	axmlData := GenerateSyntheticAXML(pkg, versionName, versionCode, permissions)
	manifest, err := ParseAXML(bytes.NewReader(axmlData))
	if err != nil {
		t.Fatalf("ParseAXML failed: %v", err)
	}

	if manifest.Package != pkg {
		t.Errorf("expected package %s, got %s", pkg, manifest.Package)
	}
	if manifest.VersionName != versionName {
		t.Errorf("expected versionName %s, got %s", versionName, manifest.VersionName)
	}
	if manifest.VersionCode != versionCode {
		t.Errorf("expected versionCode %d, got %d", versionCode, manifest.VersionCode)
	}
	if !manifest.HasOverlay {
		t.Errorf("expected HasOverlay=true")
	}
	if len(manifest.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(manifest.Permissions))
	}
}

func TestDEXParserAndAdSDKDetection(t *testing.T) {
	strings := []string{
		"Lcom/example/MainActivity;",
		"com/unity3d/ads/UnityAds",
		"com/applovin/sdk/AppLovinSdk",
		"onCreate",
	}

	dexData := GenerateSyntheticDEX(strings)
	analysis, err := ParseDEX(bytes.NewReader(dexData))
	if err != nil {
		t.Fatalf("ParseDEX failed: %v", err)
	}

	if analysis.TotalStrings != 4 {
		t.Errorf("expected 4 strings, got %d", analysis.TotalStrings)
	}
	if len(analysis.DetectedAdSDKs) != 2 {
		t.Errorf("expected 2 detected ad SDKs, got %d: %+v", len(analysis.DetectedAdSDKs), analysis.DetectedAdSDKs)
	}
	if analysis.DetectedAdSDKs["com/unity3d/ads"] != "Unity Ads" {
		t.Errorf("expected Unity Ads, got %s", analysis.DetectedAdSDKs["com/unity3d/ads"])
	}
	if analysis.DetectedAdSDKs["com/applovin"] != "AppLovin MAX" {
		t.Errorf("expected AppLovin MAX, got %s", analysis.DetectedAdSDKs["com/applovin"])
	}
}

func TestFullAPKInspection(t *testing.T) {
	tmpDir := t.TempDir()
	apkPath := filepath.Join(tmpDir, "sample_test.apk")

	pkg := "com.adware.synthetic"
	versionName := "1.0.0"
	versionCode := int64(100)
	perms := []string{"android.permission.SYSTEM_ALERT_WINDOW"}
	dexStrings := []string{
		"com/google/android/gms/ads/MobileAds",
		"com/mbridge/msdk/MBridgeSDK",
	}

	certFP, fileSHA, err := CreateSyntheticAPK(apkPath, pkg, versionName, versionCode, perms, dexStrings, "Jerion Test Org")
	if err != nil {
		t.Fatalf("CreateSyntheticAPK failed: %v", err)
	}

	inspector := NewInspector()
	detail, err := inspector.Inspect(apkPath)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	if detail.SHA256 != fileSHA {
		t.Errorf("expected file SHA %s, got %s", fileSHA, detail.SHA256)
	}
	if detail.Manifest.Package != pkg {
		t.Errorf("expected package %s, got %s", pkg, detail.Manifest.Package)
	}
	if !detail.Manifest.HasOverlay {
		t.Errorf("expected HasOverlay=true")
	}
	if len(detail.CertFingerprints) == 0 || detail.CertFingerprints[0] != certFP {
		t.Errorf("expected cert fingerprint %s, got %v", certFP, detail.CertFingerprints)
	}
	if len(detail.DetectedAdSDKs) != 2 {
		t.Errorf("expected 2 ad SDKs detected, got %d: %+v", len(detail.DetectedAdSDKs), detail.DetectedAdSDKs)
	}
}
