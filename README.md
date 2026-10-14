# Jerion

**Android Threat Analyzer** — Desktop application for analyzing adware, aggressive popups, stalkerware, and Android threats via ADB with controlled remediation.

## Overview
Jerion connects to Android devices over USB/ADB to inspect packages, analyze permissions, audit system configurations, hunt background and overlay popup triggers, and provide evidence-based threat reports without false "safety percentages".

## Core Features
- **P1–P7 Launch Preflight**: Verifies ADB binary availability, ADB daemon status, pure-Go SQLite storage, ruleset packages, and OS keychain access on every launch with official Google platform-tools setup guidance.
- **Application Inventory & Inspection**: High-performance virtualized package table (`@tanstack/react-virtual`) inspecting user/system packages, requested vs granted permissions, installer sources, and APK SHA-256 fingerprints.
- **Evidence-Based Scan Engine**: Declarative YAML forensic rules evaluating overlay misuse, banking trojan patterns, boot persistence, and DEX ad SDK networks. Displays clear severity badges and technical evidence without misleading security percentage scores.
- **Popup Hunter Monitor**: Real-time background telemetry daemon tracking window focus changes and overlay events (`TYPE_APPLICATION_OVERLAY`), allowing instant candidate correlation when the user clicks "Popup Just Appeared!".
- **System Security Audit**: In-depth device hardening posture checking accessibility abusers, notification listeners, device administrators, unknown source installation, and proxy settings.
- **Safety-Guarded Remediation**: Reversible remediation engine with critical system package denylists (including Android Core and Infinix XOS 10.6 baseline apps), mandatory pre-uninstall APK backup, and full SQLite audit trails with one-click undo.
- **Offline Forensic Reports**: Self-contained HTML and JSON export with zero external CDN dependencies, embedded CSS tokens, sensitive data redaction, and scan comparison diffs.

## Tech Stack
- **Desktop Shell**: Wails v2 (`v2.16.0`)
- **Backend**: Go (Pure-Go SQLite via `modernc.org/sqlite`, zero CGO)
- **Frontend**: React 19, TypeScript (strict), Vite, Tailwind CSS, Aceternity UI, Three.js
- **Primary Hardware Baseline**: Infinix Note 11S — XOS 10.6 (Android 12, API 31)

## Development Prerequisites
- Go 1.25+
- Node.js 20+ & npm
- Wails CLI v2 (`v2.16.0`)
- Android Platform Tools (ADB)

## Live Development & Building
- **Live Dev Mode**: `wails dev`
  Runs a Vite development server with rapid hot reload of frontend changes and backend bindings.
- **Production Build**: `wails build`
  Produces a standalone native desktop executable in `build/bin/Jerion.exe`.
- **Project Configuration**: Settings are managed in `wails.json`.

## Test Suites & Quality Verification
```bash
# Backend test suite (all 16 packages)
go vet ./... && go test -count=1 ./...

# Frontend test suite, linting, typecheck, and dead-code verification
cd frontend
npm run lint
npm run typecheck
npm run test
npm run check:no-mock
```

## Documentation
- Product Requirements Document (PRD): `docs/PRD.md`
- Active development status and phase checklist: `docs/PROGRESS.md`
- Agent and developer rules: `AGENTS.md`
- Device baseline and testing notes: `docs/devices/infinix-note-11s.md`
- Architecture Decision Records: `docs/adr/`

## License
Proprietary — All rights reserved.
