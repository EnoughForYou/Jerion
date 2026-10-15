package contract

// PreflightCheck represents requirement verification status matching contract v1.0.0.
type PreflightCheck struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Status           string   `json:"status"` // "checking" | "ok" | "warning" | "missing" | "downloading" | "failed"
	Detail           string   `json:"detail"`
	Remedy           *string  `json:"remedy,omitempty"`
	SourceURL        *string  `json:"sourceUrl,omitempty"`
	DownloadProgress *float64 `json:"downloadProgress,omitempty"`
}

// Device represents connected hardware metadata.
type Device struct {
	ID              string `json:"id"`
	Serial          string `json:"serial"`
	Model           string `json:"model"`
	Manufacturer    string `json:"manufacturer"`
	AndroidVersion  string `json:"androidVersion"`
	SDKInt          int    `json:"sdkInt"`
	BatteryLevel    int    `json:"batteryLevel"`
	IsCharging      bool   `json:"isCharging"`
	ConnectionState string `json:"connectionState"` // "ready" | "unauthorized" | "offline" | "no_device"
	IsEmulator      bool   `json:"isEmulator"`
}

// SecuritySummary represents aggregate security posture without percent scores.
type SecuritySummary struct {
	Counts     map[string]int `json:"counts"` // key: severity ("info", "low", "medium", "high", "critical")
	TotalApps  int            `json:"totalApps"`
	LastScanAt *string        `json:"lastScanAt,omitempty"`
	Verdict    string         `json:"verdict"` // "clean" | "review" | "high_risk"
}

// AppInfo represents inspected application metadata.
type AppInfo struct {
	Package                    string   `json:"package"`
	Label                      string   `json:"label"`
	VersionName                string   `json:"versionName"`
	VersionCode                string   `json:"versionCode"` // String representation to avoid JS 53-bit overflow
	IsSystem                   bool     `json:"isSystem"`
	IsSideloaded               bool     `json:"isSideloaded"`
	InstallerPackage           *string  `json:"installerPackage,omitempty"`
	TargetSDK                  int      `json:"targetSdk"`
	MinSDK                     int      `json:"minSdk"`
	InstalledAt                string   `json:"installedAt"`
	UpdatedAt                  string   `json:"updatedAt"`
	Permissions                []string `json:"permissions"`
	DangerousPermissions       []string `json:"dangerousPermissions"`
	HasOverlayPermission       bool     `json:"hasOverlayPermission"`
	HasAccessibilityPermission bool     `json:"hasAccessibilityPermission"`
	IsDeviceAdmin              bool     `json:"isDeviceAdmin"`
	IsBatteryOptimized         bool     `json:"isBatteryOptimized"`
	Signatures                 []string `json:"signatures"`
	SuspiciousFlags            []string `json:"suspiciousFlags"`
}

// InventoryFilter filters application inventory.
type InventoryFilter struct {
	Category string `json:"category"` // "all" | "user" | "system" | "disabled" | "risky"
	Search   string `json:"search"`
}

// ScanOptions configures scan mode and scope.
type ScanOptions struct {
	Mode          string `json:"mode"` // "quick" | "deep"
	IncludeSystem bool   `json:"includeSystem"`
}

// ScanProgress communicates real-time scanning percentage.
type ScanProgress struct {
	Step           string  `json:"step"`
	Progress       int     `json:"progress"` // 0 - 100
	CurrentPackage *string `json:"currentPackage,omitempty"`
}

// Evidence captures proof points.
type Evidence struct {
	Type      string  `json:"type"` // "permission" | "manifest" | "class_prefix" | "hash" | "setting" | "log"
	Title     string  `json:"title"`
	Details   string  `json:"details"`
	RawOutput *string `json:"rawOutput,omitempty"`
}

// RemediationAction represents a guarded action.
type RemediationAction struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"` // "disable" | "uninstall" | "revoke_permission" | "clear_appop"
	Package       string  `json:"package"`
	Title         string  `json:"title"`
	Impact        string  `json:"impact"`
	Undoable      bool    `json:"undoable"`
	IsBlocked     bool    `json:"isBlocked"`
	BlockedReason *string `json:"blockedReason,omitempty"`
}

// Finding represents an individual forensic detection.
type Finding struct {
	ID                string             `json:"id"`
	RuleID            string             `json:"ruleId"`
	Title             string             `json:"title"`
	Severity          string             `json:"severity"`   // "info" | "low" | "medium" | "high" | "critical"
	Confidence        string             `json:"confidence"` // "low" | "medium" | "high"
	Package           *string            `json:"package,omitempty"`
	WhyFlagged        string             `json:"whyFlagged"`
	FalsePositiveHint string             `json:"falsePositiveHint"`
	Evidence          []Evidence         `json:"evidence"`
	RecommendedAction *RemediationAction `json:"recommendedAction,omitempty"`
}

// ScanResult represents complete or intermediate scan findings.
type ScanResult struct {
	ID          string          `json:"id"`
	DeviceID    string          `json:"deviceId"`
	Mode        string          `json:"mode"` // "quick" | "deep"
	StartedAt   string          `json:"startedAt"`
	CompletedAt *string         `json:"completedAt,omitempty"`
	Status      string          `json:"status"` // "running" | "completed" | "cancelled" | "failed" | "partial"
	Findings    []Finding       `json:"findings"`
	Summary     SecuritySummary `json:"summary"`
}

// MonitorEvent represents a runtime event signal.
type MonitorEvent struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Type      string         `json:"type"` // "foreground_switch" | "overlay_shown" | "notification_posted"
	Package   string         `json:"package"`
	Details   map[string]any `json:"details"`
}

// MonitorCandidate represents an adware suspect ranked by confidence and score.
type MonitorCandidate struct {
	Package        string         `json:"package"`
	AppLabel       string         `json:"appLabel"`
	Confidence     string         `json:"confidence"` // "low" | "medium" | "high"
	Score          int            `json:"score"`
	TimelineEvents []MonitorEvent `json:"timelineEvents"`
	WhyFlagged     string         `json:"whyFlagged"`
}

// MonitorStatus represents real-time daemon state.
type MonitorStatus struct {
	IsActive      bool     `json:"isActive"`
	StartedAt     *string  `json:"startedAt,omitempty"`
	SourcesActive []string `json:"sourcesActive"`
}

// MonitorSnapshot captures real-time monitoring state, events, and candidates.
type MonitorSnapshot struct {
	Status     MonitorStatus      `json:"status"`
	Events     []MonitorEvent     `json:"events"`
	Candidates []MonitorCandidate `json:"candidates"`
}

// AuditItem represents a system hardening audit item.
type AuditItem struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Title    string   `json:"title"`
	Status   string   `json:"status"` // "ok" | "warning" | "issue" | "unavailable"
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence"`
	Remedy   *string  `json:"remedy,omitempty"`
}

// RemediationPlan aggregates actions requiring execution.
type RemediationPlan struct {
	Actions        []RemediationAction `json:"actions"`
	RequiresBackup bool                `json:"requiresBackup"`
}

// ActionLogEntry represents persistent audit trail of remediation.
type ActionLogEntry struct {
	ID         string            `json:"id"`
	Timestamp  string            `json:"timestamp"`
	Action     RemediationAction `json:"action"`
	Status     string            `json:"status"` // "success" | "failed" | "undone"
	Error      *string           `json:"error,omitempty"`
	BackupPath *string           `json:"backupPath,omitempty"`
	AuditSaved *bool             `json:"auditSaved,omitempty"`
	AuditError *string           `json:"auditError,omitempty"`
}

// ScanSummaryItem represents a summary record for scan history.
type ScanSummaryItem struct {
	ID            string         `json:"id"`
	StartedAt     string         `json:"startedAt"`
	CompletedAt   *string        `json:"completedAt,omitempty"`
	Mode          string         `json:"mode"`
	Verdict       string         `json:"verdict"`
	FindingsCount map[string]int `json:"findingsCount"`
}

// ScanComparison represents diff between two scan results.
type ScanComparison struct {
	ScanA             ScanSummaryItem `json:"scanA"`
	ScanB             ScanSummaryItem `json:"scanB"`
	NewFindings       []Finding       `json:"newFindings"`
	ResolvedFindings  []Finding       `json:"resolvedFindings"`
	UnchangedFindings []Finding       `json:"unchangedFindings"`
}

// ReportExportOptions configures report export format and destination.
type ReportExportOptions struct {
	ScanID     string  `json:"scanId"`
	Format     string  `json:"format"` // "json" | "html"
	TargetPath *string `json:"targetPath,omitempty"`
}

// Settings represents persistent user preferences.
type Settings struct {
	Theme              string  `json:"theme"` // "system" | "dark" | "light"
	AdbPath            *string `json:"adbPath,omitempty"`
	CustomAdbPort      *int    `json:"customAdbPort,omitempty"`
	DeepScanTimeoutSec int     `json:"deepScanTimeoutSec"`
	AutoCheckPreflight bool    `json:"autoCheckPreflight"`
}

// AllowlistEntry represents a whitelist rule or package exemption.
type AllowlistEntry struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"` // "package" | "rule_package"
	Target  string  `json:"target"`
	AddedAt string  `json:"addedAt"`
	Reason  *string `json:"reason,omitempty"`
}
