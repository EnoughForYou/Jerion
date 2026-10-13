package inventory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
)

// Puller manages extracting APK binaries from an attached device.
type Puller struct {
	runner adb.Runner
	logger *logx.Logger
}

// NewPuller constructs an APK puller.
func NewPuller(runner adb.Runner, logger *logx.Logger) *Puller {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &Puller{
		runner: runner,
		logger: logger,
	}
}

// Pull extracts all APKs (base + splits) for a package into destDir and returns the base APK path and its SHA-256 hash.
func (p *Puller) Pull(ctx context.Context, serial, pkg, destDir string) (string, []string, string, error) {
	if err := adb.ValidatePackageName(pkg); err != nil {
		return "", nil, "", err
	}

	// 1. Query APK path on device
	stdout, stderr, err := p.runner.Run(ctx, serial, "shell", "pm", "path", pkg)
	if err != nil {
		return "", nil, "", adb.NewADBError(adb.ErrCodeCommandFailed, fmt.Sprintf("failed to get pm path: %s", string(stderr)), err)
	}

	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	var remotePaths []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package:") {
			remotePaths = append(remotePaths, strings.TrimPrefix(trimmed, "package:"))
		}
	}

	if len(remotePaths) == 0 {
		return "", nil, "", adb.NewADBError(adb.ErrCodeCommandFailed, fmt.Sprintf("no APK paths found for package %s", pkg), nil)
	}

	pkgDir := filepath.Join(destDir, pkg)
	if err := os.MkdirAll(pkgDir, 0o750); err != nil {
		return "", nil, "", fmt.Errorf("create package destination directory: %w", err)
	}

	var localPaths []string
	var basePath string
	var baseSHA string

	for _, remote := range remotePaths {
		fileName := filepath.Base(remote)
		localFile := filepath.Join(pkgDir, fileName)

		// Use adb pull
		_, pullErrStderr, pullErr := p.runner.Run(ctx, serial, "pull", remote, localFile)
		if pullErr != nil {
			return "", nil, "", adb.NewADBError(adb.ErrCodeCommandFailed, fmt.Sprintf("failed to pull %s: %s", remote, string(pullErrStderr)), pullErr)
		}

		localPaths = append(localPaths, localFile)

		// Base apk or first apk
		if fileName == "base.apk" || basePath == "" {
			basePath = localFile
			sha, err := computeFileSHA256(localFile)
			if err != nil {
				return "", nil, "", fmt.Errorf("hash pulled apk %s: %w", localFile, err)
			}
			baseSHA = sha
		}
	}

	return basePath, localPaths, baseSHA, nil
}

func computeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
