package engine

import "time"

// Evidence captures technical proof from manifest, dumpsys, dex, or logcat.
type Evidence struct {
	Source  string `json:"source"`
	Snippet string `json:"snippet"`
}

// Finding represents an individual detection item matching contract v1.0.0.
type Finding struct {
	ID                string     `json:"id"`
	RuleID            string     `json:"ruleId"`
	Package           string     `json:"package"`
	Category          string     `json:"category"`
	Severity          string     `json:"severity"`
	Confidence        string     `json:"confidence"`
	Weight            int        `json:"weight"`
	Title             string     `json:"title"`
	WhyFlagged        string     `json:"whyFlagged"`
	FalsePositiveHint string     `json:"falsePositiveHint"`
	Evidence          []Evidence `json:"evidence"`
	Actions           []string   `json:"actions"`
}

// ScanProgress communicates real-time scanning advancement.
type ScanProgress struct {
	Step           string  `json:"step"`
	Progress       int     `json:"progress"` // 0-100
	CurrentPackage *string `json:"currentPackage,omitempty"`
}

// ScanOptions configures scan parameters.
type ScanOptions struct {
	ScanID         string   `json:"scanId,omitempty"`
	Mode           string   `json:"mode"` // "quick" | "deep"
	TargetPackages []string `json:"targetPackages,omitempty"`
	IncludeSystem  bool     `json:"includeSystem"`
}

// ScanResult contains forensic output from a scan run.
type ScanResult struct {
	ID           string    `json:"id"`
	DeviceSerial string    `json:"deviceSerial"`
	Mode         string    `json:"mode"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
	RulesVersion string    `json:"rulesVersion"`
	AppVersion   string    `json:"appVersion"`
	Findings     []Finding `json:"findings"`
	Partial      bool      `json:"partial"`
	TotalApps    int       `json:"totalApps"`
}
