# Jerion

Jerion is a desktop forensic and auditing utility for Android devices. Operating directly over Android Debug Bridge, it identifies adware, intrusive background popups, stalkerware signatures, and suspicious application behaviors without requiring device rooting or bootloader modification.

---

## Overview

Most mobile security software relies on opaque protection percentages and decorative scanning animations that conceal actual device state. Jerion approaches mobile security as a verifiable diagnostic process. By querying official Android developer interfaces, the application extracts runtime evidence from package managers, window states, permissions, and raw application binaries.

### Design Principles

* **Evidence over Arbitrary Scores**: Jerion does not compute artificial percentage ratings. Findings present concrete facts such as abusive window overlay types, hidden launcher entries, unrequested accessibility service registrations, and known tracking signatures located in compiled DEX bytecode.
* **Controlled and Reversible Actions**: Modifications to the connected device require explicit user approval. Critical operating system packages are shielded by hardcoded protection lists, and package uninstalls automatically trigger base APK backups to local disk before execution.
* **Local-First Privacy**: Jerion functions entirely offline by default. The application contains zero telemetry, requires no remote web assets or content delivery networks, and never extracts personal user data such as communications, media, or contact lists.

---

## Core Capabilities

### Launch Environment Preflight
Every application startup runs an automated seven-point diagnostic check before loading the primary workspace. This routine verifies the presence of the ADB executable, initializes or connects to the local ADB daemon, mounts the local SQLite database, loads the embedded forensic rule catalog, and tests operating system credential store availability. If a prerequisite is missing, the interface presents verified setup instructions and direct links to official platform tools.

### Forensic Scan Engine
Jerion provides fast and deep forensic scanning pipelines governed by declarative rule sets:
* **Manifest Inspection**: A pure Go AXML decoder parses compressed Android manifests directly from APK files without requiring a Java runtime environment.
* **Bytecode Analysis**: Deep scans inspect DEX header structures and type identifiers to detect embedded ad fraud libraries, aggressive telemetry frameworks, and tracking software.
* **Certificate Attribution**: The engine reads APK signing blocks to extract SHA-256 publisher fingerprints for identity validation.

### Real-Time Window Telemetry and Popup Hunter
Unsolicited popups and overlay advertisements frequently appear out of context and vanish before users can determine their origin. Jerion resolves this by monitoring window hierarchy changes and application focus transitions via continuous system dumps. When an unwanted advertisement appears on the physical handset, pressing the capture button records the active window layer and calculates time offsets to identify the responsible background package.

### High-Density Package Inventory
Standard Android installations regularly contain hundreds of user and system applications. Jerion renders this data using a virtualized table that maintains steady performance across large package lists. Operators can filter by partition, search by package name or label, inspect granted versus requested permissions, verify installation source identity, and export application packages directly to local storage.

### System Hardening Audit
Device vulnerability frequently stems from permissive global settings rather than rogue packages alone. Jerion audits critical security toggles across the device, evaluating active device administrators, accessibility holders, notification listeners, USB debugging configuration, third-party installation flags, global network proxy settings, and private DNS status.

### Guardrailed Remediation
Remediation actions include clear protective safeguards:
* **Core System Denylist**: Crucial Android framework components and vendor baseline packages are permanently locked against accidental removal or disabling.
* **Pre-Execution Backup**: Removing a package automatically pulls the corresponding APK to local storage first.
* **Action Journal**: Every package modification records a corresponding rollback command in the local database for rapid restoration.

### Self-Contained Offline Reports
Audit findings can be exported into standalone HTML and JSON formats. Exported HTML files embed all styling, typography, and vector assets directly, allowing investigators to view reports completely offline. Reports support automated masking of hardware identifiers such as serial numbers, IMEI values, and MAC addresses.

---

## Technical Architecture

| Component | Technology | Rationale |
|---|---|---|
| Desktop Framework | Wails v2.16 | Direct Go to TypeScript IPC without heavy browser runtime overhead |
| Backend Runtime | Go 1.25+ | High throughput concurrency and standalone binary compilation |
| Local Database | Pure Go SQLite (`modernc.org/sqlite`) | Embedded relational storage without CGO compiler requirements |
| Secret Management | OS Credential Store (`zalando/go-keyring`) | Secure credential isolation preventing plaintext token persistence |
| Frontend Interface | React 19 and TypeScript | Strict typing aligned with backend contract schemas |
| Layout and Design | Tailwind CSS | Consistent four-pixel layout grid with dedicated token variables |
| Table Virtualization | TanStack Virtual | Smooth navigation through large inventories |
| 3D Visualization | Three.js with React Three Fiber | Real-time procedural rendering paused during background states |

---

## Getting Started

### Prerequisites
Before running or compiling Jerion, ensure your development environment includes:
1. Go 1.25 or later
2. Node.js 22 LTS with npm
3. Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`)
4. Android SDK Platform-Tools with `adb` available on your system path
5. A target Android handset with USB Debugging enabled under Developer Options

### Installation and Setup
Clone the repository and install frontend dependencies:

```bash
git clone https://github.com/EnoughForYou/Jerion.git
cd Jerion/frontend
npm ci
cd ..
```

### Running in Development
Start the application with live code reloading and Go method synchronization:

```bash
wails dev
```

The desktop application window will open automatically. The user interface is also accessible in any Chromium browser at `http://localhost:34115` for developer inspection.

### Production Build
To generate a standalone production executable for Windows x64:

```bash
wails build
```

The compiled binary is written to `build/bin/Jerion.exe`.

---

## Typical Usage Workflow

1. **Connect Target Device**: Connect the Android handset via USB. Verify the host computer authorization prompt on the device screen.
2. **Execute Diagnostic Preflight**: Launch Jerion. The preflight screen verifies ADB daemon connectivity and rule catalogs automatically.
3. **Run an Audit**: Choose Quick Scan for manifest and package permission analysis, or Deep Scan for DEX signature and bytecode inspection.
4. **Investigate Popups**: If the device displays unwanted ads during testing, trigger the Popup Hunter to match window focus timestamps with active background processes.
5. **Review and Remediate**: Review flagged packages, inspect supporting technical evidence, and disable or uninstall confirmed threats with automatic APK backups.
6. **Export Report**: Generate a self-contained HTML audit summary for external archiving or client delivery.

---

## Computational Reproducibility and Quality Gates

Following reproducible research and engineering best practices, every build must pass complete static analysis and automated verification suites prior to release:

```bash
# 1. Run backend verification across all 17 internal Go packages
go vet ./... && go test -count=1 ./...

# 2. Run frontend linting, strict type checking, and unit tests
cd frontend
npm run lint
npm run typecheck
npm run test

# 3. Verify zero development mocks or test markers exist in production assets
npm run check:no-mock
```

---

## Hardware Baseline

Parser reliability and ADB command compatibility are validated against real physical hardware rather than generic emulators alone:

| Property | Target Specification |
|---|---|
| Primary Hardware | Infinix Note 11S, Model X698 |
| Operating System | XOS 10.6 on Android 12, API Level 31 |
| Vendor Safeguards | Protected package lists for Transsion launcher, system update, and maintenance suites |
| Fallback Standard | Graceful "Not Available" state reporting when vendor firmware omits standard parameters |

---

## Security and Privacy Boundaries

Jerion is designed for defensive security assessments on devices you own or have explicit authorization to inspect:
* **No Unauthorized Data Access**: The tool does not query or extract call logs, text messages, stored photographs, contacts, or location records.
* **Non-Root Operation**: All diagnostic routines operate within standard Android Debug Bridge boundaries.
* **Hash-Based Intelligence**: When external threat intelligence lookups are configured, queries send only cryptographic SHA-256 digests. Raw APK uploads require manual confirmation on a per-file basis.
* **Memory Isolation**: External API keys stored in the system credential manager are write-only from the frontend perspective and are never transmitted back to the user interface.

---

## Project Documentation

| Document | Topic |
|---|---|
| [`docs/PRD.md`](docs/PRD.md) | Comprehensive Product Requirements Document and specifications |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records catalog |
| [`docs/PROGRESS.md`](docs/PROGRESS.md) | Milestone tracking, verification logs, and active status |
| [`docs/contract.md`](docs/contract.md) | Single-source data transfer contracts between Go and TypeScript |
| [`docs/devices/infinix-note-11s.md`](docs/devices/infinix-note-11s.md) | Reference hardware testing notes and vendor quirks |
| [`docs/rules-authoring.md`](docs/rules-authoring.md) | Forensic detection rule schemas and authoring standards |
| [`docs/ui-components.md`](docs/ui-components.md) | Design token specifications and component catalog |

---

## License

Copyright 2026 EnoughForYou. All rights reserved.

This source code and related documentation are proprietary and confidential. Unauthorized copying, distribution, or modification of this project is strictly prohibited.
