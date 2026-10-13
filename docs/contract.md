# Jerion Contract Specification & Governance

| | |
|---|---|
| **Contract Version** | `1.0.0` (FROZEN) |
| **Status** | **FROZEN** (Strictly Frozen at conclusion of Stage T3; ready for Stage T4 Backend Go) |
| **Freeze Target** | Achieved (`1.0.0` on 2026-10-13) |
| **Last Updated** | 2026-10-13 |

---

## 1. Overview & Single Source of Truth
This document defines the formal data contract and bridge interface connecting the Go backend and React frontend. 

1. **Frontend Source of Truth**: Defined via strict Zod schemas in `frontend/src/lib/bridge/schema.ts` with inferred TypeScript types (`z.infer`).
2. **Shared JSON Fixtures**: Stored in `internal/testdata/contract/*.json`. Vitest validates these fixtures strictly on every test run.
3. **Backend Source of Truth (Phase T4)**: Handled via Go DTO structs in `internal/contract/` tested against the same fixtures using `json.NewDecoder.DisallowUnknownFields()` and two-way round-trip marshaling.

---

## 2. Wire Protocol & Data Representation

### 2.1 Transport Error Format (Wails v2)
In Wails v2, Go backend errors are passed across the binding interface as standard string errors (`err.Error()`). To provide typed error handling without guesswork, errors use a structured wire format:

```text
JERION_ERR:{"code":"<BridgeErrorCode>","message":"<Description>","remedy":"<SuggestedAction>"}
```

#### Standard Error Codes (`BridgeErrorCode`):
- `AdbMissing`: ADB binary not found in configured path or system PATH.
- `DeviceUnauthorized`: Device connected via USB but RSA debugging prompt not authorized by user.
- `DeviceOffline`: Device detected in offline state or disconnected.
- `Timeout`: Command execution exceeded timeout deadline.
- `ParseFailure`: Defensive parser encountered unrecognized output or missing required fields.
- `PermissionDenied`: OEM or system restricted execution (e.g. security settings).
- `NotSupportedOnThisOS`: Feature requires OS/API version higher than target device or stubbed in current phase.
- `ActionBlocked`: Remediation action denied by safety guardrails or denylist.
- `InternalError`: Unhandled application error.

The frontend class `BridgeError` parses this wire format via `BridgeError.fromWire(raw)`.

### 2.2 `versionCode` Representation (int64 > 2^53)
Android API 28+ uses `versionCodeMajor << 32 | versionCode`, which can exceed JavaScript's `Number.MAX_SAFE_INTEGER` ($2^{53} - 1 = 9{,}007{,}199{,}254{,}740{,}991$).
- **Rule**: `versionCode` is represented across the contract strictly as a **decimal string** (e.g. `"2147483648"`, `"9223372036854775807"`).
- Never parse `versionCode` directly into a JavaScript `number`.

### 2.3 Timestamps & Dates
- **Rule**: All timestamps across the contract are represented as **ISO 8601 strings in UTC** (RFC 3339, e.g. `"2026-10-13T12:00:00Z"`).
- In Phase T4, Go DTOs will define time fields as `string` (formatted via `time.RFC3339`) rather than `time.Time`. This prevents Wails v2 binding generation from emitting `any` for `time.Time` fields in TypeScript.

---

## 3. Data Models Catalog

### 3.1 Base Enums
- **Severity**: `'info' | 'low' | 'medium' | 'high' | 'critical'`
- **Confidence**: `'low' | 'medium' | 'high'`
- **Verdict**: `'clean' | 'review' | 'high_risk'`
- **ConnectionState**: `'ready' | 'unauthorized' | 'offline' | 'no_device'`
- **PreflightCheckId**: `'adb_binary' | 'adb_server' | 'data_dir' | 'rules_pack' | 'work_dir' | 'keychain' | 'webview2'`
- **PreflightStatus**: `'checking' | 'ok' | 'warning' | 'missing' | 'downloading' | 'failed'`
- **ScanMode**: `'quick' | 'deep'`
- **ScanStatus**: `'running' | 'completed' | 'cancelled' | 'failed' | 'partial'`
- **EvidenceType**: `'permission' | 'manifest' | 'class_prefix' | 'hash' | 'setting' | 'log'`
- **RemediationActionType**: `'disable' | 'uninstall' | 'revoke_permission' | 'clear_appop'`
- **ActionLogStatus**: `'success' | 'failed' | 'undone'`
- **AuditItemStatus**: `'ok' | 'warning' | 'issue' | 'unavailable'`
- **MonitorEventType**: `'foreground_switch' | 'overlay_shown' | 'notification_posted'`
- **ReportFormat**: `'json' | 'html'`
- **Theme**: `'system' | 'dark' | 'light'`
- **AllowlistType**: `'package' | 'rule_package'`

### 3.2 Domain Models
| Model | Description | Strict Fields Key Attributes |
|---|---|---|
| `PreflightCheck` | Status of launch requirement P1–P7 | `id`, `title`, `status`, `detail`, `remedy?`, `sourceUrl?`, `downloadProgress?` |
| `Device` | Connected hardware profile | `id`, `serial`, `model`, `manufacturer`, `androidVersion`, `sdkInt`, `batteryLevel`, `isCharging`, `connectionState`, `isEmulator` |
| `SecuritySummary` | Device assessment summary | `counts: Record<Severity, number>`, `totalApps`, `lastScanAt?`, `verdict` |
| `AppInfo` | Inspected application metadata | `package`, `label`, `versionName`, `versionCode` (string), `isSystem`, `isSideloaded`, `installerPackage?`, `targetSdk`, `minSdk`, `installedAt`, `updatedAt`, `permissions`, `dangerousPermissions`, `hasOverlayPermission`, `hasAccessibilityPermission`, `isDeviceAdmin`, `isBatteryOptimized`, `signatures`, `suspiciousFlags` |
| `InventoryFilter` | App list filter parameters | `category` (`all/user/system/disabled/risky`), `search` |
| `ScanOptions` | Scan execution options | `mode` (`quick/deep`), `includeSystem` |
| `ScanProgress` | Incremental scan progress | `step`, `progress` (0–100), `currentPackage?` |
| `Evidence` | Raw evidence fragment | `type`, `title`, `details`, `rawOutput?` |
| `Finding` | Security detection finding | `id`, `ruleId`, `title`, `severity`, `confidence`, `package?` (optional for system findings), `whyFlagged`, `falsePositiveHint`, `evidence[]`, `recommendedAction?` |
| `ScanResult` | Completed or partial scan output | `id`, `deviceId`, `mode`, `startedAt`, `completedAt?`, `status`, `findings[]`, `summary` |
| `MonitorEvent` | Real-time surveillance event | `id`, `timestamp`, `type`, `package`, `details` |
| `MonitorCandidate` | Suspect ranked by heuristic score | `package`, `appLabel`, `confidence`, `score`, `timelineEvents[]`, `whyFlagged` |
| `MonitorSnapshot` | Live Popup Hunter snapshot | `status: MonitorStatus`, `events[]`, `candidates[]` |
| `AuditItem` | System configuration audit check | `id`, `category`, `title`, `status`, `summary`, `evidence[]`, `remedy?` |
| `RemediationAction` | Guarded remediation command | `id`, `type`, `package`, `title`, `impact`, `undoable`, `isBlocked`, `blockedReason?` |
| `RemediationPlan` | Ordered plan of actions | `actions[]`, `requiresBackup` |
| `ActionLogEntry` | Audit trail of executed action | `id`, `timestamp`, `action`, `status`, `error?`, `backupPath?` |
| `ScanSummaryItem` | Historical scan record | `id`, `startedAt`, `completedAt?`, `mode`, `verdict`, `findingsCount` |
| `ScanComparison` | Differential comparison of scans | `scanA`, `scanB`, `newFindings[]`, `resolvedFindings[]`, `unchangedFindings[]` |
| `ReportExportOptions` | Report export configuration | `scanId`, `format`, `targetPath?` |
| `Settings` | User configuration profile | `theme`, `adbPath?`, `customAdbPort?`, `deepScanTimeoutSec`, `autoCheckPreflight` |
| `AllowlistEntry` | Exemption entry | `id`, `type`, `target`, `addedAt`, `reason?` |

---

## 4. Bridge Services & Methods

The Bridge interface (`frontend/src/lib/bridge/bridge.ts`) groups domain operations into 10 services:

1. **`preflight`**: `run()`, `status()`, `downloadAdb()`, `selectAdbPath(path)`, `retry()`
2. **`device`**: `getConnected()`, `list()`, `select(serial)`
3. **`inventory`**: `list(filter?)`, `get(packageName)`, `extractApk(packageName, targetPath?)`
4. **`scan`**: `start(options)`, `status(scanId)`, `cancel(scanId)`, `get(scanId)`, `quickScan()`
5. **`monitor`**: `start()`, `stop()`, `snapshot()`
6. **`audit`**: `run()`, `getLatest()`
7. **`remediate`**: `createPlan(findingIds)`, `execute(plan)`, `getHistory()`, `undo(logEntryId)`
8. **`report`**: `list()`, `get(id)`, `compare(scanIdA, scanIdB)`, `export(options)`
9. **`settings`**: `get()`, `update(partial)`
10. **`allowlist`**: `list()`, `add(entry)`, `remove(id)`

### Snapshot Getters
To prevent race conditions and lost events for late-subscribing components, all real-time streams provide snapshot getters:
- `scan.status(scanId)`: retrieves the current state and intermediate findings of an ongoing or completed scan.
- `monitor.snapshot()`: retrieves current monitoring activity state, full event buffer, and candidate ranking.

### Event Map (`BridgeEventMap`):
- `'preflight:update'`: `PreflightCheck`
- `'device:changed'`: `Device | null`
- `'scan:progress'`: `ScanProgress`
- `'scan:finding'`: `Finding`
- `'scan:done'`: `ScanResult`
- `'monitor:event'`: `MonitorEvent`
- `'monitor:candidate'`: `MonitorCandidate`

---

## 5. Contract Governance & Freeze Timeline

### 5.1 Stages T1 to T3 (Draft Phase: `0.1.0` -> `0.x.x`)
- The contract is in **DRAFT** state.
- During Stage T2 (Design System & App Shell) and Stage T3 (Screen development), additive changes (new optional fields, additional service methods) are permitted.
- Additive changes require a brief proposal approved by the project owner, followed by:
  1. Minor version bump in this document (e.g. `0.2.0`).
  2. Synchronized updates to `schema.ts`.
  3. Updates to corresponding JSON fixtures in `internal/testdata/contract/`.
  4. Validation passing in Vitest (`npm run test`).

### 5.2 Strict Freeze at End of Stage T3 (`1.0.0`)
- At the conclusion of Stage T3, before Stage T4 (Backend Go) commences, the contract will be frozen as **`1.0.0`**.
- Once frozen, unilateral changes are strictly forbidden. Any future modification requires synchronized updates across TypeScript schemas, Go DTO structs, and shared JSON fixtures in a single atomic task.

### 5.3 Phase T4 Backend Go Two-Way Verification
In Phase T4, the backend Go test suite will verify fidelity against the shared fixtures:
1. **Strict Unmarshal**: `decoder.DisallowUnknownFields()` must successfully parse each fixture JSON into Go DTO structs.
2. **Round-Trip Marshal**: Marshaling the Go DTO struct back to JSON and re-verifying structural equivalence ensures zero omitted or renamed fields in either direction.

### 5.4 Production Dead-Code Elimination
- The mock bridge and all runtime Zod schemas are excluded from production builds.
- The npm script `npm run check:no-mock` scans all bundles in `frontend/dist/` for forbidden markers (`MOCK-0001`, `com.example.mock.`, `MockBridge`, `ZodObject`, `ZodError`) and is enforced in GitHub Actions CI.
