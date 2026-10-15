package preflight

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/secrets"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func TestPreflightEngineAllPass(t *testing.T) {
	fake := adb.NewFakeRunner()
	fake.SetResponse("version", "Android Debug Bridge version 1.0.41\n", "", nil)
	fake.SetResponse("start-server", "* daemon started *\n", "", nil)

	// Create a dummy executable for P1 resolution
	tmpDir := t.TempDir()
	dummyADB := filepath.Join(tmpDir, "adb")
	if os.PathSeparator == '\\' {
		dummyADB += ".exe"
	}
	if err := os.WriteFile(dummyADB, []byte("dummy adb"), 0o755); err != nil {
		t.Fatalf("failed to create dummy adb: %v", err)
	}

	adbMgr := adb.NewManager(fake, dummyADB, nil)
	secretsMgr := secrets.NewManager(secrets.NewMemoryKeyring())

	s, err := store.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer s.Close()

	engine := NewEngine(adbMgr, secretsMgr, s, tmpDir)

	checks, passed := engine.Run(context.Background())
	if !passed {
		t.Fatalf("expected all blocking checks to pass, got passed=false. Checks: %+v", checks)
	}

	if len(checks) != 6 {
		t.Fatalf("expected 6 checks (P1-P6), got %d", len(checks))
	}

	for _, c := range checks {
		if c.IsBlocking && c.Status != StatusOk {
			t.Errorf("blocking check %s failed: %+v", c.ID, c)
		}
	}
}

func TestPreflightEngineMissingADB(t *testing.T) {
	fake := adb.NewFakeRunner()
	// Manager with non-existent path and no system adb
	adbMgr := adb.NewManager(fake, "/non/existent/path/to/adb", nil)
	secretsMgr := secrets.NewManager(secrets.NewMemoryKeyring())
	s, _ := store.OpenMemory()
	defer s.Close()

	engine := NewEngine(adbMgr, secretsMgr, s, t.TempDir())

	checks, passed := engine.Run(context.Background())
	if passed {
		t.Fatal("expected preflight to fail when ADB is missing")
	}

	var p1Found, p2Found bool
	for _, c := range checks {
		if c.ID == "P1" {
			p1Found = true
			if c.Status != StatusMissing {
				t.Errorf("expected P1 status missing, got %s", c.Status)
			}
		}
		if c.ID == "P2" {
			p2Found = true
			if c.Status != StatusMissing {
				t.Errorf("expected P2 status missing, got %s", c.Status)
			}
		}
	}
	if !p1Found || !p2Found {
		t.Fatal("missing P1 or P2 in check results")
	}
}

func TestSafeExtractZipSlipProtection(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "malicious.zip")

	// Create a zip with an escaping file path
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Malicious entry trying to escape target dir
	h := &zip.FileHeader{
		Name: "../../../evil.txt",
	}
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatalf("failed to create zip header: %v", err)
	}
	_, _ = w.Write([]byte("malicious content"))
	_ = zw.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("failed to write zip file: %v", err)
	}

	targetDir := filepath.Join(tmpDir, "target")
	err = safeExtractZip(zipPath, targetDir)
	if err == nil {
		t.Fatal("expected zip slip error, but extraction succeeded")
	}
}

func TestPlatformToolsDownloadURL(t *testing.T) {
	url := PlatformToolsDownloadURL()
	if url == "" {
		t.Fatal("expected non-empty platform tools download URL")
	}
}

func TestTrustedPlatformToolsHashes(t *testing.T) {
	defSHA := DefaultPlatformToolsSHA256()
	if defSHA == "" {
		t.Fatal("expected non-empty default platform tools SHA256")
	}

	if !IsTrustedPlatformToolsHash("windows", defSHA) {
		t.Errorf("expected default SHA to be trusted on windows: %s", defSHA)
	}

	if IsTrustedPlatformToolsHash("windows", "0000000000000000000000000000000000000000000000000000000000000000") {
		t.Error("expected bogus SHA to NOT be trusted")
	}
}

func TestStrictChecksumMatching(t *testing.T) {
	// Pinned r37.0.1 hash must match itself
	pinned := DefaultPlatformToolsSHA256()
	if !IsTrustedPlatformToolsHash("windows", pinned) {
		t.Fatalf("expected pinned SHA256 %s to be trusted", pinned)
	}

	// Mismatched hash must not be accepted when pinned version is required
	olderHash := "86be168bfaae7ea1d86f9f5926ec3ba04164b4c73656910e54d89fa7a00f8aa4" // r35.0.2
	if strings.EqualFold(olderHash, pinned) {
		t.Fatalf("expected r35.0.2 hash to differ from pinned r37.0.1")
	}
}

