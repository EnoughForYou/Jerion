package engine

import (
	"fmt"
	"strings"

	"github.com/EnoughForYou/Jerion/internal/apk"
	"github.com/EnoughForYou/Jerion/internal/inventory"
	"github.com/EnoughForYou/Jerion/internal/rules"
)

// Evaluator checks application attributes against declarative rules.
type Evaluator struct {
	rules []rules.RuleDefinition
}

// NewEvaluator creates an Evaluator loaded with detection rules.
func NewEvaluator(ruleDefs []rules.RuleDefinition) *Evaluator {
	return &Evaluator{rules: ruleDefs}
}

// EvaluateApp checks an individual app (and optional APK binary detail) against all loaded rules.
func (e *Evaluator) EvaluateApp(app inventory.AppInfo, apkDetail *apk.APKDetail) []Finding {
	var findings []Finding

	for _, r := range e.rules {
		if finding, matched := matchRule(app, apkDetail, r); matched {
			findings = append(findings, *finding)
		}
	}

	return findings
}

func matchRule(app inventory.AppInfo, apkDetail *apk.APKDetail, r rules.RuleDefinition) (*Finding, bool) {
	cond := r.Conditions

	// 1. IsSystem condition
	if cond.IsSystem != nil {
		if *cond.IsSystem != app.IsSystem {
			return nil, false
		}
	}

	// 2. HasLauncher condition
	if cond.HasLauncher != nil {
		if *cond.HasLauncher != app.HasLauncher {
			return nil, false
		}
	}

	// 3. Installer condition
	if cond.Installer != "" {
		if cond.Installer == "sideload" {
			if app.Installer != nil && *app.Installer == "com.android.vending" {
				return nil, false // Official store install is not sideloaded
			}
		} else if app.Installer == nil || *app.Installer != cond.Installer {
			return nil, false
		}
	}

	// Gather all known permissions from dumpsys + apk manifest
	allPerms := make(map[string]bool)
	for _, p := range app.RequestedPermissions {
		allPerms[p] = true
	}
	for _, p := range app.GrantedPermissions {
		allPerms[p] = true
	}
	if apkDetail != nil {
		for _, p := range apkDetail.Manifest.Permissions {
			allPerms[p] = true
		}
	}

	var evidenceList []Evidence

	// 4. RequiredPermissions condition (all must match)
	for _, req := range cond.RequiredPermissions {
		if !allPerms[req] {
			return nil, false
		}
		evidenceList = append(evidenceList, Evidence{
			Source:  "permission",
			Snippet: fmt.Sprintf("Declared/Requested: %s", req),
		})
	}

	// 5. AnyPermissions condition (at least one must match)
	if len(cond.AnyPermissions) > 0 {
		matchedAny := false
		for _, anyP := range cond.AnyPermissions {
			if allPerms[anyP] {
				matchedAny = true
				evidenceList = append(evidenceList, Evidence{
					Source:  "permission",
					Snippet: fmt.Sprintf("Declared/Requested: %s", anyP),
				})
				break
			}
		}
		if !matchedAny {
			return nil, false
		}
	}

	// 6. ForbiddenPermissions condition (none may match)
	for _, forb := range cond.ForbiddenPermissions {
		if allPerms[forb] {
			return nil, false
		}
	}

	// 7. Ad SDK conditions
	if len(cond.AdSDKs) > 0 || cond.MinAdSDKs > 0 {
		if apkDetail == nil {
			return nil, false // Ad SDK rules require APK inspection
		}

		matchedSDKCount := 0
		for _, targetSDK := range cond.AdSDKs {
			if label, ok := apkDetail.DetectedAdSDKs[targetSDK]; ok {
				matchedSDKCount++
				evidenceList = append(evidenceList, Evidence{
					Source:  "dex",
					Snippet: fmt.Sprintf("Embedded Ad SDK: %s (%s)", label, targetSDK),
				})
			}
		}

		if cond.MinAdSDKs > 0 {
			if len(apkDetail.DetectedAdSDKs) < cond.MinAdSDKs {
				return nil, false
			}
			for sig, label := range apkDetail.DetectedAdSDKs {
				evidenceList = append(evidenceList, Evidence{
					Source:  "dex",
					Snippet: fmt.Sprintf("Embedded Ad SDK: %s (%s)", label, sig),
				})
			}
		} else if len(cond.AdSDKs) > 0 && matchedSDKCount == 0 {
			return nil, false
		}
	}

	// If no specific evidence collected yet, record general reason
	if len(evidenceList) == 0 {
		evidenceList = append(evidenceList, Evidence{
			Source:  "package",
			Snippet: fmt.Sprintf("Package: %s (UID: %d)", app.Package, app.UID),
		})
	}

	finding := &Finding{
		ID:                fmt.Sprintf("f-%s-%s", strings.ToLower(r.ID), app.Package),
		RuleID:            r.ID,
		Package:           app.Package,
		Category:          r.Category,
		Severity:          r.Severity,
		Confidence:        r.Confidence,
		Weight:            r.Weight,
		Title:             r.Title,
		WhyFlagged:        r.Why,
		FalsePositiveHint: r.FalsePositiveHint,
		Evidence:          evidenceList,
		Actions:           r.Remediation,
	}

	return finding, true
}
