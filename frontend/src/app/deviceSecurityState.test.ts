import { describe, it, expect } from 'vitest';
import { getCanonicalSecurityState, RULE_VERSION } from './deviceSecurityState';
import type { Device, ScanResult } from '../lib/bridge/schema';

describe('Canonical Device Security State Machine', () => {
  const readyDevice: Device = {
    id: 'DEVICE_SERIAL_123',
    serial: 'DEVICE_SERIAL_123',
    model: 'Infinix Note 11S',
    manufacturer: 'Infinix',
    androidVersion: '12',
    sdkInt: 31,
    connectionState: 'ready',
    batteryLevel: 85,
    isCharging: false,
    isEmulator: false,
  };

  const cleanScanResult: ScanResult = {
    id: 'scan-1',
    deviceId: 'DEVICE_SERIAL_123',
    mode: 'quick',
    startedAt: '2026-10-16T10:00:00Z',
    completedAt: '2026-10-16T10:01:00Z',
    status: 'completed',
    findings: [],
    summary: {
      counts: { critical: 0, high: 0, medium: 0, low: 0, info: 0 },
      totalApps: 42,
      verdict: 'clean',
    },
  };

  it('1. returns "no-device" when device is null', () => {
    const res = getCanonicalSecurityState(null, cleanScanResult);
    expect(res.state).toBe('no-device');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(false);
    expect(res.disabledReason).toContain('Connect an Android device');
    expect(res.contextText).toBe('No device connected');
  });

  it('2. returns "unauthorized" when device connectionState is unauthorized', () => {
    const dev: Device = { ...readyDevice, connectionState: 'unauthorized' };
    const res = getCanonicalSecurityState(dev, cleanScanResult);
    expect(res.state).toBe('unauthorized');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(false);
    expect(res.disabledReason).toContain('Authorize USB debugging');
    expect(res.contextText).toContain('Authorization required');
  });

  it('3. returns "offline" when device connectionState is offline', () => {
    const dev: Device = { ...readyDevice, connectionState: 'offline' };
    const res = getCanonicalSecurityState(dev, cleanScanResult);
    expect(res.state).toBe('offline');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(false);
    expect(res.disabledReason).toContain('offline');
    expect(res.contextText).toContain('Device offline');
  });

  it('4. returns "scanning" when isScanning is true regardless of existing scan', () => {
    const res = getCanonicalSecurityState(readyDevice, cleanScanResult, true);
    expect(res.state).toBe('scanning');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(false);
    expect(res.contextText).toContain('Scanning in progress');
  });

  it('5. returns "never-scanned" when device is ready but scanResult is null', () => {
    const res = getCanonicalSecurityState(readyDevice, null, false);
    expect(res.state).toBe('never-scanned');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(true);
    expect(res.contextText).toContain('No scan performed yet');
  });

  it('6. prevents cross-device leakage: ignores scanResult belonging to a different device serial', () => {
    const otherDeviceScan: ScanResult = {
      ...cleanScanResult,
      deviceId: 'DIFFERENT_SERIAL_999',
    };
    const res = getCanonicalSecurityState(readyDevice, otherDeviceScan, false);
    expect(res.state).toBe('never-scanned');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(true);
    expect(res.scanResult).toBeNull();
  });

  it('7. prevents empty deviceId leakage: rejects scanResult with empty deviceId string', () => {
    const emptyIdScan: ScanResult = {
      ...cleanScanResult,
      deviceId: '',
    };
    const res = getCanonicalSecurityState(readyDevice, emptyIdScan, false);
    expect(res.state).toBe('never-scanned');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.scanResult).toBeNull();
  });

  it('8. returns "partial-failed" when scan ended in failed, cancelled, or partial status', () => {
    const failedScan: ScanResult = {
      ...cleanScanResult,
      status: 'failed',
    };
    const resFailed = getCanonicalSecurityState(readyDevice, failedScan, false);
    expect(resFailed.state).toBe('partial-failed');
    expect(resFailed.allowsCleanVerdict).toBe(false);
    expect(resFailed.canStartScan).toBe(true);
    expect(resFailed.contextText).toContain('Scan failed');

    const partialScan: ScanResult = {
      ...cleanScanResult,
      status: 'partial',
    };
    const resPartial = getCanonicalSecurityState(readyDevice, partialScan, false);
    expect(resPartial.state).toBe('partial-failed');
    expect(resPartial.allowsCleanVerdict).toBe(false);
    expect(resPartial.contextText).toContain('Partial scan');
  });

  it('9. returns "scanned-clean" ONLY when scan is completed, serial matches, and findingsCount === 0', () => {
    const res = getCanonicalSecurityState(readyDevice, cleanScanResult, false);
    expect(res.state).toBe('scanned-clean');
    expect(res.allowsCleanVerdict).toBe(true);
    expect(res.canStartScan).toBe(true);
    expect(res.findingsCount).toBe(0);
    expect(res.contextText).toContain('No findings in the last scan');
    expect(res.contextText).toContain('Quick scan');
    expect(res.contextText).toContain('42 apps');
    expect(res.contextText).toContain(RULE_VERSION);
  });

  it('10. returns "scanned-findings" when scan is completed and findingsCount > 0', () => {
    const findingsScan: ScanResult = {
      ...cleanScanResult,
      findings: [
        {
          id: 'F-1',
          ruleId: 'POP-001',
          severity: 'critical',
          confidence: 'high',
          package: 'com.shady.ad',
          title: 'Aggressive Adware Detected',
          whyFlagged: 'Displays popups outside app',
          falsePositiveHint: 'None',
          evidence: [],
        },
      ],
      summary: {
        counts: { critical: 1, high: 0, medium: 0, low: 0, info: 0 },
        totalApps: 42,
        verdict: 'review',
      },
    };
    const res = getCanonicalSecurityState(readyDevice, findingsScan, false);
    expect(res.state).toBe('scanned-findings');
    expect(res.allowsCleanVerdict).toBe(false);
    expect(res.canStartScan).toBe(true);
    expect(res.findingsCount).toBe(1);
    expect(res.contextText).toContain('1 findings');
    expect(res.contextText).toContain('Quick scan');
    expect(res.contextText).toContain(RULE_VERSION);
  });
});
