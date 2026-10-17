package rules

import (
	"bytes"
	"regexp"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadEmbeddedRules(t *testing.T) {
	rules, err := LoadEmbeddedRules()
	if err != nil {
		t.Fatalf("LoadEmbeddedRules failed: %v", err)
	}

	if len(rules) < 5 {
		t.Fatalf("expected at least 5 embedded rules, got %d", len(rules))
	}

	ruleMap := make(map[string]RuleDefinition)
	for _, r := range rules {
		ruleMap[r.ID] = r
	}

	// Verify required rules exist
	expectedRuleIDs := []string{"ADW-001", "ADW-002", "ADW-003", "STK-001", "DRO-001", "ADM-001", "BNK-001", "PUA-001"}
	for _, id := range expectedRuleIDs {
		if _, ok := ruleMap[id]; !ok {
			t.Errorf("missing expected rule: %s", id)
		}
	}
}

func TestValidateRule(t *testing.T) {
	valid := &RuleDefinition{
		ID:                "TEST-001",
		Title:             "Test Rule",
		Category:          "adware",
		Severity:          "high",
		Confidence:        "high",
		Weight:            50,
		Why:               "Test description",
		FalsePositiveHint: "Test hint",
		Remediation:       []string{"uninstall"},
		Conditions: RuleConditions{
			Installer: "sideload",
		},
	}
	if err := ValidateRule(valid); err != nil {
		t.Errorf("expected valid rule, got error: %v", err)
	}

	// Invalid category
	invalidCat := *valid
	invalidCat.Category = "malware-virus" // not in taxonomy
	if err := ValidateRule(&invalidCat); err == nil {
		t.Errorf("expected rejection for invalid category")
	}

	// Invalid severity
	invalidSev := *valid
	invalidSev.Severity = "urgent"
	if err := ValidateRule(&invalidSev); err == nil {
		t.Errorf("expected rejection for invalid severity")
	}

	// Invalid weight
	invalidWeight := *valid
	invalidWeight.Weight = 150
	if err := ValidateRule(&invalidWeight); err == nil {
		t.Errorf("expected rejection for weight > 100")
	}

	// Invalid remediation action
	invalidRem := *valid
	invalidRem.Remediation = []string{"delete_permanently"}
	if err := ValidateRule(&invalidRem); err == nil {
		t.Errorf("expected rejection for invalid remediation action")
	}

	// Empty conditions
	emptyCond := *valid
	emptyCond.Conditions = RuleConditions{}
	if err := ValidateRule(&emptyCond); err == nil {
		t.Errorf("expected rejection for rule with empty conditions")
	}
}

func TestStrictYAMLDecodingRejectsUnknownFields(t *testing.T) {
	badYAML := `
id: TEST-001
title: Unknown Field Rule
category: adware
severity: high
confidence: high
weight: 80
why: Test
false_positive_hint: Hint
remediation:
  - disable
conditions:
  installer: sideload
unknown_unsupported_field: true
`
	dec := yaml.NewDecoder(bytes.NewReader([]byte(badYAML)))
	dec.KnownFields(true)
	var rule RuleDefinition
	err := dec.Decode(&rule)
	if err == nil {
		t.Fatal("expected strict decoder to reject unknown field, but it succeeded")
	}
}

func TestRulesVersion(t *testing.T) {
	ver := RulesVersion()
	re := regexp.MustCompile(`^2026\.03\.1\+[0-9a-f]{8}$`)
	if !re.MatchString(ver) {
		t.Fatalf("expected RulesVersion matching '2026.03.1+<8hex>', got %q", ver)
	}
}

