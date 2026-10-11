# Jerion

**Android Threat Analyzer** — Desktop application for analyzing adware, aggressive popups, stalkerware, and Android threats via ADB with controlled remediation.

## Overview
Jerion connects to Android devices over USB/ADB to inspect packages, analyze permissions, audit system configurations, hunt background and overlay popup triggers, and provide evidence-based threat reports without false "safety percentages".

## Tech Stack
- **Desktop Shell**: Wails v2 (Go)
- **Frontend**: React, TypeScript (strict), Vite, Tailwind CSS, Aceternity UI (structural), Three.js (single scene)
- **Local Storage**: Pure Go SQLite (`modernc.org/sqlite`)
- **Primary Test Device**: Infinix Note 11S — XOS 10.6 (Android 12, API 31)

## Development Prerequisites
- Go 1.25+
- Node.js 20+ & npm
- Wails CLI v2 (`v2.16.0`)
- Android Platform Tools (ADB)

## Quick Start
```bash
# Run in development mode (live reload)
wails dev

# Build production binary
wails build

# Run test suites
go vet ./... && go test ./...
cd frontend && npm run lint && npm run typecheck && npm run test
```

## Documentation
- Architecture and product requirements: `docs/PRD.md`
- Active development status and phase checklist: `docs/PROGRESS.md`
- Agent and developer rules: `AGENTS.md`
- Architecture Decision Records: `docs/adr/`

## License
Proprietary — All rights reserved.
