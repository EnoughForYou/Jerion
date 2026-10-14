package inventory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
	"github.com/EnoughForYou/Jerion/internal/logx"
)

// AppInfo represents application metadata matching contract v1.0.0.
type AppInfo struct {
	Package              string    `json:"package"`
	Label                string    `json:"label"`
	VersionName          string    `json:"versionName"`
	VersionCode          int64     `json:"versionCode"`
	UID                  int       `json:"uid"`
	Installer            *string   `json:"installer,omitempty"`
	TargetSDK            int       `json:"targetSdk"`
	MinSDK               int       `json:"minSdk"`
	FirstInstall         time.Time `json:"firstInstall"`
	LastUpdate           time.Time `json:"lastUpdate"`
	IsSystem             bool      `json:"isSystem"`
	Enabled              bool      `json:"enabled"`
	HasLauncher          bool      `json:"hasLauncher"`
	RequestedPermissions []string  `json:"requestedPermissions"`
	GrantedPermissions   []string  `json:"grantedPermissions"`
	CertSHA256           []string  `json:"certSha256"`
	APKSHA256            *string   `json:"apkSha256,omitempty"`
	CodePath             string    `json:"codePath,omitempty"`
	SplitPaths           []string  `json:"splitPaths,omitempty"`
}

// Service handles application inventory discovery and inspection.
type Service struct {
	runner adb.Runner
	logger *logx.Logger
}

// NewService creates a new inventory Service.
func NewService(runner adb.Runner, logger *logx.Logger) *Service {
	if logger == nil {
		logger = logx.GetDefault()
	}
	return &Service{
		runner: runner,
		logger: logger,
	}
}

// ListPackages returns all installed packages (or user-installed only if userOnly is true).
func (s *Service) ListPackages(ctx context.Context, serial string, userOnly bool) ([]AppInfo, error) {
	args := []string{"shell", "pm", "list", "packages", "-f", "-u"}
	if userOnly {
		args = []string{"shell", "pm", "list", "packages", "-f", "-3"}
	}

	stdout, stderr, err := s.runner.Run(ctx, serial, args...)
	if err != nil {
		return nil, adb.NewADBError(adb.ErrCodeCommandFailed, fmt.Sprintf("failed to list packages: %s", string(stderr)), err)
	}

	entries := ParsePMList(string(stdout))
	var apps []AppInfo

	for _, entry := range entries {
		info := AppInfo{
			Package:              entry.Package,
			Label:                prettifyPackageLabel(entry.Package),
			CodePath:             entry.CodePath,
			IsSystem:             strings.HasPrefix(entry.CodePath, "/system") || strings.HasPrefix(entry.CodePath, "/vendor") || strings.HasPrefix(entry.CodePath, "/product"),
			Enabled:              true,
			RequestedPermissions: []string{},
			GrantedPermissions:   []string{},
			CertSHA256:           []string{},
		}
		apps = append(apps, info)
	}

	return apps, nil
}

// GetPackageDetail queries `dumpsys package <pkg>` and extracts detailed permissions, installer, and certs.
func (s *Service) GetPackageDetail(ctx context.Context, serial, pkg string) (*AppInfo, error) {
	if err := adb.ValidatePackageName(pkg); err != nil {
		return nil, err
	}

	stdout, stderr, err := s.runner.Run(ctx, serial, "shell", "dumpsys", "package", pkg)
	if err != nil {
		return nil, adb.NewADBError(adb.ErrCodeCommandFailed, fmt.Sprintf("failed to dumpsys package %s: %s", pkg, string(stderr)), err)
	}

	info := ParseDumpsysPackage(string(stdout))
	if info == nil || info.Package == "" {
		info = &AppInfo{
			Package: pkg,
			Label:   prettifyPackageLabel(pkg),
		}
	}

	return info, nil
}

func prettifyPackageLabel(pkg string) string {
	parts := strings.Split(pkg, ".")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) > 0 {
			return strings.ToUpper(last[:1]) + last[1:]
		}
	}
	return pkg
}
