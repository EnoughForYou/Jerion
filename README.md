# Jerion

**Android Threat Analyzer**
A desktop forensic utility designed to identify adware, intrusive popups, stalkerware, and malicious application behaviors over ADB with controlled, reversible remediation.

---

## Executive Summary

Most mobile security utilities rely on arbitrary "protection scores" or decorative scanning animations that offer little transparency. Jerion takes a fundamentally different engineering approach. Built on top of Android Debug Bridge (ADB), it treats threat analysis as a rigorous forensic process. It extracts evidence directly from runtime package managers, window managers, and application binaries, presenting verifiable facts categorized by severity and confidence.

Jerion operates on three foundational principles:
1. Every flagged threat includes raw evidence snippets, attribution rules, and false-positive context.
2. No device changes occur without explicit user confirmation and automated package backups.
3. Zero telemetry, zero external CDNs, and zero third-party tracking ensure complete data privacy.

---

## Why Jerion?

### Evidence over Percentages
Jerion never assigns a blanket "safety percentage" to a device. Threat verdicts reflect tangible findings: overlay window abuse, hidden launcher activities, accessibility service exploitation, persistent boot receivers, or known tracking SDKs embedded inside DEX bytecode.

### Non-Root Forensic Access
Rooting introduces security risks and voids device integrity. Jerion extracts deep diagnostic information entirely over standard Android developer interfaces using ADB, making advanced mobile forensics accessible without bootloader unlocks.

### Calm Desktop Engineering
The user experience reflects thoughtful desktop design principles inspired by macOS aesthetics and utilities like 3uTools. Data is displayed densely in structured cards, high-performance virtualized tables, and legible typography, avoiding decorative gradients, glowing effects, and visual noise.

---

## Core Capabilities

### 1. Launch Preflight Diagnostic
Every launch executes an automated seven-point environment check before opening the primary workspace. The preflight subsystem validates ADB binary existence, starts or verifies local ADB server daemons, establishes local SQLite databases, loads embedded rule packs, and tests OS keychain accessibility. If a component is missing, users receive clear operational guidance with official Google platform-tools setup options.

### 2. Forensic Scan Engine
Jerion features both Quick and Deep scan pipelines powered by declarative YAML detection rules. The engine combines multiple forensic sources:

**Binary Manifest Parser**
A native pure-Go AXML decoder extracts permissions, activities, services, receivers, and intent filters directly from compressed Android APK manifests without requiring Java runtime dependencies.

**DEX Bytecode Scanner**
Deep scans inspect DEX header structures and type identifier tables to recognize signatures from over 11 aggressive ad networks and tracking frameworks.

**Certificate Validation**
Extracts SHA-256 fingerprints from APK signing blocks to establish publisher identity.

### 3. Popup Hunter & Focus Telemetry
Hunting aggressive background popups and out-of-context advertisements is notoriously difficult because bad actors hide until the user navigates elsewhere. Jerion solves this through real-time telemetry.

A lightweight background monitor tracks window state transitions, focus shifts, and overlay layers (`TYPE_APPLICATION_OVERLAY`) using `dumpsys window`. When an unsolicited popup appears on the physical device, the user clicks **Popup Just Appeared!** in Jerion. The system immediately cross-references the exact timestamp with recorded window events, calculating time deltas to rank the most probable candidate applications.

### 4. High-Performance Package Inventory
Modern Android devices often run 400 to 600 packages across system and user partitions. Jerion implements a virtualized table powered by `@tanstack/react-virtual` capable of rendering thousands of items at consistent 60 FPS. Packages can be filtered by partition (User, System, Risky) or searched by package name and label. The detail drawer reveals requested versus granted permissions, installer identity, APK split configurations, and direct export actions.

### 5. System Hardening Audit
Beyond installed applications, overall system configuration dictates device vulnerability. Jerion audits critical hardening parameters including accessibility service holders, notification listeners, active Device Administrator privileges, non-market package installation toggles (`install_non_market_apps`), USB debugging exposure, global HTTP proxies, and Private DNS configuration.

### 6. Guardrailed Remediation & Rollback
Taking action against suspicious applications requires strict safeguards.

**Critical System Denylist**: Core Android packages (`com.android.systemui`, `android`) and Transsion OEM baseline packages (`com.transsion.launcher`, `com.transsion.systemupdate`, `com.transsion.phonemaster`, `com.transsion.smartpanel`) are protected against accidental removal or disabling.

**Automated APK Backup**: Before any package is uninstalled, Jerion pulls the base APK to local storage.

**Reversible Actions Log**: Every modification is logged in SQLite with one-click undo support (`enable`, `install-existing`, `grant`).

### 7. Self-Contained Offline Reports
Forensic investigations require shareable documentation. Jerion generates standalone HTML and JSON audit reports. Exported HTML files embed all styles, typography, and SVG assets inline, allowing complete offline inspection without network connectivity or remote CDN dependencies. Reports support automated data redaction for sensitive identifiers like device serials, IMEIs, and MAC addresses.

---

## Technical Specifications

| Subsystem | Technology | Architectural Rationale |
|---|---|---|
| Desktop Shell | **Wails v2** (`v2.16.0`) | Lightweight native desktop shell with direct Go-to-TypeScript bindings. |
| Backend Runtime | **Go 1.25+** | Concurrency safety, fast binary parsing, and single-file portability. |
| Database Engine | **Pure-Go SQLite** (`modernc.org/sqlite`) | Zero CGO compilation requirement with full WAL-mode persistence. |
| Credentials | **OS Keychain** (`zalando/go-keyring`) | Write-only secrets handling with zero plaintext token exposure. |
| User Interface | **React 19 + TypeScript (Strict)** | Component modularity, full type safety, and strict contract fidelity. |
| Styling | **Tailwind CSS + Tokens** | Single 4px grid token system with full dark and light mode support. |
| Typography | **Local Inter Variable** | Bundled font assets without remote web font downloads. |
| Virtualization | **TanStack Virtual** | Smooth rendering of extensive application inventories. |
| 3D Visualization | **Three.js + R3F** | Procedural hardware visualization that automatically pauses on blur. |

---

## Getting Started

### Development Prerequisites
Ensure your local development environment has the following toolchains installed:
1. Go 1.25 or higher
2. Node.js 22 LTS with npm
3. Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`)
4. Android Platform Tools (`adb`) installed and available in system PATH

### Running in Development Mode
Start live development mode with Vite hot module replacement and live Go method bindings:

```bash
wails dev
```

The desktop window opens immediately. The Vite development server also serves assets locally at `http://localhost:34115`, enabling standard browser devtools.

### Building for Production
Compile a standalone, optimized Windows x64 binary with embedded assets:

```bash
wails build
```

The resulting executable is generated at `build/bin/Jerion.exe`. For Windows installer distribution, project configurations in `build/windows/installer/project.nsi` provide automated WebView2 bootstrapper packaging and clean uninstallation routines.

---

## Quality & Verification Gates

Jerion enforces strict testing and verification criteria across the entire codebase. Every stage must pass all gates before release:

```bash
# 1. Run complete Go backend verification across all 16 internal packages
go vet ./... && go test -count=1 ./...

# 2. Run frontend linting, TypeScript compiler, and Vitest test suite
cd frontend
npm run lint
npm run typecheck
npm run test

# 3. Ensure zero mock artifacts or debug schemas exist in the production bundle
npm run check:no-mock
```

---

## Verified Hardware Baseline

Jerion is verified against real hardware to ensure parser stability across vendor customizations.

| Attribute | Configuration |
|---|---|
| Primary Test Device | Infinix Note 11S (Model X698) |
| Operating System | XOS 10.6 on Android 12 (API Level 31) |
| Protected OEM Packages | Transsion launcher, system update, smart panel, and phone master |
| Parser Behavior | Defensive extraction with graceful "Not Available" fallbacks |

For detailed hardware setup instructions and USB debugging procedures, refer to [`docs/devices/infinix-note-11s.md`](docs/devices/infinix-note-11s.md).

---

## Privacy & Operational Ethics

Jerion is designed for security audits on authorized devices:
1. **No Data Extraction**: The tool does not extract SMS messages, photos, call records, contacts, or location coordinates.
2. **No Root Exploits**: Operates entirely within standard developer permissions.
3. **Local Hash Lookups**: When threat intelligence integrations are active, only SHA-256 hashes are evaluated. File uploads remain strictly opt-in per individual binary.
4. **Key Isolation**: External API keys stored in operating system keychains are write-only from the frontend perspective and are never exposed in UI payloads or log files.

---

## Project Documentation

| Document | Description |
|---|---|
| [`docs/PRD.md`](docs/PRD.md) | Product Requirements Document and technical specifications |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records |
| [`docs/PROGRESS.md`](docs/PROGRESS.md) | Project progress and milestone execution history |
| [`docs/devices/infinix-note-11s.md`](docs/devices/infinix-note-11s.md) | Hardware baseline setup and testing notes |
| [`docs/rules-authoring.md`](docs/rules-authoring.md) | Detection rule authoring guidelines |
| [`docs/ui-components.md`](docs/ui-components.md) | Design system and component catalog |

---

## License

Proprietary. All rights reserved.
