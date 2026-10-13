import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { BridgeError } from './errors';
import { MockBridge } from './mock';
import type { Finding, ScanProgress, ScanResult } from './schema';

describe('MockBridge Lifecycle & Event Emission Suite', () => {
  let bridge: MockBridge;

  beforeEach(() => {
    vi.useFakeTimers();
    bridge = new MockBridge();
  });

  afterEach(() => {
    bridge.destroy();
    vi.useRealTimers();
  });

  describe('Preflight Service', () => {
    it('returns all 7 preflight checks', async () => {
      const checks = await bridge.preflight.run();
      expect(checks.length).toBe(7);
      expect(checks.map((c) => c.id)).toEqual([
        'adb_binary',
        'adb_server',
        'data_dir',
        'rules_pack',
        'work_dir',
        'keychain',
        'webview2',
      ]);
    });

    it('simulates ADB download with incremental updates using fake timers', async () => {
      const updates: string[] = [];
      bridge.on('preflight:update', (check) => {
        if (check.id === 'adb_binary') {
          updates.push(check.status);
        }
      });

      const downloadPromise = bridge.preflight.downloadAdb();
      expect(updates).toContain('downloading');

      // Advance fake timers by 100ms
      await vi.advanceTimersByTimeAsync(150);
      await downloadPromise;

      expect(updates).toContain('ok');
    });

    it('updates adb path and re-checks on retry', async () => {
      await bridge.preflight.selectAdbPath('D:\\custom\\adb.exe');
      const checks = await bridge.preflight.status();
      const binaryCheck = checks.find((c) => c.id === 'adb_binary');
      expect(binaryCheck?.detail).toContain('D:\\custom\\adb.exe');
    });
  });

  describe('Device Service', () => {
    it('returns connected mock device', async () => {
      const dev = await bridge.device.getConnected();
      expect(dev).not.toBeNull();
      expect(dev?.serial).toBe('MOCK-0001');
      expect(dev?.model).toBe('Infinix Note 11S (mock)');
    });

    it('throws BridgeError when selecting an unknown serial', async () => {
      await expect(bridge.device.select('UNKNOWN-SERIAL')).rejects.toThrow(
        BridgeError
      );
      try {
        await bridge.device.select('UNKNOWN-SERIAL');
      } catch (err) {
        expect(err).toBeInstanceOf(BridgeError);
        expect((err as BridgeError).code).toBe('DeviceOffline');
      }
    });
  });

  describe('Inventory Service', () => {
    it('filters packages by category and search term', async () => {
      const all = await bridge.inventory.list();
      expect(all.length).toBeGreaterThan(1);

      const userOnly = await bridge.inventory.list({ category: 'user', search: '' });
      expect(userOnly.every((a) => !a.isSystem)).toBe(true);

      const searchResult = await bridge.inventory.list({
        category: 'all',
        search: 'flashlight',
      });
      expect(searchResult.length).toBe(1);
      expect(searchResult[0].package).toBe('com.example.mock.adware');
    });

    it('extracts apk and returns simulated target path', async () => {
      const apkPath = await bridge.inventory.extractApk('com.example.mock.adware');
      expect(apkPath).toContain('com.example.mock.adware');
      expect(apkPath.endsWith('.apk')).toBe(true);
    });

    it('throws BridgeError for nonexistent package extraction', async () => {
      await expect(
        bridge.inventory.extractApk('com.nonexistent.app')
      ).rejects.toThrow(BridgeError);
    });
  });

  describe('Scan Service with Simulated Timers', () => {
    it('executes scan, emits progress, findings, and done in sequence', async () => {
      const progressSteps: number[] = [];
      const findings: Finding[] = [];
      let finalResult: ScanResult | null = null;

      bridge.on('scan:progress', (p: ScanProgress) => {
        progressSteps.push(p.progress);
      });
      bridge.on('scan:finding', (f: Finding) => {
        findings.push(f);
      });
      bridge.on('scan:done', (r: ScanResult) => {
        finalResult = r;
      });

      const scanId = await bridge.scan.start({ mode: 'quick', includeSystem: false });
      expect(scanId.startsWith('scan-mock-')).toBe(true);

      // Verify initial running snapshot
      const initialSnapshot = await bridge.scan.status(scanId);
      expect(initialSnapshot?.status).toBe('running');

      // Step timers through stages
      await vi.advanceTimersByTimeAsync(60);
      expect(progressSteps).toContain(25);

      await vi.advanceTimersByTimeAsync(60);
      expect(progressSteps).toContain(60);
      expect(findings.length).toBeGreaterThanOrEqual(1);

      await vi.advanceTimersByTimeAsync(60);
      expect(progressSteps).toContain(90);

      await vi.advanceTimersByTimeAsync(60);

      const doneResult = finalResult as ScanResult | null;
      expect(doneResult).not.toBeNull();
      expect(doneResult?.status).toBe('completed');
      expect(doneResult?.findings.length).toBeGreaterThanOrEqual(2);

      // Snapshot getter returns completed result
      const completedSnapshot = await bridge.scan.status(scanId);
      expect(completedSnapshot?.status).toBe('completed');
    });

    it('cancels an in-progress scan', async () => {
      const scanId = await bridge.scan.start({ mode: 'deep', includeSystem: true });
      await vi.advanceTimersByTimeAsync(60);

      await bridge.scan.cancel(scanId);
      const cancelledStatus = await bridge.scan.status(scanId);
      expect(cancelledStatus?.status).toBe('cancelled');
    });
  });

  describe('Monitor Service (Popup Hunter)', () => {
    it('starts monitoring and captures event and candidate timeline', async () => {
      let candidateEmitted = false;
      bridge.on('monitor:candidate', () => {
        candidateEmitted = true;
      });

      await bridge.monitor.start();
      expect(candidateEmitted).toBe(true);

      const snapshot = await bridge.monitor.snapshot();
      expect(snapshot.status.isActive).toBe(true);
      expect(snapshot.events.length).toBeGreaterThan(0);
      expect(snapshot.candidates.length).toBeGreaterThan(0);
      expect(snapshot.candidates[0].package).toBe('com.example.mock.popup');

      await bridge.monitor.stop();
      const stoppedSnapshot = await bridge.monitor.snapshot();
      expect(stoppedSnapshot.status.isActive).toBe(false);
    });
  });

  describe('Audit Service', () => {
    it('returns all system audit items including warning, issue, unavailable', async () => {
      const items = await bridge.audit.run();
      expect(items.length).toBeGreaterThanOrEqual(4);
      const statuses = items.map((i) => i.status);
      expect(statuses).toContain('ok');
      expect(statuses).toContain('warning');
      expect(statuses).toContain('issue');
      expect(statuses).toContain('unavailable');
    });
  });

  describe('Remediate Service', () => {
    it('creates a remediation plan from findings, executes it, and allows undo', async () => {
      const plan = await bridge.remediate.createPlan(['find-mock-01']);
      expect(plan.actions.length).toBe(1);
      expect(plan.actions[0].type).toBe('disable');

      const logs = await bridge.remediate.execute(plan);
      expect(logs.length).toBe(1);
      expect(logs[0].status).toBe('success');

      const history = await bridge.remediate.getHistory();
      expect(history.length).toBeGreaterThanOrEqual(1);

      const undone = await bridge.remediate.undo(logs[0].id);
      expect(undone.status).toBe('undone');
    });
  });

  describe('Report Service', () => {
    it('lists scan summaries and compares two scans', async () => {
      const summaries = await bridge.report.list();
      expect(summaries.length).toBeGreaterThan(0);

      // Create a second completed scan using public API
      const scanId2 = await bridge.scan.start({ mode: 'quick', includeSystem: false });
      await vi.advanceTimersByTimeAsync(250);

      const comparison = await bridge.report.compare('scan-mock-seed-01', scanId2);
      expect(comparison.scanA.id).toBe('scan-mock-seed-01');
      expect(comparison.scanB.id).toBe(scanId2);
      expect(comparison.scanA.findingsCount.high).toBe(1);

      const exportPath = await bridge.report.export({
        scanId: 'scan-mock-seed-01',
        format: 'json',
      });
      expect(exportPath).toBe('reports/scan_scan-mock-seed-01.json');
    });
  });

  describe('Settings & Allowlist Services', () => {
    it('updates and retrieves settings', async () => {
      const original = await bridge.settings.get();
      expect(original.theme).toBe('dark');

      const updated = await bridge.settings.update({ theme: 'light', customAdbPort: 5038 });
      expect(updated.theme).toBe('light');
      expect(updated.customAdbPort).toBe(5038);
    });

    it('adds and removes allowlist entries', async () => {
      const added = await bridge.allowlist.add({
        type: 'package',
        target: 'com.example.trusted',
        reason: 'User trusted utility',
      });
      expect(added.id.startsWith('allow-')).toBe(true);

      const listAfterAdd = await bridge.allowlist.list();
      expect(listAfterAdd.some((e) => e.target === 'com.example.trusted')).toBe(true);

      await bridge.allowlist.remove(added.id);
      const listAfterRemove = await bridge.allowlist.list();
      expect(listAfterRemove.some((e) => e.target === 'com.example.trusted')).toBe(false);
    });
  });

  describe('Event Unsubscription', () => {
    it('stops invoking callback after unsubscribe is called', () => {
      let callCount = 0;
      const unsubscribe = bridge.on('scan:progress', () => {
        callCount++;
      });

      bridge.emit('scan:progress', { step: 'step 1', progress: 10 });
      expect(callCount).toBe(1);

      unsubscribe();
      bridge.emit('scan:progress', { step: 'step 2', progress: 20 });
      expect(callCount).toBe(1);
    });
  });
});
