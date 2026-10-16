package preflight

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
	t.Setenv("LOCALAPPDATA", t.TempDir())
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

	// Default hash must be trusted for the host environment OS
	if !IsTrustedPlatformToolsHash(runtime.GOOS, defSHA) {
		t.Errorf("expected default SHA to be trusted on host OS %s: %s", runtime.GOOS, defSHA)
	}

	// Table-driven verification across all supported operating systems (Windows, Darwin, Linux)
	tests := []struct {
		name     string
		goos     string
		hash     string
		expected bool
	}{
		{
			name:     "Windows r37.0.1 trusted",
			goos:     "windows",
			hash:     "45f4d63113e895ebde0c90f194099a4676b6ac653bd28d54314a9e022bbc1a99",
			expected: true,
		},
		{
			name:     "Windows r35.0.2 trusted",
			goos:     "windows",
			hash:     "86be168bfaae7ea1d86f9f5926ec3ba04164b4c73656910e54d89fa7a00f8aa4",
			expected: true,
		},
		{
			name:     "Darwin r37.0.1 trusted",
			goos:     "darwin",
			hash:     "ee39ad5967e95c2a07f04dbcbde96b1a0c916ba376096db5d2f498b7727a5d1d",
			expected: true,
		},
		{
			name:     "Linux r37.0.1 trusted",
			goos:     "linux",
			hash:     "d230f13842f60f782a8645f9c813f8f845bf36089ea7289f28c48f17979313f1",
			expected: true,
		},
		{
			name:     "Unsupported OS untrusted",
			goos:     "freebsd",
			hash:     "45f4d63113e895ebde0c90f194099a4676b6ac653bd28d54314a9e022bbc1a99",
			expected: false,
		},
		{
			name:     "Windows r37.0.1 trusted uppercase and whitespace",
			goos:     "windows",
			hash:     "  45F4D63113E895EBDE0C90F194099A4676B6AC653BD28D54314A9E022BBC1A99 \n",
			expected: true,
		},
		{
			name:     "Bogus SHA untrusted on windows",
			goos:     "windows",
			hash:     "0000000000000000000000000000000000000000000000000000000000000000",
			expected: false,
		},
		{
			name:     "Bogus SHA untrusted on linux",
			goos:     "linux",
			hash:     "0000000000000000000000000000000000000000000000000000000000000000",
			expected: false,
		},
		{
			name:     "Bogus SHA untrusted on darwin",
			goos:     "darwin",
			hash:     "0000000000000000000000000000000000000000000000000000000000000000",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTrustedPlatformToolsHash(tc.goos, tc.hash)
			if got != tc.expected {
				t.Errorf("IsTrustedPlatformToolsHash(%q, %q) = %v; want %v", tc.goos, tc.hash, got, tc.expected)
			}
		})
	}
}

func TestStrictChecksumMatching(t *testing.T) {
	// Pinned r37.0.1 hash for host OS must match itself
	pinned := DefaultPlatformToolsSHA256()
	if pinned == "" {
		t.Fatal("expected non-empty default platform tools SHA256")
	}

	if !IsTrustedPlatformToolsHash(runtime.GOOS, pinned) {
		t.Fatalf("expected pinned SHA256 %s to be trusted on %s", pinned, runtime.GOOS)
	}

	// Mismatched hash must not be accepted when pinned version is required
	bogusHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if strings.EqualFold(bogusHash, pinned) {
		t.Fatalf("expected bogus hash to differ from pinned")
	}
	if IsTrustedPlatformToolsHash(runtime.GOOS, bogusHash) {
		t.Fatalf("expected bogus hash to not be trusted on %s", runtime.GOOS)
	}
}

func TestDownloaderChecksumVerification(t *testing.T) {
	// Build a valid synthetic platform-tools zip archive in memory
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	execName := "adb"
	if os.PathSeparator == '\\' {
		execName = "adb.exe"
	}
	fHeader := &zip.FileHeader{
		Name: "platform-tools/" + execName,
	}
	w, err := zw.CreateHeader(fHeader)
	if err != nil {
		t.Fatalf("failed to create zip file header: %v", err)
	}
	sampleContent := []byte("synthetic-adb-binary-v37.0.1")
	if _, err := w.Write(sampleContent); err != nil {
		t.Fatalf("failed to write file to zip: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to finalize zip: %v", err)
	}

	zipBytes := zipBuf.Bytes()
	hasher := sha256.New()
	hasher.Write(zipBytes)
	realSHA := hex.EncodeToString(hasher.Sum(nil))

	// Start test HTTP server serving the synthetic archive
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipBytes)
	}))
	defer ts.Close()

	// Case 1: Mismatched checksum must be rejected and must not extract
	tempToolsDir1 := t.TempDir()
	d1 := NewDownloaderWithURL(ts.URL, tempToolsDir1)
	bogusSHA := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	_, err = d1.DownloadAndExtract(context.Background(), bogusSHA, nil)
	if err == nil {
		t.Fatalf("expected DownloadAndExtract to fail with mismatched checksum, got nil error")
	}
	if !strings.Contains(err.Error(), "checksum verification failed") {
		t.Fatalf("expected checksum verification failed error, got: %v", err)
	}

	// Verify target executable was NOT created and temp zip was cleaned up
	extractedExec1 := filepath.Join(tempToolsDir1, "platform-tools", execName)
	if _, statErr := os.Stat(extractedExec1); !os.IsNotExist(statErr) {
		t.Fatalf("executable should not exist after checksum failure: %s", extractedExec1)
	}
	tempZip1 := filepath.Join(tempToolsDir1, "platform-tools.zip.tmp")
	if _, statErr := os.Stat(tempZip1); !os.IsNotExist(statErr) {
		t.Fatalf("temporary zip should have been cleaned up after checksum failure: %s", tempZip1)
	}

	// Case 2: Matching checksum must succeed and extract the binary
	tempToolsDir2 := t.TempDir()
	d2 := NewDownloaderWithURL(ts.URL, tempToolsDir2)

	extractedPath, err := d2.DownloadAndExtract(context.Background(), realSHA, nil)
	if err != nil {
		t.Fatalf("expected DownloadAndExtract with matching checksum to succeed, got: %v", err)
	}
	if extractedPath == "" {
		t.Fatalf("expected non-empty extracted executable path")
	}

	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("failed to read extracted binary: %v", err)
	}
	if string(extractedContent) != string(sampleContent) {
		t.Fatalf("extracted binary content mismatch: expected %q, got %q", sampleContent, extractedContent)
	}
}


