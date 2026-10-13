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
		return "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
	case "darwin":
		return "https://dl.google.com/android/repository/platform-tools-latest-darwin.zip"
	default:
		return "https://dl.google.com/android/repository/platform-tools-latest-linux.zip"
	}
}

// ProgressFunc reports download advancement.
type ProgressFunc func(downloadedBytes, totalBytes int64)

// Downloader securely downloads and unpacks Google Android platform-tools.
type Downloader struct {
	client *http.Client
}

// NewDownloader creates an ADB Downloader.
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// DownloadAndExtract downloads the official zip, verifies checksum if expectedSHA is non-empty,
// and safely extracts platform-tools into destDir.
func (d *Downloader) DownloadAndExtract(ctx context.Context, expectedSHA string, onProgress ProgressFunc) (string, error) {
	url := PlatformToolsDownloadURL()

	destExec, err := adb.DefaultAppToolsPath()
	if err != nil {
		return "", fmt.Errorf("resolve tools path: %w", err)
	}
	toolsDir := filepath.Dir(filepath.Dir(destExec)) // %LOCALAPPDATA%\Jerion\tools

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

	// 2. Checksum verification
	computedSHA := hex.EncodeToString(hasher.Sum(nil))
	if expectedSHA != "" && !strings.EqualFold(computedSHA, expectedSHA) {
		return "", fmt.Errorf("checksum verification failed: expected %s, got %s", expectedSHA, computedSHA)
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
