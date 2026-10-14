# ADR-0005: ClamAV Daemon (clamd TCP) Integration for Local Threat Scanning

## Status
Accepted

## Date
2026-10-14

## Context
Mobile threat analysis in Jerion originally supported local heuristics/YAML rules and optional VirusTotal hash lookups. However, VirusTotal requires an API key, imposes cloud rate limits on the public free tier, and evaluates hashes remotely.

For privacy-sensitive security practitioners, an on-premise, zero-cloud antivirus engine is required to inspect extracted Android APK payloads locally. ClamAV is the industry-standard open-source antivirus engine. Unlike cloud hash aggregators, ClamAV inspects actual binary content and byte streams.

## Decisions

### 1. Direct Socket Communication via Native ClamAV Daemon (`clamd` TCP)
- **Protocol**: Pure-Go implementation of the native ClamAV daemon protocol (`zPING`, `zVERSION`, and `zINSTREAM`) using standard library `net`, `io`, and `encoding/binary`.
- **Zero Third-Party Dependencies**: No external CGO or unmaintained wrapper libraries.
- **Port**: Default endpoint `127.0.0.1:3310` configurable in Jerion Settings.

### 2. Streaming Inspection Protocol (`zINSTREAM`)
- APK payloads are streamed as bounded chunks (`[uint32 length][chunk bytes]`) directly to the `clamd` TCP socket, terminated by a zero-length chunk.
- No temporary files are written to external disk locations during scanning.

### 3. Opt-in Architecture
- ClamAV integration is strictly opt-in (`clamav_enabled` flag stored in SQLite).
- If ClamAV is disabled or unreachable, Jerion continues executing all declarative YAML rule evaluations and local heuristic scans without interruption.

### 4. User Interface & Connection Diagnostics
- Dedicated configuration card in `SettingsView.tsx` with live "Test Connection" button querying `zVERSION` on the target `host:port`.
- Real-time visual connection status and version reporting.

## Consequences
- Enables completely offline, high-throughput APK signature scanning without third-party API quotas.
- Zero risk to the host PC: Android APK payloads are streamed as passive binary data to the scanner and are never executed on Windows.
- Clean isolation between heuristic rule evaluation and antivirus signature matching.
