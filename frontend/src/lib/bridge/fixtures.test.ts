import { describe, expect, it } from 'vitest';
import actionLogEntryFixture from '../../../../internal/testdata/contract/action_log_entry.json';
import allowlistEntryFixture from '../../../../internal/testdata/contract/allowlist_entry.json';
import appInfoFixture from '../../../../internal/testdata/contract/app_info.json';
import auditItemFixture from '../../../../internal/testdata/contract/audit_item.json';
import deviceFixture from '../../../../internal/testdata/contract/device.json';
import evidenceFixture from '../../../../internal/testdata/contract/evidence.json';
import findingFixture from '../../../../internal/testdata/contract/finding.json';
import monitorCandidateFixture from '../../../../internal/testdata/contract/monitor_candidate.json';
import monitorEventFixture from '../../../../internal/testdata/contract/monitor_event.json';
import monitorSnapshotFixture from '../../../../internal/testdata/contract/monitor_snapshot.json';
import preflightCheckFixture from '../../../../internal/testdata/contract/preflight_check.json';
import remediationActionFixture from '../../../../internal/testdata/contract/remediation_action.json';
import remediationPlanFixture from '../../../../internal/testdata/contract/remediation_plan.json';
import reportExportOptionsFixture from '../../../../internal/testdata/contract/report_export_options.json';
import scanComparisonFixture from '../../../../internal/testdata/contract/scan_comparison.json';
import scanOptionsFixture from '../../../../internal/testdata/contract/scan_options.json';
import scanProgressFixture from '../../../../internal/testdata/contract/scan_progress.json';
import scanResultFixture from '../../../../internal/testdata/contract/scan_result.json';
import scanSummaryItemFixture from '../../../../internal/testdata/contract/scan_summary_item.json';
import securitySummaryFixture from '../../../../internal/testdata/contract/security_summary.json';
import settingsFixture from '../../../../internal/testdata/contract/settings.json';
import {
  ActionLogEntrySchema,
  AllowlistEntrySchema,
  AppInfoSchema,
  AuditItemSchema,
  DeviceSchema,
  EvidenceSchema,
  FindingSchema,
  MonitorCandidateSchema,
  MonitorEventSchema,
  MonitorSnapshotSchema,
  PreflightCheckSchema,
  RemediationActionSchema,
  RemediationPlanSchema,
  ReportExportOptionsSchema,
  ScanComparisonSchema,
  ScanOptionsSchema,
  ScanProgressSchema,
  ScanResultSchema,
  ScanSummaryItemSchema,
  SecuritySummarySchema,
  SettingsSchema,
} from './schema';

describe('Shared Contract JSON Fixtures (Strict Schema Validation)', () => {
  it('validates preflight_check.json strictly', () => {
    expect(() => PreflightCheckSchema.parse(preflightCheckFixture)).not.toThrow();
  });

  it('validates device.json strictly', () => {
    const parsed = DeviceSchema.parse(deviceFixture);
    expect(parsed.serial).toBe('MOCK-0001');
    expect(parsed.model).toBe('Infinix Note 11S (mock)');
  });

  it('validates security_summary.json strictly', () => {
    expect(() => SecuritySummarySchema.parse(securitySummaryFixture)).not.toThrow();
  });

  it('validates app_info.json strictly and confirms decimal string versionCode', () => {
    const parsed = AppInfoSchema.parse(appInfoFixture);
    expect(parsed.package).toBe('com.example.mock.adware');
    expect(typeof parsed.versionCode).toBe('string');
  });

  it('validates scan_options.json strictly', () => {
    expect(() => ScanOptionsSchema.parse(scanOptionsFixture)).not.toThrow();
  });

  it('validates scan_progress.json strictly', () => {
    expect(() => ScanProgressSchema.parse(scanProgressFixture)).not.toThrow();
  });

  it('validates evidence.json strictly', () => {
    expect(() => EvidenceSchema.parse(evidenceFixture)).not.toThrow();
  });

  it('validates finding.json strictly', () => {
    const parsed = FindingSchema.parse(findingFixture);
    expect(parsed.ruleId).toBe('ADW-003');
    expect(parsed.package).toBe('com.example.mock.adware');
  });

  it('validates scan_result.json strictly', () => {
    const parsed = ScanResultSchema.parse(scanResultFixture);
    expect(parsed.findings.length).toBeGreaterThan(0);
    expect(parsed.summary.verdict).toBe('review');
  });

  it('validates monitor_event.json strictly', () => {
    const parsed = MonitorEventSchema.parse(monitorEventFixture);
    expect(parsed.type).toBe('overlay_shown');
  });

  it('validates monitor_candidate.json strictly', () => {
    const parsed = MonitorCandidateSchema.parse(monitorCandidateFixture);
    expect(parsed.package).toBe('com.example.mock.popup');
    expect(parsed.timelineEvents.length).toBeGreaterThan(0);
  });

  it('validates monitor_snapshot.json strictly', () => {
    const parsed = MonitorSnapshotSchema.parse(monitorSnapshotFixture);
    expect(parsed.status.isActive).toBe(true);
    expect(parsed.candidates.length).toBeGreaterThan(0);
  });

  it('validates audit_item.json strictly', () => {
    const parsed = AuditItemSchema.parse(auditItemFixture);
    expect(parsed.status).toBe('warning');
  });

  it('validates remediation_action.json strictly', () => {
    const parsed = RemediationActionSchema.parse(remediationActionFixture);
    expect(parsed.type).toBe('disable');
    expect(parsed.undoable).toBe(true);
  });

  it('validates remediation_plan.json strictly', () => {
    const parsed = RemediationPlanSchema.parse(remediationPlanFixture);
    expect(parsed.actions.length).toBe(1);
    expect(parsed.requiresBackup).toBe(true);
  });

  it('validates action_log_entry.json strictly', () => {
    const parsed = ActionLogEntrySchema.parse(actionLogEntryFixture);
    expect(parsed.status).toBe('success');
  });

  it('validates scan_summary_item.json strictly', () => {
    expect(() => ScanSummaryItemSchema.parse(scanSummaryItemFixture)).not.toThrow();
  });

  it('validates scan_comparison.json strictly', () => {
    const parsed = ScanComparisonSchema.parse(scanComparisonFixture);
    expect(parsed.resolvedFindings.length).toBe(1);
  });

  it('validates report_export_options.json strictly', () => {
    expect(() => ReportExportOptionsSchema.parse(reportExportOptionsFixture)).not.toThrow();
  });

  it('validates settings.json strictly', () => {
    const parsed = SettingsSchema.parse(settingsFixture);
    expect(parsed.theme).toBe('dark');
  });

  it('validates allowlist_entry.json strictly', () => {
    const parsed = AllowlistEntrySchema.parse(allowlistEntryFixture);
    expect(parsed.target).toBe('com.android.settings');
  });
});
