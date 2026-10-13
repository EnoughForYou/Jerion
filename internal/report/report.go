package report

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/EnoughForYou/Jerion/internal/engine"
	"github.com/EnoughForYou/Jerion/internal/store"
)

// DeviceInfo represents hardware profile for report generation.
type DeviceInfo struct {
	Serial       string `json:"serial"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	AndroidVer   string `json:"androidVersion"`
	SDKInt       int    `json:"sdkInt"`
}

// ScanSummaryItem represents a lightweight scan header.
type ScanSummaryItem struct {
	ID            string         `json:"id"`
	StartedAt     string         `json:"startedAt"`
	CompletedAt   *string        `json:"completedAt,omitempty"`
	Mode          string         `json:"mode"`
	Verdict       string         `json:"verdict"`
	FindingsCount map[string]int `json:"findingsCount"`
}

// ScanComparison represents differences between two scan executions.
type ScanComparison struct {
	ScanA             ScanSummaryItem  `json:"scanA"`
	ScanB             ScanSummaryItem  `json:"scanB"`
	NewFindings       []store.FindingRecord `json:"newFindings"`
	ResolvedFindings  []store.FindingRecord `json:"resolvedFindings"`
	UnchangedFindings []store.FindingRecord `json:"unchangedFindings"`
}

// CalculateVerdict determines high_risk | review | clean based on finding counts.
func CalculateVerdict(findings []store.FindingRecord) (string, map[string]int) {
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, f := range findings {
		sev := strings.ToLower(f.Severity)
		counts[sev]++
	}

	verdict := "clean"
	if counts["critical"] > 0 || counts["high"] > 0 {
		verdict = "high_risk"
	} else if counts["medium"] > 0 {
		verdict = "review"
	}

	return verdict, counts
}

// BuildSummaryItem creates a ScanSummaryItem from a store.ScanRecord.
func BuildSummaryItem(scan store.ScanRecord) ScanSummaryItem {
	verdict, counts := CalculateVerdict(scan.Findings)
	var compStr *string
	if !scan.FinishedAt.IsZero() {
		s := scan.FinishedAt.UTC().Format(time.RFC3339)
		compStr = &s
	}

	return ScanSummaryItem{
		ID:            scan.ID,
		StartedAt:     scan.StartedAt.UTC().Format(time.RFC3339),
		CompletedAt:   compStr,
		Mode:          scan.Mode,
		Verdict:       verdict,
		FindingsCount: counts,
	}
}

// CompareScans compares findings between two scan executions.
func CompareScans(scanA, scanB store.ScanRecord) ScanComparison {
	itemA := BuildSummaryItem(scanA)
	itemB := BuildSummaryItem(scanB)

	mapA := make(map[string]store.FindingRecord)
	for _, f := range scanA.Findings {
		key := fmt.Sprintf("%s:%s", f.RuleID, f.Package)
		mapA[key] = f
	}

	mapB := make(map[string]store.FindingRecord)
	for _, f := range scanB.Findings {
		key := fmt.Sprintf("%s:%s", f.RuleID, f.Package)
		mapB[key] = f
	}

	var newFindings []store.FindingRecord
	var resolvedFindings []store.FindingRecord
	var unchangedFindings []store.FindingRecord

	// Scan B is newer than Scan A
	for key, f := range mapB {
		if _, exists := mapA[key]; exists {
			unchangedFindings = append(unchangedFindings, f)
		} else {
			newFindings = append(newFindings, f)
		}
	}

	for key, f := range mapA {
		if _, exists := mapB[key]; !exists {
			resolvedFindings = append(resolvedFindings, f)
		}
	}

	return ScanComparison{
		ScanA:             itemA,
		ScanB:             itemB,
		NewFindings:       newFindings,
		ResolvedFindings:  resolvedFindings,
		UnchangedFindings: unchangedFindings,
	}
}

// ExportJSON generates formatted JSON for a scan result with optional redacting.
func ExportJSON(scan store.ScanRecord, redact bool) ([]byte, error) {
	exportScan := scan
	if redact {
		exportScan.DeviceSerial = "REDACTED"
	}
	return json.MarshalIndent(exportScan, "", "  ")
}

// ExportHTML produces a 100% self-contained offline HTML audit report.
func ExportHTML(scan store.ScanRecord, device DeviceInfo, redact bool) (string, error) {
	verdict, counts := CalculateVerdict(scan.Findings)

	serial := device.Serial
	if redact {
		serial = "REDACTED"
	}

	verdictColor := "#10b981"
	verdictText := "CLEAN"
	if verdict == "high_risk" {
		verdictColor = "#ef4444"
		verdictText = "HIGH RISK"
	} else if verdict == "review" {
		verdictColor = "#f59e0b"
		verdictText = "NEEDS REVIEW"
	}

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Jerion Forensic Threat Report - ` + html.EscapeString(scan.ID) + `</title>
<style>
  :root {
    --bg-page: #090a0f;
    --bg-card: #131722;
    --border: #232936;
    --text-primary: #f3f4f6;
    --text-secondary: #9ca3af;
    --accent: #3b82f6;
    --critical: #ef4444;
    --high: #f97316;
    --medium: #f59e0b;
    --low: #3b82f6;
    --info: #6b7280;
  }
  body {
    background-color: var(--bg-page);
    color: var(--text-primary);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    margin: 0;
    padding: 32px 20px;
    line-height: 1.5;
  }
  .container { max-width: 960px; margin: 0 auto; }
  .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); padding-bottom: 24px; margin-bottom: 28px; }
  .title { font-size: 24px; font-weight: 700; margin: 0; }
  .subtitle { color: var(--text-secondary); font-size: 14px; margin-top: 4px; }
  .badge { padding: 4px 12px; border-radius: 9999px; font-size: 13px; font-weight: 600; text-transform: uppercase; }
  .card-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 28px; }
  .card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 8px; padding: 16px; }
  .card-label { font-size: 12px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; margin-bottom: 6px; }
  .card-value { font-size: 18px; font-weight: 600; }
  .counts-row { display: flex; gap: 12px; margin-bottom: 28px; flex-wrap: wrap; }
  .count-pill { background: var(--bg-card); border: 1px solid var(--border); border-radius: 6px; padding: 8px 16px; font-size: 14px; display: flex; align-items: center; gap: 8px; }
  .findings-title { font-size: 18px; font-weight: 600; margin-bottom: 16px; }
  .finding-item { background: var(--bg-card); border: 1px solid var(--border); border-radius: 8px; padding: 20px; margin-bottom: 16px; }
  .finding-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
  .finding-name { font-size: 16px; font-weight: 600; }
  .finding-pkg { font-family: monospace; font-size: 13px; color: var(--text-secondary); margin-top: 2px; }
  .finding-why { font-size: 14px; color: var(--text-primary); margin-bottom: 12px; }
  .evidence-box { background: rgba(0,0,0,0.4); border-left: 3px solid var(--accent); padding: 10px 14px; border-radius: 4px; font-family: monospace; font-size: 12px; overflow-x: auto; margin-top: 8px; color: #d1d5db; }
  .footer { margin-top: 48px; border-top: 1px solid var(--border); padding-top: 16px; font-size: 12px; color: var(--text-secondary); text-align: center; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <div>
      <h1 class="title">Jerion Threat Analysis Report</h1>
      <div class="subtitle">Generated: ` + html.EscapeString(time.Now().UTC().Format(time.RFC3339)) + ` • Scan ID: ` + html.EscapeString(scan.ID) + `</div>
    </div>
    <div>
      <span class="badge" style="background: ` + verdictColor + `22; color: ` + verdictColor + `; border: 1px solid ` + verdictColor + `;">` + html.EscapeString(verdictText) + `</span>
    </div>
  </div>

  <div class="card-grid">
    <div class="card">
      <div class="card-label">Target Device</div>
      <div class="card-value">` + html.EscapeString(device.Manufacturer+" "+device.Model) + `</div>
      <div style="font-size: 12px; color: var(--text-secondary); margin-top: 4px;">Serial: ` + html.EscapeString(serial) + `</div>
    </div>
    <div class="card">
      <div class="card-label">Android OS</div>
      <div class="card-value">Android ` + html.EscapeString(device.AndroidVer) + `</div>
      <div style="font-size: 12px; color: var(--text-secondary); margin-top: 4px;">API Level ` + fmt.Sprintf("%d", device.SDKInt) + `</div>
    </div>
    <div class="card">
      <div class="card-label">Scan Profile</div>
      <div class="card-value" style="text-transform: capitalize;">` + html.EscapeString(scan.Mode) + ` Scan</div>
      <div style="font-size: 12px; color: var(--text-secondary); margin-top: 4px;">Ruleset ` + html.EscapeString(scan.RulesVersion) + `</div>
    </div>
  </div>

  <div class="counts-row">
    <div class="count-pill"><strong style="color: var(--critical);">` + fmt.Sprintf("%d", counts["critical"]) + `</strong> Critical</div>
    <div class="count-pill"><strong style="color: var(--high);">` + fmt.Sprintf("%d", counts["high"]) + `</strong> High</div>
    <div class="count-pill"><strong style="color: var(--medium);">` + fmt.Sprintf("%d", counts["medium"]) + `</strong> Medium</div>
    <div class="count-pill"><strong style="color: var(--low);">` + fmt.Sprintf("%d", counts["low"]) + `</strong> Low</div>
    <div class="count-pill"><strong style="color: var(--info);">` + fmt.Sprintf("%d", counts["info"]) + `</strong> Info</div>
  </div>

  <div class="findings-title">Identified Forensic Findings (` + fmt.Sprintf("%d", len(scan.Findings)) + `)</div>
`)

	if len(scan.Findings) == 0 {
		sb.WriteString(`
  <div class="card" style="text-align: center; padding: 40px; color: var(--text-secondary);">
    No adware, suspicious overlays, or malicious indicators detected during this scan.
  </div>
`)
	} else {
		for _, f := range scan.Findings {
			sevColor := "var(--info)"
			switch strings.ToLower(f.Severity) {
			case "critical":
				sevColor = "var(--critical)"
			case "high":
				sevColor = "var(--high)"
			case "medium":
				sevColor = "var(--medium)"
			case "low":
				sevColor = "var(--low)"
			}

			sb.WriteString(`
  <div class="finding-item">
    <div class="finding-header">
      <div>
        <div class="finding-name">` + html.EscapeString(f.Title) + `</div>
        <div class="finding-pkg">` + html.EscapeString(f.Package) + ` (` + html.EscapeString(f.RuleID) + `)</div>
      </div>
      <div>
        <span class="badge" style="background: ` + sevColor + `22; color: ` + sevColor + `; font-size: 11px;">` + html.EscapeString(strings.ToUpper(f.Severity)) + `</span>
      </div>
    </div>
    <div class="finding-why">` + html.EscapeString(f.WhyFlagged) + `</div>
`)
			for _, ev := range f.Evidence {
				sb.WriteString(`
    <div class="evidence-box"><strong>` + html.EscapeString(ev.Source) + `:</strong> ` + html.EscapeString(ev.Snippet) + `</div>
`)
			}

			sb.WriteString(`
  </div>
`)
		}
	}

	sb.WriteString(`
  <div class="footer">
    Report produced by Jerion Android Threat Analyzer • Offline Forensic Engine
  </div>
</div>
</body>
</html>
`)

	return sb.String(), nil
}

// ConvertEngineFindingsToRecords bridges engine findings into store format.
func ConvertEngineFindingsToRecords(findings []engine.Finding) []store.FindingRecord {
	var records []store.FindingRecord
	for _, f := range findings {
		var evs []store.EvidenceRecord
		for _, e := range f.Evidence {
			evs = append(evs, store.EvidenceRecord{
				Source:  e.Source,
				Snippet: e.Snippet,
			})
		}
		records = append(records, store.FindingRecord{
			ID:                f.ID,
			RuleID:            f.RuleID,
			Package:           f.Package,
			Category:          f.Category,
			Severity:          f.Severity,
			Confidence:        f.Confidence,
			Weight:            f.Weight,
			Title:             f.Title,
			WhyFlagged:        f.WhyFlagged,
			FalsePositiveHint: f.FalsePositiveHint,
			Evidence:          evs,
			Actions:           f.Actions,
		})
	}
	return records
}
