import { z } from 'zod';

export const CONTRACT_VERSION = '1.0.0';

// ==========================================
// Base Enumerations
// ==========================================

export const SeveritySchema = z.enum(['info', 'low', 'medium', 'high', 'critical']);
export type Severity = z.infer<typeof SeveritySchema>;

export const ConfidenceSchema = z.enum(['low', 'medium', 'high']);
export type Confidence = z.infer<typeof ConfidenceSchema>;

export const VerdictSchema = z.enum(['clean', 'review', 'high_risk']);
export type Verdict = z.infer<typeof VerdictSchema>;

export const ConnectionStateSchema = z.enum(['ready', 'unauthorized', 'offline', 'no_device']);
export type ConnectionState = z.infer<typeof ConnectionStateSchema>;

export const PreflightCheckIdSchema = z.enum([
  'adb_binary',
  'adb_server',
  'data_dir',
  'rules_pack',
  'work_dir',
  'keychain',
  'webview2',
]);
export type PreflightCheckId = z.infer<typeof PreflightCheckIdSchema>;

export const PreflightStatusSchema = z.enum([
  'checking',
  'ok',
  'warning',
  'missing',
  'downloading',
  'failed',
]);
export type PreflightStatus = z.infer<typeof PreflightStatusSchema>;

export const ScanModeSchema = z.enum(['quick', 'deep']);
export type ScanMode = z.infer<typeof ScanModeSchema>;

export const ScanStatusSchema = z.enum(['running', 'completed', 'cancelled', 'failed', 'partial']);
export type ScanStatus = z.infer<typeof ScanStatusSchema>;

export const EvidenceTypeSchema = z.enum([
  'permission',
  'manifest',
  'class_prefix',
  'hash',
  'setting',
  'log',
]);
export type EvidenceType = z.infer<typeof EvidenceTypeSchema>;

export const RemediationActionTypeSchema = z.enum([
  'disable',
  'uninstall',
  'revoke_permission',
  'clear_appop',
]);
export type RemediationActionType = z.infer<typeof RemediationActionTypeSchema>;

export const ActionLogStatusSchema = z.enum(['success', 'failed', 'undone']);
export type ActionLogStatus = z.infer<typeof ActionLogStatusSchema>;

export const AuditItemStatusSchema = z.enum(['ok', 'warning', 'issue', 'unavailable']);
export type AuditItemStatus = z.infer<typeof AuditItemStatusSchema>;

export const MonitorEventTypeSchema = z.enum([
  'foreground_switch',
  'overlay_shown',
  'notification_posted',
]);
export type MonitorEventType = z.infer<typeof MonitorEventTypeSchema>;

export const ReportFormatSchema = z.enum(['json', 'html']);
export type ReportFormat = z.infer<typeof ReportFormatSchema>;

export const ThemeSchema = z.enum(['system', 'dark', 'light']);
export type Theme = z.infer<typeof ThemeSchema>;

export const AllowlistTypeSchema = z.enum(['package', 'rule_package']);
export type AllowlistType = z.infer<typeof AllowlistTypeSchema>;

// ==========================================
// Domain Models (All Strictly Validated)
// ==========================================

export const PreflightCheckSchema = z
  .object({
    id: PreflightCheckIdSchema,
    title: z.string(),
    status: PreflightStatusSchema,
    detail: z.string(),
    remedy: z.string().optional(),
    sourceUrl: z.string().optional(),
    downloadProgress: z.number().min(0).max(100).optional(),
  })
  .strict();
export type PreflightCheck = z.infer<typeof PreflightCheckSchema>;

export const DeviceSchema = z
  .object({
    id: z.string(),
    serial: z.string(),
    model: z.string(),
    manufacturer: z.string(),
    androidVersion: z.string(),
    sdkInt: z.number(),
    batteryLevel: z.number().min(0).max(100),
    isCharging: z.boolean(),
    connectionState: ConnectionStateSchema,
    isEmulator: z.boolean(),
  })
  .strict();
export type Device = z.infer<typeof DeviceSchema>;

export const SecuritySummarySchema = z
  .object({
    counts: z.record(SeveritySchema, z.number()),
    totalApps: z.number(),
    lastScanAt: z.string().optional(),
    verdict: VerdictSchema,
  })
  .strict();
export type SecuritySummary = z.infer<typeof SecuritySummarySchema>;

export const AppInfoSchema = z
  .object({
    package: z.string(),
    label: z.string(),
    versionName: z.string(),
    versionCode: z.string(), // Decimal string per Point C (handles int64 > 2^53)
    isSystem: z.boolean(),
    isSideloaded: z.boolean(),
    installerPackage: z.string().optional(),
    targetSdk: z.number(),
    minSdk: z.number(),
    installedAt: z.string(),
    updatedAt: z.string(),
    permissions: z.array(z.string()),
    dangerousPermissions: z.array(z.string()),
    hasOverlayPermission: z.boolean(),
    hasAccessibilityPermission: z.boolean(),
    isDeviceAdmin: z.boolean(),
    isBatteryOptimized: z.boolean(),
    signatures: z.array(z.string()),
    suspiciousFlags: z.array(z.string()),
  })
  .strict();
export type AppInfo = z.infer<typeof AppInfoSchema>;

export const InventoryFilterSchema = z
  .object({
    category: z.enum(['all', 'user', 'system', 'disabled', 'risky']),
    search: z.string(),
  })
  .strict();
export type InventoryFilter = z.infer<typeof InventoryFilterSchema>;

export const ScanOptionsSchema = z
  .object({
    mode: ScanModeSchema,
    includeSystem: z.boolean(),
  })
  .strict();
export type ScanOptions = z.infer<typeof ScanOptionsSchema>;

export const ScanProgressSchema = z
  .object({
    step: z.string(),
    progress: z.number().min(0).max(100),
    currentPackage: z.string().optional(),
  })
  .strict();
export type ScanProgress = z.infer<typeof ScanProgressSchema>;

export const EvidenceSchema = z
  .object({
    type: EvidenceTypeSchema,
    title: z.string(),
    details: z.string(),
    rawOutput: z.string().optional(),
  })
  .strict();
export type Evidence = z.infer<typeof EvidenceSchema>;

export const RemediationActionSchema = z
  .object({
    id: z.string(),
    type: RemediationActionTypeSchema,
    package: z.string(),
    title: z.string(),
    impact: z.string(),
    undoable: z.boolean(),
    isBlocked: z.boolean(),
    blockedReason: z.string().optional(),
  })
  .strict();
export type RemediationAction = z.infer<typeof RemediationActionSchema>;

export const FindingSchema = z
  .object({
    id: z.string(),
    ruleId: z.string(),
    title: z.string(),
    severity: SeveritySchema,
    confidence: ConfidenceSchema,
    package: z.string().optional(), // Optional per Point E (system-level findings)
    whyFlagged: z.string(),
    falsePositiveHint: z.string(),
    evidence: z.array(EvidenceSchema),
    recommendedAction: RemediationActionSchema.optional(),
  })
  .strict();
export type Finding = z.infer<typeof FindingSchema>;

export const ScanResultSchema = z
  .object({
    id: z.string(),
    deviceId: z.string(),
    mode: ScanModeSchema,
    startedAt: z.string(),
    completedAt: z.string().optional(),
    status: ScanStatusSchema,
    findings: z.array(FindingSchema),
    summary: SecuritySummarySchema,
  })
  .strict();
export type ScanResult = z.infer<typeof ScanResultSchema>;

export const MonitorEventSchema = z
  .object({
    id: z.string(),
    timestamp: z.string(),
    type: MonitorEventTypeSchema,
    package: z.string(),
    details: z.record(z.string(), z.unknown()),
  })
  .strict();
export type MonitorEvent = z.infer<typeof MonitorEventSchema>;

export const MonitorCandidateSchema = z
  .object({
    package: z.string(),
    appLabel: z.string(),
    confidence: ConfidenceSchema,
    score: z.number(),
    timelineEvents: z.array(MonitorEventSchema),
    whyFlagged: z.string(),
  })
  .strict();
export type MonitorCandidate = z.infer<typeof MonitorCandidateSchema>;

export const MonitorStatusSchema = z
  .object({
    isActive: z.boolean(),
    startedAt: z.string().optional(),
    sourcesActive: z.array(z.string()),
  })
  .strict();
export type MonitorStatus = z.infer<typeof MonitorStatusSchema>;

export const MonitorSnapshotSchema = z
  .object({
    status: MonitorStatusSchema,
    events: z.array(MonitorEventSchema),
    candidates: z.array(MonitorCandidateSchema),
  })
  .strict();
export type MonitorSnapshot = z.infer<typeof MonitorSnapshotSchema>;

export const AuditItemSchema = z
  .object({
    id: z.string(),
    category: z.string(),
    title: z.string(),
    status: AuditItemStatusSchema,
    summary: z.string(),
    evidence: z.array(z.string()),
    remedy: z.string().optional(),
  })
  .strict();
export type AuditItem = z.infer<typeof AuditItemSchema>;

export const RemediationPlanSchema = z
  .object({
    actions: z.array(RemediationActionSchema),
    requiresBackup: z.boolean(),
  })
  .strict();
export type RemediationPlan = z.infer<typeof RemediationPlanSchema>;

export const ActionLogEntrySchema = z
  .object({
    id: z.string(),
    timestamp: z.string(),
    action: RemediationActionSchema,
    status: ActionLogStatusSchema,
    error: z.string().optional(),
    backupPath: z.string().optional(),
    auditSaved: z.boolean().optional(),
    auditError: z.string().optional(),
  })
  .strict();
export type ActionLogEntry = z.infer<typeof ActionLogEntrySchema>;

export const ScanSummaryItemSchema = z
  .object({
    id: z.string(),
    startedAt: z.string(),
    completedAt: z.string().optional(),
    mode: ScanModeSchema,
    verdict: VerdictSchema,
    findingsCount: z.record(SeveritySchema, z.number()),
  })
  .strict();
export type ScanSummaryItem = z.infer<typeof ScanSummaryItemSchema>;

export const ScanComparisonSchema = z
  .object({
    scanA: ScanSummaryItemSchema,
    scanB: ScanSummaryItemSchema,
    newFindings: z.array(FindingSchema),
    resolvedFindings: z.array(FindingSchema),
    unchangedFindings: z.array(FindingSchema),
  })
  .strict();
export type ScanComparison = z.infer<typeof ScanComparisonSchema>;

export const ReportExportOptionsSchema = z
  .object({
    scanId: z.string(),
    format: ReportFormatSchema,
    targetPath: z.string().optional(),
  })
  .strict();
export type ReportExportOptions = z.infer<typeof ReportExportOptionsSchema>;

export const SettingsSchema = z
  .object({
    theme: ThemeSchema,
    adbPath: z.string().optional(),
    customAdbPort: z.number().optional(),
    deepScanTimeoutSec: z.number(),
    autoCheckPreflight: z.boolean(),
  })
  .strict();
export type Settings = z.infer<typeof SettingsSchema>;

export const AllowlistEntrySchema = z
  .object({
    id: z.string(),
    type: AllowlistTypeSchema,
    target: z.string(),
    addedAt: z.string(),
    reason: z.string().optional(),
  })
  .strict();
export type AllowlistEntry = z.infer<typeof AllowlistEntrySchema>;
