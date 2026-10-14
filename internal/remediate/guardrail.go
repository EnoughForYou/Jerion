package remediate

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	pkgRegex  = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	permRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)

	// criticalSystemPackages are protected from destructive actions (disable / uninstall).
	criticalSystemPackages = map[string]string{
		"android":                               "Core Android operating system framework",
		"com.android.systemui":                  "System UI (status bar, navigation, notifications)",
		"com.android.settings":                  "Android System Settings",
		"com.android.phone":                     "Telephony and cellular radio stack",
		"com.android.server.telecom":            "Telecom call management service",
		"com.android.dialer":                    "Default dialer application",
		"com.google.android.dialer":             "Google Phone and dialer application",
		"com.google.android.gms":                "Google Play Services runtime framework",
		"com.android.vending":                   "Google Play Store client",
		"com.google.android.gsf":                "Google Services Framework",
		"com.android.packageinstaller":          "Package installer service",
		"com.google.android.packageinstaller":   "Google package installer",
		"com.android.inputmethod.latin":         "Default Android Latin keyboard",
		"com.google.android.inputmethod.latin":  "Google Gboard keyboard",
		"com.android.launcher":                  "Default Android home launcher",
		"com.android.launcher3":                 "AOSP home launcher",
		"com.google.android.apps.nexuslauncher": "Google Pixel launcher",
		"com.transsion.launcher":                "Transsion / XOS desktop launcher",
		"com.transsion.systemupdate":            "Transsion XOS system update daemon",
		"com.transsion.smartpanel":              "Transsion XOS smart panel accessibility tool",
		"com.transsion.phonemaster":             "Transsion XOS phone master device manager",
	}
)

// ValidatePackageName validates that a package name is safe from command injection.
func ValidatePackageName(pkg string) error {
	if !pkgRegex.MatchString(pkg) {
		return fmt.Errorf("invalid package name '%s': contains forbidden characters", pkg)
	}
	return nil
}

// ValidatePermissionName validates that a permission string is safe from command injection.
func ValidatePermissionName(perm string) error {
	if !permRegex.MatchString(perm) {
		return fmt.Errorf("invalid permission name '%s': contains forbidden characters", perm)
	}
	return nil
}

// CheckGuardrail evaluates whether a package is protected by safety denylists.
func CheckGuardrail(pkg string) (isBlocked bool, reason string) {
	lower := strings.ToLower(strings.TrimSpace(pkg))
	if desc, found := criticalSystemPackages[lower]; found {
		return true, fmt.Sprintf("Critical system package protected by guardrail (%s)", desc)
	}

	// Defensive check: protect any core android prefix
	if lower == "android" || strings.HasPrefix(lower, "com.android.systemui") {
		return true, "Protected system component required for device stability"
	}

	return false, ""
}
