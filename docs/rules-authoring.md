# Rules Authoring Guide

All Jerion threat detection rules are written as declarative YAML files stored in `internal/rules/*.yaml` and embedded into the application binary.

## Rule Schema Requirements
Every detection rule must include:
- `id`: Unique identifier (e.g., `ADW-001`, `STK-002`, `DRO-003`).
- `title`: Short factual description.
- `category`: Fixed taxonomy (`adware`, `aggressive-ads`, `stalkerware`, `banking-trojan-indicators`, `dropper`, `device-admin-abuse`, `system-tampering`, `pua`, `suspicious-install`, `info`).
- `severity`: `info | low | medium | high | critical`
- `confidence`: `low | medium | high`
- `weight`: Integer weight for scoring caps.
- `why`: 1–2 factual sentences explaining why this signal was flagged.
- `evidence`: Raw output keys extracted during analysis.
- `false_positive_hint`: Realistic scenarios where this might be legitimate.
- `remediation`: Recommended user action.

## Testing Requirement (AGENTS §9)
Rules without automated tests will be rejected:
- Every rule must have both a **positive test case** (triggers the rule) and a **negative test case** (verifies that benign apps do not trigger false positives).
