package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EnoughForYou/Jerion/internal/report"
	"github.com/EnoughForYou/Jerion/internal/store"
)

func TestVerdictCalculation(t *testing.T) {
	// Clean
	v1, _ := report.CalculateVerdict([]store.FindingRecord{})
	if v1 != "clean" {
		t.Fatalf("expected clean, got %s", v1)
	}

	// Review (medium)
	v2, _ := report.CalculateVerdict([]store.FindingRecord{
		{Severity: "medium"},
	})
	if v2 != "review" {
		t.Fatalf("expected review, got %s", v2)
	}

	// High risk (high/critical)
	v3, counts := report.CalculateVerdict([]store.FindingRecord{
		{Severity: "low"},
		{Severity: "high"},
	})
	if v3 != "high_risk" {
		t.Fatalf("expected high_risk, got %s", v3)
	}
	if counts["high"] != 1 || counts["low"] != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}

func TestExportJSON(t *testing.T) {
	scan := store.ScanRecord{
		ID:           "scan-123",
		DeviceSerial: "ABCDEF123456",
		Mode:         "quick",
		StartedAt:    time.Now().UTC(),
		Findings: []store.FindingRecord{
			{
				ID:       "f-1",
				RuleID:   "ADW-001",
				Package:  "com.spam.ad",
				Severity: "high",
				Title:    "Adware Overlay Detected",
			},
		},
	}

	// Non-redacted
	jsonBytes, err := report.ExportJSON(scan, false)
	if err != nil {
		t.Fatalf("export json failed: %v", err)
	}
	if !strings.Contains(string(jsonBytes), "ABCDEF123456") {
		t.Fatalf("expected serial in unredacted json")
	}

	// Redacted
	redactedBytes, err := report.ExportJSON(scan, true)
	if err != nil {
		t.Fatalf("export json redacted failed: %v", err)
	}
	if strings.Contains(string(redactedBytes), "ABCDEF123456") {
		t.Fatalf("serial should be redacted")
	}
	if !strings.Contains(string(redactedBytes), "REDACTED") {
		t.Fatalf("expected REDACTED placeholder")
	}
}

func TestExportHTML(t *testing.T) {
	scan := store.ScanRecord{
		ID:           "scan-html-test",
		DeviceSerial: "SER12345",
		Mode:         "deep",
		StartedAt:    time.Now().UTC(),
		RulesVersion: "1.0.0",
		Findings: []store.FindingRecord{
			{
				ID:         "f-2",
				RuleID:     "STK-001",
				Package:    "com.stalk.track",
				Severity:   "critical",
				Title:      "Stalkerware Service",
				WhyFlagged: "Persistent hidden service running in background",
				Evidence: []store.EvidenceRecord{
					{Source: "manifest", Snippet: "RECEIVER_BOOT_COMPLETED"},
				},
			},
		},
	}

	dev := report.DeviceInfo{
		Serial:       "SER12345",
		Model:        "Note 11S",
		Manufacturer: "Infinix",
		AndroidVer:   "12",
		SDKInt:       31,
	}

	htmlContent, err := report.ExportHTML(scan, dev, false)
	if err != nil {
		t.Fatalf("export html failed: %v", err)
	}

	// Assertions for offline self-contained HTML
	if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
		t.Fatalf("expected html doctype")
	}
	if strings.Contains(htmlContent, "http://") || strings.Contains(htmlContent, "https://") {
		t.Fatalf("offline html must not contain external http/https CDN references")
	}
	if !strings.Contains(htmlContent, "HIGH RISK") {
		t.Fatalf("expected HIGH RISK verdict badge in HTML")
	}
	if !strings.Contains(htmlContent, "Infinix Note 11S") {
		t.Fatalf("expected device model in HTML")
	}
	if !strings.Contains(htmlContent, "com.stalk.track") {
		t.Fatalf("expected package name in HTML")
	}
}

func TestCompareScans(t *testing.T) {
	scan1 := store.ScanRecord{
		ID:        "scan-1",
		StartedAt: time.Now().UTC().Add(-1 * time.Hour),
		Findings: []store.FindingRecord{
			{RuleID: "ADW-001", Package: "com.spam.ad", Title: "Old Finding"},
			{RuleID: "ADW-002", Package: "com.cleaned.app", Title: "App to be cleaned"},
		},
	}

	scan2 := store.ScanRecord{
		ID:        "scan-2",
		StartedAt: time.Now().UTC(),
		Findings: []store.FindingRecord{
			{RuleID: "ADW-001", Package: "com.spam.ad", Title: "Old Finding"},
			{RuleID: "STK-001", Package: "com.new.threat", Title: "New Threat"},
		},
	}

	comp := report.CompareScans(scan1, scan2)
	if len(comp.UnchangedFindings) != 1 {
		t.Fatalf("expected 1 unchanged finding, got %d", len(comp.UnchangedFindings))
	}
	if len(comp.NewFindings) != 1 || comp.NewFindings[0].Package != "com.new.threat" {
		t.Fatalf("expected 1 new finding (com.new.threat), got %+v", comp.NewFindings)
	}
	if len(comp.ResolvedFindings) != 1 || comp.ResolvedFindings[0].Package != "com.cleaned.app" {
		t.Fatalf("expected 1 resolved finding (com.cleaned.app), got %+v", comp.ResolvedFindings)
	}
}
