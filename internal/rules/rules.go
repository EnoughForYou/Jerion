package rules

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var embeddedRulesFS embed.FS

// RuleConditions specifies matching criteria against an application package or binary.
type RuleConditions struct {
	IsSystem             *bool    `yaml:"is_system,omitempty"`
	HasLauncher          *bool    `yaml:"has_launcher,omitempty"`
	Installer            string   `yaml:"installer,omitempty"` // "sideload" or package name
	RequiredPermissions []string `yaml:"required_permissions,omitempty"`
	AnyPermissions      []string `yaml:"any_permissions,omitempty"`
	ForbiddenPermissions []string `yaml:"forbidden_permissions,omitempty"`
	AdSDKs               []string `yaml:"ad_sdks,omitempty"`
	MinAdSDKs            int      `yaml:"min_ad_sdks,omitempty"`
}

// RuleDefinition encapsulates declarative threat detection metadata.
type RuleDefinition struct {
	ID                string         `yaml:"id"`
	Title             string         `yaml:"title"`
	Category          string         `yaml:"category"`
	Severity          string         `yaml:"severity"`
	Confidence        string         `yaml:"confidence"`
	Weight            int            `yaml:"weight"`
	Why               string         `yaml:"why"`
	FalsePositiveHint string         `yaml:"false_positive_hint"`
	Remediation       []string       `yaml:"remediation"`
	Conditions        RuleConditions `yaml:"conditions"`
}

var validCategories = map[string]bool{
	"adware":                     true,
	"aggressive-ads":             true,
	"stalkerware":                true,
	"banking-trojan-indicators":  true,
	"dropper":                    true,
	"device-admin-abuse":         true,
	"system-tampering":           true,
	"pua":                        true,
	"suspicious-install":         true,
	"info":                       true,
}

var validSeverities = map[string]bool{
	"info":     true,
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

var validConfidences = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

var validRemediations = map[string]bool{
	"disable":           true,
	"uninstall":         true,
	"revoke_permission": true,
	"clear_appop":       true,
}

// HasAnyCondition returns true if at least one matching criterion is defined.
func (c *RuleConditions) HasAnyCondition() bool {
	return c.IsSystem != nil ||
		c.HasLauncher != nil ||
		c.Installer != "" ||
		len(c.RequiredPermissions) > 0 ||
		len(c.AnyPermissions) > 0 ||
		len(c.ForbiddenPermissions) > 0 ||
		len(c.AdSDKs) > 0 ||
		c.MinAdSDKs > 0
}

// ValidateRule verifies compliance with rules authoring specifications.
func ValidateRule(r *RuleDefinition) error {
	if r.ID == "" {
		return fmt.Errorf("rule missing id")
	}
	if r.Title == "" {
		return fmt.Errorf("rule %s missing title", r.ID)
	}
	if !validCategories[r.Category] {
		return fmt.Errorf("rule %s has invalid category: %s", r.ID, r.Category)
	}
	if !validSeverities[r.Severity] {
		return fmt.Errorf("rule %s has invalid severity: %s", r.ID, r.Severity)
	}
	if !validConfidences[r.Confidence] {
		return fmt.Errorf("rule %s has invalid confidence: %s", r.ID, r.Confidence)
	}
	if r.Weight < 1 || r.Weight > 100 {
		return fmt.Errorf("rule %s has invalid weight: %d (must be 1-100)", r.ID, r.Weight)
	}
	if r.Why == "" {
		return fmt.Errorf("rule %s missing why description", r.ID)
	}
	if r.FalsePositiveHint == "" {
		return fmt.Errorf("rule %s missing false_positive_hint", r.ID)
	}
	if len(r.Remediation) == 0 {
		return fmt.Errorf("rule %s must define at least one remediation action", r.ID)
	}
	for _, act := range r.Remediation {
		if !validRemediations[act] {
			return fmt.Errorf("rule %s has invalid remediation action: %s", r.ID, act)
		}
	}
	if !r.Conditions.HasAnyCondition() {
		return fmt.Errorf("rule %s must define at least one matching condition", r.ID)
	}
	return nil
}

// LoadEmbeddedRules reads and parses all YAML rules bundled with the binary.
func LoadEmbeddedRules() ([]RuleDefinition, error) {
	var rules []RuleDefinition
	seenIDs := make(map[string]bool)

	err := fs.WalkDir(embeddedRulesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".yaml") {
			return nil
		}

		data, err := fs.ReadFile(embeddedRulesFS, path)
		if err != nil {
			return fmt.Errorf("read rule file %s: %w", path, err)
		}

		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		var rule RuleDefinition
		if err := dec.Decode(&rule); err != nil {
			return fmt.Errorf("unmarshal rule %s: %w", path, err)
		}

		if err := ValidateRule(&rule); err != nil {
			return fmt.Errorf("invalid rule %s: %w", path, err)
		}

		if seenIDs[rule.ID] {
			return fmt.Errorf("duplicate rule ID: %s in %s", rule.ID, path)
		}
		seenIDs[rule.ID] = true

		rules = append(rules, rule)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return rules, nil
}

// RulesVersion defines the canonical version identifier for the bundled threat rules catalog.
const RulesVersion = "2026.03.1"

// Count returns the dynamic count of valid threat rules loaded from the embedded catalog.
func Count() int {
	loaded, err := LoadEmbeddedRules()
	if err != nil {
		return 0
	}
	return len(loaded)
}

