package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

// AccessibilityCollector checks for active accessibility services.
type AccessibilityCollector struct{}

func (c *AccessibilityCollector) Name() string { return "audit-accessibility" }

func (c *AccessibilityCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "settings", "get", "secure", "enabled_accessibility_services")
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(stdout))
	remAction := "open_accessibility"
	remedy := "Review enabled accessibility services in device settings and disable unverified apps."

	item := &AuditItem{
		ID:               "audit-accessibility",
		Title:            "Accessibility Services",
		Description:      "Checks for background accessibility bindings that can inspect screen contents and inject inputs.",
		Category:         CategoryAccessibility,
		RecommendedValue: "None (or system screen readers only)",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	if raw == "" || raw == "null" || raw == "value is null" {
		item.Status = StatusOk
		item.CurrentValue = "None enabled"
	} else {
		// Enabled services present
		services := strings.Split(raw, ":")
		item.CurrentValue = fmt.Sprintf("%d service(s) active: %s", len(services), raw)
		item.Status = StatusWarning
	}

	return item, nil
}

// NotificationListenerCollector audits apps with notification listening permissions.
type NotificationListenerCollector struct{}

func (c *NotificationListenerCollector) Name() string { return "audit-notification-listeners" }

func (c *NotificationListenerCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "cmd", "notification", "get_listeners")
	if err != nil {
		// Fallback to settings query if cmd notification is not supported
		stdout, _, err = runner.Run(ctx, serial, "shell", "settings", "get", "secure", "enabled_notification_listeners")
		if err != nil {
			return nil, err
		}
	}

	raw := strings.TrimSpace(string(stdout))
	remedy := "Revoke notification listener access from unverified applications."
	remAction := "open_notification_access"

	item := &AuditItem{
		ID:               "audit-notification-listeners",
		Title:            "Notification Listeners",
		Description:      "Audits applications capable of reading all incoming notifications, OTPs, and messages.",
		Category:         CategoryNotificationListener,
		RecommendedValue: "Trusted companion or wearable apps only",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	if raw == "" || raw == "null" || raw == "value is null" || strings.Contains(raw, "No listeners") {
		item.Status = StatusOk
		item.CurrentValue = "None active"
	} else {
		item.Status = StatusWarning
		item.CurrentValue = raw
	}

	return item, nil
}

// DeviceAdminCollector checks for registered device administrators.
type DeviceAdminCollector struct{}

func (c *DeviceAdminCollector) Name() string { return "audit-device-admin" }

func (c *DeviceAdminCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "dpm", "list-owners")
	if err != nil {
		stdout, _, err = runner.Run(ctx, serial, "shell", "dumpsys", "device_policy")
		if err != nil {
			return nil, err
		}
	}

	raw := strings.TrimSpace(string(stdout))
	remedy := "Deactivate device administrator privileges in Security Settings before uninstalling."
	remAction := "open_device_admin"

	item := &AuditItem{
		ID:               "audit-device-admin",
		Title:            "Device Administrators",
		Description:      "Audits apps with administrative permissions that can lock the device, erase data, or block uninstallation.",
		Category:         CategoryDeviceAdmin,
		RecommendedValue: "Find My Device or MDM only",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	if raw == "" || strings.Contains(raw, "no active admins") || strings.Contains(raw, "None") {
		item.Status = StatusOk
		item.CurrentValue = "No active device administrators"
	} else if strings.Contains(raw, "com.google.android.gms") {
		item.Status = StatusOk
		item.CurrentValue = "Google Play Services (Find My Device)"
	} else {
		item.Status = StatusCritical
		item.CurrentValue = raw
	}

	return item, nil
}

// UnknownSourcesCollector checks whether global non-market app installation is allowed.
type UnknownSourcesCollector struct{}

func (c *UnknownSourcesCollector) Name() string { return "audit-unknown-sources" }

func (c *UnknownSourcesCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "settings", "get", "secure", "install_non_market_apps")
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(stdout))
	remedy := "Disable global non-market app installations."
	remAction := "toggle_unknown_sources"

	item := &AuditItem{
		ID:               "audit-unknown-sources",
		Title:            "Install Unknown Sources",
		Description:      "Controls whether unverified APK files can be installed outside official app stores.",
		Category:         CategoryDevSettings,
		RecommendedValue: "0 (Disabled globally)",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	if raw == "1" {
		item.Status = StatusWarning
		item.CurrentValue = "1 (Enabled Globally)"
	} else {
		item.Status = StatusOk
		item.CurrentValue = "0 (Disabled)"
	}

	return item, nil
}

// USBDebuggingCollector checks ADB debugging setting.
type USBDebuggingCollector struct{}

func (c *USBDebuggingCollector) Name() string { return "audit-usb-debugging" }

func (c *USBDebuggingCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "settings", "get", "global", "adb_enabled")
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(stdout))
	item := &AuditItem{
		ID:               "audit-usb-debugging",
		Title:            "USB Debugging",
		Description:      "Developer interface enabling computer bridge communication.",
		Category:         CategoryDevSettings,
		RecommendedValue: "Enabled during diagnostic sessions; disable when finished",
	}

	if raw == "1" {
		item.Status = StatusOk
		item.CurrentValue = "Enabled (Active Session)"
	} else {
		item.Status = StatusOk
		item.CurrentValue = "Disabled"
	}

	return item, nil
}

// PrivateDNSCollector checks DNS-over-TLS settings.
type PrivateDNSCollector struct{}

func (c *PrivateDNSCollector) Name() string { return "audit-private-dns" }

func (c *PrivateDNSCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "settings", "get", "global", "private_dns_mode")
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(stdout))
	remedy := "Enable Private DNS with an encrypted DNS provider (e.g., Cloudflare 1.1.1.1 or Quad9)."
	remAction := "configure_dns"

	item := &AuditItem{
		ID:               "audit-private-dns",
		Title:            "Encrypted Private DNS",
		Description:      "Protects network domain requests from local ISP or Wi-Fi eavesdropping and ad tracking.",
		Category:         CategoryNetworkSecurity,
		RecommendedValue: "strict / hostname mode",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	switch raw {
	case "hostname", "strict":
		item.Status = StatusOk
		item.CurrentValue = "Configured (Encrypted)"
	case "opportunistic", "automatic":
		item.Status = StatusOk
		item.CurrentValue = "Automatic"
	case "off", "":
		item.Status = StatusWarning
		item.CurrentValue = "Off"
	default:
		item.Status = StatusOk
		item.CurrentValue = raw
	}

	return item, nil
}

// ProxyCollector verifies no malicious network HTTP proxy is configured.
type ProxyCollector struct{}

func (c *ProxyCollector) Name() string { return "audit-proxy" }

func (c *ProxyCollector) Collect(ctx context.Context, runner adb.Runner, serial string) (*AuditItem, error) {
	stdout, _, err := runner.Run(ctx, serial, "shell", "settings", "get", "global", "http_proxy")
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(stdout))
	remedy := "Remove unexpected HTTP proxy settings in Wi-Fi and Network preferences."
	remAction := "clear_proxy"

	item := &AuditItem{
		ID:               "audit-proxy",
		Title:            "System HTTP Proxy",
		Description:      "Verifies device network traffic is not redirected through an unverified intermediary proxy.",
		Category:         CategoryNetworkSecurity,
		RecommendedValue: ":0 (None)",
		Remediation:      &remedy,
		ActionType:       &remAction,
	}

	if raw == "" || raw == ":0" || raw == "null" || raw == "value is null" {
		item.Status = StatusOk
		item.CurrentValue = "None (Direct connection)"
	} else {
		item.Status = StatusCritical
		item.CurrentValue = fmt.Sprintf("Proxy configured: %s", raw)
	}

	return item, nil
}
