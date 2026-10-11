# ADR-0003: English-Only UI with Mandatory i18n Abstraction

## Status
Accepted

## Date
2026-10-11

## Context
Jerion's initial release (v1.0) targets technical service professionals, students, and general users seeking clear, factual threat analysis. Maintaining multiple UI translations in v1 increases review surface area and risks inconsistent security terminology. However, hardcoding English strings directly into React components makes future localization costly and error-prone.

## Decisions

### 1. English-Only for v1.0
- The application user interface, notification copy, error states, and exported reports will be in **English only** in v1.0.
- Copywriting must be factual, calm, and specific (e.g., "Installed from outside official store 3 days ago"), avoiding hype ("AI-powered ultimate defense") and avoid vague "safety percentages".

### 2. Mandatory i18n Abstraction Layer
- All user-facing strings must pass through `react-i18next` backed by a centralized `frontend/src/i18n/en.json` file.
- Direct hardcoded user-facing strings in TSX components are strictly prohibited.
- Adding additional language packs in future versions will require zero component refactoring.

## Consequences
- Single, consistent, and strictly reviewed security vocabulary.
- Seamless localization readiness for future phases without structural changes to UI code.
