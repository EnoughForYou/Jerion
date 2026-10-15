package preflight

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

// PlatformToolsDownloadURL returns Google's official platform-tools zip URL for current OS.
func PlatformToolsDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return "https://dl.google.com/android/repository/platform-tools_r37.0.1-win.zip"
	case "darwin":
		return "https://dl.google.com/android/repository/platform-tools_r37.0.1-darwin.zip"
	default:
		return "https://dl.google.com/android/repository/platform-tools_r37.0.1-linux.zip"
	}
}

// TrustedPlatformToolsHashes maintains a catalog of known official Google platform-tools SHA-256 digests.
var TrustedPlatformToolsHashes = map[string][]string{
	"windows": {
		"45f4d63113e895ebde0c90f194099a4676b6ac653bd28d54314a9e022bbc1a99", // r37.0.1 Windows (and latest-windows.zip)
		"86be168bfaae7ea1d86f9f5926ec3ba04164b4c73656910e54d89fa7a00f8aa4", // r35.0.2 Windows
		"0c7fa823a31c50e419b4fcb5fdfcbbbe94e77749a215386db4b50c534433017a", // r35.0.1 Windows
		"e4125d038eaeb6f391b1511fa6770f5e714652c7921a2f64fcb9e38d4c1537dc", // r34.0.5 Windows
	},
	"darwin": {
		"ee39ad5967e95c2a07f04dbcbde96b1a0c916ba376096db5d2f498b7727a5d1d", // r37.0.1 Darwin
		"6106e236314f36a56e297825d19aa91f14a60155a0247fe051871a2a1a2b25c3",
	},
	"linux": {
		"d230f13842f60f782a8645f9c813f8f845bf36089ea7289f28c48f17979313f1", // r37.0.1 Linux
		"7f2ec9cf5c6628c5a2c4e857ff17135ee7d21c0ad19a7168307db1854bc2845c",
	},
}

// DefaultPlatformToolsSHA256 returns the primary known trusted SHA-256 hash for current OS.
func DefaultPlatformToolsSHA256() string {
	hashes := TrustedPlatformToolsHashes[runtime.GOOS]
	if len(hashes) > 0 {
		return hashes[0]
	}
	return ""
}

// IsTrustedPlatformToolsHash checks if hash matches the trusted Google catalog for the given OS.
func IsTrustedPlatformToolsHash(goos, hash string) bool {
	trusted, ok := TrustedPlatformToolsHashes[goos]
	if !ok {
		return false
	}
	for _, t := range trusted {
		if strings.EqualFold(t, hash) {
			return true
		}
	}
	return false
}

// ProgressFunc reports download advancement.
type ProgressFunc func(downloadedBytes, totalBytes int64)

// Downloader securely downloads and unpacks Google Android platform-tools.
type Downloader struct {
	client      *http.Client
	downloadURL string
	toolsDir    string
}

// NewDownloader creates an ADB Downloader.
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// NewDownloaderWithURL creates an ADB Downloader with custom download URL and optional tools directory.
func NewDownloaderWithURL(downloadURL string, toolsDir string) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
		downloadURL: downloadURL,
		toolsDir:    toolsDir,
	}
}

// DownloadAndExtract downloads the official zip, verifies checksum strictly (mandatory per ADR-0002),
// and safely extracts platform-tools into destDir.
func (d *Downloader) DownloadAndExtract(ctx context.Context, expectedSHA string, onProgress ProgressFunc) (string, error) {
	url := d.downloadURL
	if url == "" {
		url = PlatformToolsDownloadURL()
	}

	var toolsDir string
	var destExec string
	if d.toolsDir != "" {
		toolsDir = d.toolsDir
		execName := "adb"
		if runtime.GOOS == "windows" {
			execName = "adb.exe"
		}
		destExec = filepath.Join(toolsDir, "platform-tools", execName)
	} else {
		resolved, err := adb.DefaultAppToolsPath()
		if err != nil {
			return "", fmt.Errorf("resolve tools path: %w", err)
		}
		destExec = resolved
		toolsDir = filepath.Dir(filepath.Dir(destExec)) // %LOCALAPPDATA%\Jerion\tools
	}

	if err := os.MkdirAll(toolsDir, 0o750); err != nil {
		return "", fmt.Errorf("create tools directory: %w", err)
	}

	tempZip := filepath.Join(toolsDir, "platform-tools.zip.tmp")
	defer func() {
		_ = os.Remove(tempZip)
	}()

	// 1. Download
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create download request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	out, err := os.OpenFile(tempZip, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("create temp zip: %w", err)
	}

	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)

	totalBytes := resp.ContentLength
	var downloadedBytes int64

	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			_ = out.Close()
			return "", ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := writer.Write(buf[:n]); wErr != nil {
				_ = out.Close()
				return "", fmt.Errorf("write download chunk: %w", wErr)
			}
			downloadedBytes += int64(n)
			if onProgress != nil {
				onProgress(downloadedBytes, totalBytes)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			_ = out.Close()
			return "", fmt.Errorf("read download stream: %w", readErr)
		}
	}
	_ = out.Close()

	// 2. Checksum verification (mandatory per ADR-0002)
	computedSHA := hex.EncodeToString(hasher.Sum(nil))
	if expectedSHA != "" {
		if !strings.EqualFold(computedSHA, expectedSHA) {
			return "", fmt.Errorf("checksum verification failed: expected %s, got %s", expectedSHA, computedSHA)
		}
	} else {
		if !IsTrustedPlatformToolsHash(runtime.GOOS, computedSHA) {
			return "", fmt.Errorf("checksum verification required: unverified download hash %s does not match any trusted Google platform-tools checksum", computedSHA)
		}
	}

	// 3. Safe Extraction
	if err := safeExtractZip(tempZip, toolsDir); err != nil {
		return "", fmt.Errorf("extract platform-tools: %w", err)
	}

	// Ensure executable has execution permissions on POSIX
	if runtime.GOOS != "windows" {
		_ = os.Chmod(destExec, 0o755)
	}

	return destExec, nil
}

// safeExtractZip unpacks an archive while preventing directory traversal attacks (zip slip).
func safeExtractZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanTarget := filepath.Clean(targetDir)

	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		destPath := filepath.Join(cleanTarget, cleanName)

		if !strings.HasPrefix(destPath, cleanTarget+string(filepath.Separator)) && destPath != cleanTarget {
			return fmt.Errorf("illegal zip file path escaping target: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o750); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()

		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}
