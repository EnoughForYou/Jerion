package inventory

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

// PackagePathEntry captures pm list output.
type PackagePathEntry struct {
	Package  string
	CodePath string
}

var (
	rePMListLine   = regexp.MustCompile(`^package:(.+)=([a-zA-Z0-9_.]+)$`)
	reUID          = regexp.MustCompile(`(?:userId|appId|uid)=(\d+)`)
	reVersionCode  = regexp.MustCompile(`versionCode=(\d+)`)
	reVersionName  = regexp.MustCompile(`versionName=([^\s]+)`)
	reInstaller    = regexp.MustCompile(`installerPackageName=([a-zA-Z0-9_.]+)`)
	reTime         = regexp.MustCompile(`(?:firstInstallTime|lastUpdateTime)=(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2})`)
	reCodePath     = regexp.MustCompile(`codePath=([^\s]+)`)
	reTargetSDK    = regexp.MustCompile(`targetSdk=(\d+)`)
	reMinSDK       = regexp.MustCompile(`minSdk=(\d+)`)
	reSignerSHA256 = regexp.MustCompile(`(?:Signer #\d+ certificate SHA-256 digest:|[sS]ha-?256:?)\s*([A-Fa-f0-9:]{32,})`)
)

// ParsePMList extracts package names and apk paths from `pm list packages -f`.
func ParsePMList(output string) []PackagePathEntry {
	var entries []PackagePathEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		m := rePMListLine.FindStringSubmatch(trimmed)
		if len(m) == 3 {
			entries = append(entries, PackagePathEntry{
				CodePath: m[1],
				Package:  m[2],
			})
		} else if strings.HasPrefix(trimmed, "package:") {
			// Fallback if no path separator
			pkg := strings.TrimPrefix(trimmed, "package:")
			entries = append(entries, PackagePathEntry{
				Package: pkg,
			})
		}
	}

	return entries
}

// ParseDumpsysPackage parses detailed `dumpsys package <pkg>` output into AppInfo.
func ParseDumpsysPackage(output string) *AppInfo {
	info := &AppInfo{
		Enabled:              true,
		RequestedPermissions: []string{},
		GrantedPermissions:   []string{},
		CertSHA256:           []string{},
	}

	lines := strings.Split(output, "\n")
	inRequestedPerms := false
	inInstallPerms := false
	inRuntimePerms := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect package name if not set
		if info.Package == "" && strings.HasPrefix(trimmed, "Package [") {
			start := len("Package [")
			end := strings.Index(trimmed[start:], "]")
			if end != -1 {
				pkg := trimmed[start : start+end]
				info.Package = pkg
				info.Label = prettifyPackageLabel(pkg)
			}
		}

		// UID
		if info.UID == 0 {
			if m := reUID.FindStringSubmatch(trimmed); len(m) == 2 {
				if uid, err := strconv.Atoi(m[1]); err == nil {
					info.UID = uid
				}
			}
		}

		// Version code
		if info.VersionCode == 0 {
			if m := reVersionCode.FindStringSubmatch(trimmed); len(m) == 2 {
				if vc, err := strconv.ParseInt(m[1], 10, 64); err == nil {
					info.VersionCode = vc
				}
			}
		}

		// Version name
		if info.VersionName == "" {
			if m := reVersionName.FindStringSubmatch(trimmed); len(m) == 2 {
				info.VersionName = m[1]
			}
		}

		// Target SDK
		if info.TargetSDK == 0 {
			if m := reTargetSDK.FindStringSubmatch(trimmed); len(m) == 2 {
				if s, err := strconv.Atoi(m[1]); err == nil {
					info.TargetSDK = s
				}
			}
		}

		// Min SDK
		if info.MinSDK == 0 {
			if m := reMinSDK.FindStringSubmatch(trimmed); len(m) == 2 {
				if s, err := strconv.Atoi(m[1]); err == nil {
					info.MinSDK = s
				}
			}
		}

		// Installer
		if info.Installer == nil {
			if m := reInstaller.FindStringSubmatch(trimmed); len(m) == 2 {
				inst := m[1]
				info.Installer = &inst
			}
		}

		// First install time
		if info.FirstInstall.IsZero() && strings.Contains(trimmed, "firstInstallTime=") {
			if m := reTime.FindStringSubmatch(trimmed); len(m) == 2 {
				if t, err := time.Parse("2006-01-02 15:04:05", m[1]); err == nil {
					info.FirstInstall = t
				}
			}
		}

		// Last update time
		if info.LastUpdate.IsZero() && strings.Contains(trimmed, "lastUpdateTime=") {
			if m := reTime.FindStringSubmatch(trimmed); len(m) == 2 {
				if t, err := time.Parse("2006-01-02 15:04:05", m[1]); err == nil {
					info.LastUpdate = t
				}
			}
		}

		// Code path
		if info.CodePath == "" {
			if m := reCodePath.FindStringSubmatch(trimmed); len(m) == 2 {
				info.CodePath = m[1]
			}
		}

		// Flags: check SYSTEM
		if strings.Contains(trimmed, "flags=[") {
			if strings.Contains(trimmed, "SYSTEM") {
				info.IsSystem = true
			}
		}

		// Enabled state
		if strings.Contains(trimmed, "enabled=") {
			if strings.Contains(trimmed, "enabled=2") || strings.Contains(trimmed, "enabled=3") {
				info.Enabled = false
			}
		}

		// Launcher check
		if strings.Contains(trimmed, "android.intent.action.MAIN") || strings.Contains(trimmed, "android.intent.category.LAUNCHER") {
			info.HasLauncher = true
		}

		// Certificate digests
		if m := reSignerSHA256.FindStringSubmatch(trimmed); len(m) == 2 {
			fingerprint := strings.ToUpper(strings.ReplaceAll(m[1], ":", ""))
			if !containsString(info.CertSHA256, fingerprint) {
				info.CertSHA256 = append(info.CertSHA256, fingerprint)
			}
		}

		// Permissions section tracking
		switch {
		case strings.HasPrefix(trimmed, "requested permissions:"):
			inRequestedPerms = true
			inInstallPerms = false
			inRuntimePerms = false
			continue
		case strings.HasPrefix(trimmed, "install permissions:"):
			inRequestedPerms = false
			inInstallPerms = true
			inRuntimePerms = false
			continue
		case strings.HasPrefix(trimmed, "runtime permissions:"):
			inRequestedPerms = false
			inInstallPerms = false
			inRuntimePerms = true
			continue
		case strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "permissions"):
			inRequestedPerms = false
			inInstallPerms = false
			inRuntimePerms = false
		}

		if inRequestedPerms {
			perm := strings.TrimSpace(trimmed)
			if strings.HasPrefix(perm, "android.permission.") || strings.Contains(perm, ".permission.") {
				if !containsString(info.RequestedPermissions, perm) {
					info.RequestedPermissions = append(info.RequestedPermissions, perm)
				}
			}
		}

		if inInstallPerms || inRuntimePerms {
			// Format: android.permission.CAMERA: granted=true
			if strings.Contains(trimmed, "granted=true") {
				permName := strings.Split(trimmed, ":")[0]
				permName = strings.TrimSpace(permName)
				if !containsString(info.GrantedPermissions, permName) {
					info.GrantedPermissions = append(info.GrantedPermissions, permName)
				}
			}
		}
	}

	return info
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ParseDeviceAdmins extracts active device admin packages from dumpsys device_policy or dpm list-owners.
// It handles formats such as:
// "User 0: admin=com.example.admin/.AdminReceiver"
// "admin=com.example.admin/.AdminReceiver"
// "com.google.android.apps.work.clouddpc/.receivers.CloudDeviceAdminReceiver"
func ParseDeviceAdmins(output string) map[string]bool {
	admins := make(map[string]bool)
	reAdmin := regexp.MustCompile(`(?:admin=|\s|^)([a-zA-Z0-9_.]+)/[a-zA-Z0-9_.]*`)
	inAdminsSection := false

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Enabled Device Admins") || strings.Contains(trimmed, "Active Admins") {
			inAdminsSection = true
			continue
		}
		if inAdminsSection {
			// A non-indented header that ends with ":" and is not "Enabled Device Admins (user ...):" marks section end
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "Admins") {
				inAdminsSection = false
			}
		}

		if inAdminsSection || strings.Contains(trimmed, "admin=") {
			if m := reAdmin.FindStringSubmatch(trimmed); len(m) >= 2 {
				pkg := m[1]
				if err := adb.ValidatePackageName(pkg); err == nil {
					admins[pkg] = true
				}
			}
		}
	}
	return admins
}

// ParseAppOpsOverlay parses dumpsys appops output to identify packages with SYSTEM_ALERT_WINDOW actively allowed.
// It returns the parsed map and a boolean indicating whether the output was recognized as valid AppOps output.
func ParseAppOpsOverlay(output string) (map[string]bool, bool) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, false
	}

	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "could not find") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "permission denial") ||
		strings.Contains(lower, "can't find service") ||
		strings.Contains(lower, "error") ||
		strings.Contains(lower, "unknown command") {
		return nil, false
	}

	// Must contain recognizable AppOps signatures
	isRecognized := strings.Contains(lower, "appops service state") ||
		strings.Contains(lower, "package ") ||
		strings.Contains(lower, "no operations") ||
		strings.Contains(lower, "system_alert_window") ||
		strings.Contains(lower, "op system_alert_window")

	if !isRecognized {
		return nil, false
	}

	overlayMap := make(map[string]bool)
	lines := strings.Split(output, "\n")
	currentPkg := ""

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}

		lowerLine := strings.ToLower(l)
		if strings.HasPrefix(lowerLine, "package ") {
			pkgPart := strings.TrimSpace(l[len("package "):])
			fields := strings.Fields(pkgPart)
			if len(fields) > 0 {
				token := strings.Trim(fields[0], "[]:()")
				if err := adb.ValidatePackageName(token); err == nil {
					currentPkg = token
				} else {
					currentPkg = ""
				}
			}
			continue
		}

		if currentPkg != "" && strings.Contains(l, "SYSTEM_ALERT_WINDOW") {
			lineLower := strings.ToLower(l)
			// Check for negative modes
			if strings.Contains(lineLower, "ignore") || strings.Contains(lineLower, "deny") || strings.Contains(lineLower, "default") || strings.Contains(lineLower, "mode=1") || strings.Contains(lineLower, "mode=2") || strings.Contains(lineLower, "mode=3") {
				continue
			}
			if strings.Contains(lineLower, "allow") || strings.Contains(lineLower, "mode=0") || strings.Contains(lineLower, "mode=allow") {
				overlayMap[currentPkg] = true
			}
		}
	}
	return overlayMap, true
}

// ParseSingleAppOpOverlay parses the output of `appops get <pkg> SYSTEM_ALERT_WINDOW` or `cmd appops get <pkg> SYSTEM_ALERT_WINDOW`.
// It returns (isAllowed, isRecognized).
func ParseSingleAppOpOverlay(output string) (bool, bool) {
	lower := strings.ToLower(strings.TrimSpace(output))
	if lower == "" {
		return false, false
	}

	if strings.Contains(lower, "could not find") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "permission denial") ||
		strings.Contains(lower, "can't find service") ||
		strings.Contains(lower, "error") ||
		strings.Contains(lower, "unknown command") ||
		strings.Contains(lower, "bad component") {
		return false, false
	}

	// Recognized negative responses (valid response indicating not allowed)
	if strings.Contains(lower, "no operations") ||
		strings.Contains(lower, "ignore") ||
		strings.Contains(lower, "deny") ||
		strings.Contains(lower, "default") ||
		strings.Contains(lower, "mode=1") ||
		strings.Contains(lower, "mode=2") ||
		strings.Contains(lower, "mode=3") ||
		strings.Contains(lower, "mode=ignore") ||
		strings.Contains(lower, "mode=deny") ||
		strings.Contains(lower, "mode=default") {
		return false, true
	}

	// Recognized positive responses
	if strings.Contains(lower, "allow") ||
		strings.Contains(lower, "mode=0") ||
		strings.Contains(lower, "mode=allow") {
		return true, true
	}

	// Unrecognized format
	return false, false
}

