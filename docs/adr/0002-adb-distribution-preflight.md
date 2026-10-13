# ADR-0002: ADB Distribution Policy and Launch Preflight

## Status
Accepted

## Date
2026-10-11

## Context
Jerion relies on Android Debug Bridge (ADB) for all device queries, package inspections, dumpsys audits, and remediation actions. Bundling ADB binaries in the repository or installer poses licensing, maintenance, and platform bloat concerns. Furthermore, unexpected ADB version mismatches or broken environments cause application failure.

## Decisions

### 1. No Bundled ADB
- ADB will **not** be bundled inside the repository nor inside the application installer.
- The installer does not silently drop platform-tools onto the user's system.

### 2. Mandatory Preflight on Launch
Every application launch runs `internal/preflight` before displaying the primary dashboard. Preflight verifies essential requirements:
1. **P1 (ADB Binary)**: Resolved hierarchically:
   - Configured custom path in Jerion Settings.
   - System `adb` in PATH.
   - Application-managed directory (`%LOCALAPPDATA%\Jerion\tools\platform-tools`).
2. **P2 (ADB Server)**: Validates connectivity and version conflicts.
3. **P3 (Data Directory & Database)**: Validates SQLite storage write permissions.
4. **P4 (Embedded Rules Package)**: Validates integrity of embedded YAML rule files.
5. **P5 (Working Directory & Disk Space)**: Non-blocking warning if free disk space < 500 MB.
6. **P6 (OS Keychain)**: Non-blocking test of credentials storage.
7. **P7 (WebView2 Runtime)**: Validates presence of Evergreen WebView2 runtime on Windows (verified at installer level and checked by preflight).

### 3. Explicit User Consent for Downloads
- If ADB is missing (P1 failure), the user is presented with a Requirement Sheet detailing:
  - What is missing and why it is required.
  - Download size (queried dynamically).
  - Source: Official Google HTTPS repository.
  - Destination target path.
- Available actions: **Download**, **Choose existing...**, **Quit**.
- **No silent downloads**: Jerion will never download files without explicit user interaction and confirmation.
- Checksum verification must succeed before archive extraction.

## Consequences
- Clean legal compliance regarding third-party binary redistribution.
- Resilient startup sequence with transparent user feedback when requirements are missing.
