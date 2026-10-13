import type {
  AllowlistService,
  AuditService,
  Bridge,
  DeviceService,
  InventoryService,
  MonitorService,
  PreflightService,
  RemediateService,
  ReportService,
  ScanService,
  SettingsService,
} from './bridge';
import { BridgeError } from './errors';
import type { BridgeEventCallback, BridgeEventKey, BridgeEventMap, Unsubscribe } from './events';
import type {
  ActionLogEntry,
  AllowlistEntry,
  AppInfo,
  AuditItem,
  Device,
  Finding,
  InventoryFilter,
  MonitorCandidate,
  MonitorEvent,
  MonitorSnapshot,
  MonitorStatus,
  PreflightCheck,
  RemediationAction,
  RemediationPlan,
  ReportExportOptions,
  ScanComparison,
  ScanOptions,
  ScanProgress,
  ScanResult,
  ScanSummaryItem,
  SecuritySummary,
  Settings,
} from './schema';

// ==========================================
// Mock Data Baseline (Sanitized per Point G)
// ==========================================

export const MOCK_DEVICE: Device = {
  id: 'dev-mock-01',
  serial: 'MOCK-0001',
  model: 'Infinix Note 11S (mock)',
  manufacturer: 'Infinix',
  androidVersion: '12',
  sdkInt: 31,
  batteryLevel: 88,
  isCharging: false,
  connectionState: 'ready',
  isEmulator: false,
};

export const MOCK_PREFLIGHT_CHECKS: PreflightCheck[] = [
  {
    id: 'adb_binary',
    title: 'ADB Binary (P1)',
    status: 'ok',
    detail: 'Android Debug Bridge version 1.0.41',
  },
  {
    id: 'adb_server',
    title: 'ADB Server Daemon (P2)',
    status: 'ok',
    detail: 'Running on localhost:5037 without conflicts',
  },
  {
    id: 'data_dir',
    title: 'Local Database & Storage (P3)',
    status: 'ok',
    detail: 'SQLite read/write access verified in AppData',
  },
  {
    id: 'rules_pack',
    title: 'Threat Detection Rules (P4)',
    status: 'ok',
    detail: 'Embedded rules package loaded (32 active rules)',
  },
  {
    id: 'work_dir',
    title: 'Disk Space & Temp Workspace (P5)',
    status: 'ok',
    detail: 'Workspace ready; free space > 10 GB',
  },
  {
    id: 'keychain',
    title: 'OS Credentials Keychain (P6)',
    status: 'ok',
    detail: 'Windows Credential Manager accessible',
  },
  {
    id: 'webview2',
    title: 'WebView2 Evergreen Runtime (P7)',
    status: 'ok',
    detail: 'Evergreen WebView2 runtime verified',
  },
];

export const MOCK_APPS: AppInfo[] = [
  {
    package: 'com.example.mock.adware',
    label: 'Flashlight HD (mock)',
    versionName: '2.4.1',
    versionCode: '2040100',
    isSystem: false,
    isSideloaded: true,
    installerPackage: 'com.android.packageinstaller',
    targetSdk: 28,
    minSdk: 21,
    installedAt: '2026-10-10T08:30:00Z',
    updatedAt: '2026-10-12T14:15:00Z',
    permissions: [
      'android.permission.INTERNET',
      'android.permission.ACCESS_NETWORK_STATE',
      'android.permission.SYSTEM_ALERT_WINDOW',
    ],
    dangerousPermissions: ['android.permission.SYSTEM_ALERT_WINDOW'],
    hasOverlayPermission: true,
    hasAccessibilityPermission: false,
    isDeviceAdmin: false,
    isBatteryOptimized: false,
    signatures: ['c8a0c2e391b16f1947b719468d6d6783'],
    suspiciousFlags: ['sideloaded', 'overlay_granted', 'hidden_launcher'],
  },
  {
    package: 'com.example.mock.popup',
    label: 'Cleaner Master (mock)',
    versionName: '1.0.5',
    versionCode: '10005',
    isSystem: false,
    isSideloaded: true,
    installerPackage: 'com.google.android.packageinstaller',
    targetSdk: 29,
    minSdk: 22,
    installedAt: '2026-10-11T11:20:00Z',
    updatedAt: '2026-10-11T11:20:00Z',
    permissions: [
      'android.permission.INTERNET',
      'android.permission.SYSTEM_ALERT_WINDOW',
      'android.permission.RECEIVE_BOOT_COMPLETED',
    ],
    dangerousPermissions: ['android.permission.SYSTEM_ALERT_WINDOW'],
    hasOverlayPermission: true,
    hasAccessibilityPermission: false,
    isDeviceAdmin: false,
    isBatteryOptimized: false,
    signatures: ['a1b2c3d4e5f60718293a4b5c6d7e8f90'],
    suspiciousFlags: ['aggressive_overlay', 'boot_receiver'],
  },
  {
    package: 'com.example.mock.dropper',
    label: 'Video Player Pro (mock)',
    versionName: '3.1.0',
    versionCode: '30100',
    isSystem: false,
    isSideloaded: true,
    targetSdk: 27,
    minSdk: 21,
    installedAt: '2026-10-09T09:00:00Z',
    updatedAt: '2026-10-09T09:00:00Z',
    permissions: [
      'android.permission.INTERNET',
      'android.permission.REQUEST_INSTALL_PACKAGES',
    ],
    dangerousPermissions: ['android.permission.REQUEST_INSTALL_PACKAGES'],
    hasOverlayPermission: false,
    hasAccessibilityPermission: false,
    isDeviceAdmin: false,
    isBatteryOptimized: true,
    signatures: ['fedcba9876543210fedcba9876543210'],
    suspiciousFlags: ['install_packages_permission', 'low_target_sdk'],
  },
  {
    package: 'com.android.settings',
    label: 'Settings',
    versionName: '12',
    versionCode: '31',
    isSystem: true,
    isSideloaded: false,
    targetSdk: 31,
    minSdk: 31,
    installedAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    permissions: ['android.permission.WRITE_SETTINGS'],
    dangerousPermissions: [],
    hasOverlayPermission: false,
    hasAccessibilityPermission: false,
    isDeviceAdmin: false,
    isBatteryOptimized: false,
    signatures: ['00112233445566778899aabbccddeeff'],
    suspiciousFlags: [],
  },
];

export const MOCK_FINDINGS: Finding[] = [
  {
    id: 'find-mock-01',
    ruleId: 'ADW-003',
    title: 'Sideloaded app with overlay and auto-start',
    severity: 'high',
    confidence: 'high',
    package: 'com.example.mock.adware',
    whyFlagged:
      'Application was sideloaded and holds SYSTEM_ALERT_WINDOW privilege without a regular launcher entry.',
    falsePositiveHint: 'Legitimate utilities may draw overlays with user consent.',
    evidence: [
      {
        type: 'permission',
        title: 'Overlay Permission Granted',
        details: 'SYSTEM_ALERT_WINDOW mode ALLOW in appops',
        rawOutput: 'appops get com.example.mock.adware SYSTEM_ALERT_WINDOW -> allow',
      },
    ],
    recommendedAction: {
      id: 'act-mock-01',
      type: 'disable',
      package: 'com.example.mock.adware',
      title: 'Disable com.example.mock.adware',
      impact: 'Application will be frozen and cannot draw overlays',
      undoable: true,
      isBlocked: false,
    },
  },
  {
    id: 'find-mock-02',
    ruleId: 'POP-001',
    title: 'Aggressive Out-of-Context Overlay Detected',
    severity: 'critical',
    confidence: 'high',
    package: 'com.example.mock.popup',
    whyFlagged:
      'Displayed full-screen TYPE_APPLICATION_OVERLAY window while user was in Android Settings.',
    falsePositiveHint: 'Incoming call dialer or alarm screen overlays.',
    evidence: [
      {
        type: 'setting',
        title: 'Window Manager Overlay Trace',
        details: 'TYPE_APPLICATION_OVERLAY created by UID 10182',
        rawOutput: 'mCurrentFocus=Window{c89f1d0 u0 com.example.mock.popup/AdActivity}',
      },
    ],
    recommendedAction: {
      id: 'act-mock-02',
      type: 'uninstall',
      package: 'com.example.mock.popup',
      title: 'Uninstall com.example.mock.popup',
      impact: 'Removes application package completely',
      undoable: true,
      isBlocked: false,
    },
  },
  {
    id: 'find-mock-03',
    ruleId: 'DRP-002',
    title: 'App requesting package installation with old targetSdk',
    severity: 'medium',
    confidence: 'medium',
    package: 'com.example.mock.dropper',
    whyFlagged:
      'Holds REQUEST_INSTALL_PACKAGES while targeting legacy Android SDK (targetSdk 27).',
    falsePositiveHint: 'Third-party app stores or browser download managers.',
    evidence: [
      {
        type: 'manifest',
        title: 'Manifest Permission Request',
        details: 'REQUEST_INSTALL_PACKAGES declared',
      },
    ],
  },
  {
    id: 'find-mock-04',
    ruleId: 'SYS-001',
    title: 'ADB Over Wi-Fi Port 5555 Listening',
    severity: 'low',
    confidence: 'high',
    // package is undefined: system-level finding per Point E
    whyFlagged:
      'Wireless ADB debugging is enabled on port 5555 without authentication restrictions.',
    falsePositiveHint: 'Developer actively testing over Wi-Fi.',
    evidence: [
      {
        type: 'setting',
        title: 'TCP Port Config',
        details: 'service.adb.tcp.port is set to 5555',
        rawOutput: 'getprop service.adb.tcp.port -> 5555',
      },
    ],
  },
  {
    id: 'find-mock-05',
    ruleId: 'INF-001',
    title: 'Ad Network SDK Present in Benign Application',
    severity: 'info',
    confidence: 'low',
    package: 'com.example.mock.adware',
    whyFlagged: 'Standard advertising SDK classes identified in dex classes.',
    falsePositiveHint: 'Almost all ad-supported free applications contain Ad SDKs.',
    evidence: [
      {
        type: 'class_prefix',
        title: 'Ad SDK Prefix',
        details: 'com.google.android.gms.ads found in dex',
      },
    ],
  },
];

export const MOCK_AUDIT_ITEMS: AuditItem[] = [
  {
    id: 'audit-01',
    category: 'system',
    title: 'USB Debugging Authorization',
    status: 'ok',
    summary: 'ADB host key is authenticated via RSA',
    evidence: ['ro.boot.flash.locked=1'],
  },
  {
    id: 'audit-02',
    category: 'connectivity',
    title: 'ADB over Wi-Fi Enabled',
    status: 'warning',
    summary: 'Port 5555 listening on local network',
    evidence: ['service.adb.tcp.port=5555'],
    remedy: 'Disable Wireless Debugging in Developer Options',
  },
  {
    id: 'audit-03',
    category: 'security',
    title: 'Unknown Sources Installation',
    status: 'issue',
    summary: 'At least one third-party package has sideload install privileges',
    evidence: ['com.example.mock.dropper has REQUEST_INSTALL_PACKAGES'],
    remedy: 'Revoke install unknown apps permission in Settings',
  },
  {
    id: 'audit-04',
    category: 'oem',
    title: 'OEM Private DNS Security Profile',
    status: 'unavailable',
    summary: 'Private DNS inspection is restricted by OEM security policies on API 31',
    evidence: ['getprop net.dns.private -> permission denied'],
  },
];

// ==========================================
// MockBridge Implementation
// ==========================================

type AnyCallback = (payload: never) => void;

export class MockBridge implements Bridge {
  private listeners: Map<BridgeEventKey, Set<AnyCallback>> = new Map();
  private timers: Set<ReturnType<typeof setTimeout>> = new Set();

  private activeDevice: Device | null = { ...MOCK_DEVICE };
  private preflightChecks: PreflightCheck[] = [...MOCK_PREFLIGHT_CHECKS];
  private apps: AppInfo[] = [...MOCK_APPS];
  private scanResults: Map<string, ScanResult> = new Map();
  private activeScanId: string | null = null;
  private monitorIsActive = false;
  private monitorEvents: MonitorEvent[] = [];
  private monitorCandidates: MonitorCandidate[] = [];
  private actionHistory: ActionLogEntry[] = [];
  private allowlistEntries: AllowlistEntry[] = [
    {
      id: 'allow-01',
      type: 'package',
      target: 'com.android.settings',
      addedAt: '2026-10-13T10:00:00Z',
      reason: 'Default Android Settings whitelist',
    },
  ];
  private settingsState: Settings = {
    theme: 'dark',
    adbPath: 'C:\\platform-tools\\adb.exe',
    customAdbPort: 5037,
    deepScanTimeoutSec: 120,
    autoCheckPreflight: true,
  };

  constructor() {
    // Seed initial completed scan
    const initialSummary: SecuritySummary = {
      counts: { info: 1, low: 1, medium: 1, high: 1, critical: 1 },
      totalApps: this.apps.length,
      lastScanAt: '2026-10-13T12:00:15Z',
      verdict: 'high_risk',
    };
    const initialScan: ScanResult = {
      id: 'scan-mock-seed-01',
      deviceId: this.activeDevice?.id ?? 'dev-mock-01',
      mode: 'quick',
      startedAt: '2026-10-13T12:00:00Z',
      completedAt: '2026-10-13T12:00:15Z',
      status: 'completed',
      findings: [...MOCK_FINDINGS],
      summary: initialSummary,
    };
    this.scanResults.set(initialScan.id, initialScan);
  }

  // Event System
  on<E extends BridgeEventKey>(
    event: E,
    callback: BridgeEventCallback<E>
  ): Unsubscribe {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    const set = this.listeners.get(event)!;
    set.add(callback as AnyCallback);

    return () => {
      set.delete(callback as AnyCallback);
    };
  }

  emit<E extends BridgeEventKey>(event: E, payload: BridgeEventMap[E]): void {
    const set = this.listeners.get(event);
    if (set) {
      for (const cb of set) {
        try {
          cb(payload as never);
        } catch (e) {
          console.error(`MockBridge event handler error on ${event}:`, e);
        }
      }
    }
  }

  destroy(): void {
    for (const timer of this.timers) {
      clearTimeout(timer);
    }
    this.timers.clear();
    this.listeners.clear();
  }

  private trackTimer(timer: ReturnType<typeof setTimeout>): ReturnType<typeof setTimeout> {
    this.timers.add(timer);
    return timer;
  }

  // Preflight Service
  readonly preflight: PreflightService = {
    run: async (): Promise<PreflightCheck[]> => {
      return [...this.preflightChecks];
    },
    status: async (): Promise<PreflightCheck[]> => {
      return [...this.preflightChecks];
    },
    downloadAdb: async (): Promise<void> => {
      const idx = this.preflightChecks.findIndex((c) => c.id === 'adb_binary');
      if (idx !== -1) {
        this.preflightChecks[idx] = {
          ...this.preflightChecks[idx],
          status: 'downloading',
          downloadProgress: 25,
        };
        this.emit('preflight:update', this.preflightChecks[idx]);

        await new Promise<void>((resolve) => {
          this.trackTimer(
            setTimeout(() => {
              this.preflightChecks[idx] = {
                ...this.preflightChecks[idx],
                status: 'ok',
                downloadProgress: 100,
                detail: 'Android Debug Bridge version 1.0.41 (downloaded and verified)',
              };
              this.emit('preflight:update', this.preflightChecks[idx]);
              resolve();
            }, 100)
          );
        });
      }
    },
    selectAdbPath: async (path: string): Promise<void> => {
      const idx = this.preflightChecks.findIndex((c) => c.id === 'adb_binary');
      if (idx !== -1) {
        this.preflightChecks[idx] = {
          ...this.preflightChecks[idx],
          status: 'ok',
          detail: `Configured at ${path}`,
        };
        this.emit('preflight:update', this.preflightChecks[idx]);
      }
    },
    retry: async (): Promise<PreflightCheck[]> => {
      return this.preflight.run();
    },
  };

  // Device Service
  readonly device: DeviceService = {
    getConnected: async (): Promise<Device | null> => {
      return this.activeDevice ? { ...this.activeDevice } : null;
    },
    list: async (): Promise<Device[]> => {
      return this.activeDevice ? [{ ...this.activeDevice }] : [];
    },
    select: async (serial: string): Promise<void> => {
      if (this.activeDevice && this.activeDevice.serial === serial) {
        this.emit('device:changed', { ...this.activeDevice });
      } else {
        throw new BridgeError('DeviceOffline', `Device with serial ${serial} is not connected`);
      }
    },
  };

  // Inventory Service
  readonly inventory: InventoryService = {
    list: async (filter?: InventoryFilter): Promise<AppInfo[]> => {
      let result = [...this.apps];
      if (!filter) return result;

      if (filter.search) {
        const query = filter.search.toLowerCase();
        result = result.filter(
          (a) => a.package.toLowerCase().includes(query) || a.label.toLowerCase().includes(query)
        );
      }

      if (filter.category === 'user') {
        result = result.filter((a) => !a.isSystem);
      } else if (filter.category === 'system') {
        result = result.filter((a) => a.isSystem);
      } else if (filter.category === 'risky') {
        result = result.filter((a) => a.suspiciousFlags.length > 0);
      }

      return result;
    },
    get: async (packageName: string): Promise<AppInfo | null> => {
      const found = this.apps.find((a) => a.package === packageName);
      return found ? { ...found } : null;
    },
    extractApk: async (packageName: string, targetPath?: string): Promise<string> => {
      const app = this.apps.find((a) => a.package === packageName);
      if (!app) {
        throw new BridgeError('ParseFailure', `Package ${packageName} not found on device`);
      }
      return targetPath ?? `pulled_apks/${packageName}_${app.versionCode}.apk`;
    },
  };

  // Scan Service
  readonly scan: ScanService = {
    start: async (options: ScanOptions): Promise<string> => {
      if (!this.activeDevice) {
        throw new BridgeError('DeviceOffline', 'No active device connected');
      }

      const scanId = `scan-mock-${Date.now()}`;
      this.activeScanId = scanId;

      const initialResult: ScanResult = {
        id: scanId,
        deviceId: this.activeDevice.id,
        mode: options.mode,
        startedAt: new Date().toISOString(),
        status: 'running',
        findings: [],
        summary: {
          counts: { info: 0, low: 0, medium: 0, high: 0, critical: 0 },
          totalApps: this.apps.length,
          verdict: 'clean',
        },
      };
      this.scanResults.set(scanId, initialResult);

      // Simulation steps using timers
      this.trackTimer(
        setTimeout(() => {
          if (this.activeScanId !== scanId) return;
          const prog1: ScanProgress = {
            step: 'Inspecting package manifests and signatures',
            progress: 25,
            currentPackage: this.apps[0]?.package,
          };
          this.emit('scan:progress', prog1);
        }, 50)
      );

      this.trackTimer(
        setTimeout(() => {
          if (this.activeScanId !== scanId) return;
          const finding1 = MOCK_FINDINGS[0];
          initialResult.findings.push(finding1);
          this.emit('scan:finding', finding1);

          const prog2: ScanProgress = {
            step: 'Analyzing overlay and background receivers',
            progress: 60,
            currentPackage: this.apps[1]?.package,
          };
          this.emit('scan:progress', prog2);
        }, 100)
      );

      this.trackTimer(
        setTimeout(() => {
          if (this.activeScanId !== scanId) return;
          const finding2 = MOCK_FINDINGS[1];
          initialResult.findings.push(finding2);
          this.emit('scan:finding', finding2);

          const prog3: ScanProgress = {
            step: 'Checking system-level configurations and ports',
            progress: 90,
          };
          this.emit('scan:progress', prog3);
        }, 150)
      );

      this.trackTimer(
        setTimeout(() => {
          if (this.activeScanId !== scanId) return;
          const completedResult: ScanResult = {
            ...initialResult,
            completedAt: new Date().toISOString(),
            status: 'completed',
            summary: {
              counts: {
                info: 0,
                low: 0,
                medium: 0,
                high: 1,
                critical: 1,
              },
              totalApps: this.apps.length,
              lastScanAt: new Date().toISOString(),
              verdict: 'high_risk',
            },
          };
          this.scanResults.set(scanId, completedResult);
          this.emit('scan:done', completedResult);
          this.activeScanId = null;
        }, 200)
      );

      return scanId;
    },
    status: async (scanId: string): Promise<ScanResult | null> => {
      const res = this.scanResults.get(scanId);
      return res ? { ...res } : null;
    },
    cancel: async (scanId: string): Promise<void> => {
      const res = this.scanResults.get(scanId);
      if (res && res.status === 'running') {
        const cancelled: ScanResult = {
          ...res,
          status: 'cancelled',
          completedAt: new Date().toISOString(),
        };
        this.scanResults.set(scanId, cancelled);
        if (this.activeScanId === scanId) {
          this.activeScanId = null;
        }
      }
    },
    get: async (scanId: string): Promise<ScanResult | null> => {
      const res = this.scanResults.get(scanId);
      return res ? { ...res } : null;
    },
    quickScan: async (): Promise<ScanResult> => {
      const scanId = await this.scan.start({ mode: 'quick', includeSystem: false });
      return new Promise<ScanResult>((resolve) => {
        const unsub = this.on('scan:done', (result) => {
          if (result.id === scanId) {
            unsub();
            resolve(result);
          }
        });
      });
    },
  };

  // Monitor Service
  readonly monitor: MonitorService = {
    start: async (): Promise<void> => {
      this.monitorIsActive = true;
      const initialEvent: MonitorEvent = {
        id: `evt-mock-${Date.now()}`,
        timestamp: new Date().toISOString(),
        type: 'overlay_shown',
        package: 'com.example.mock.popup',
        details: { surface: '1080x2400', type: 'TYPE_APPLICATION_OVERLAY' },
      };
      this.monitorEvents.push(initialEvent);
      this.emit('monitor:event', initialEvent);

      const candidate: MonitorCandidate = {
        package: 'com.example.mock.popup',
        appLabel: 'Cleaner Master (mock)',
        confidence: 'high',
        score: 90,
        timelineEvents: [initialEvent],
        whyFlagged: 'Full-screen overlay created out-of-context',
      };
      this.monitorCandidates.push(candidate);
      this.emit('monitor:candidate', candidate);
    },
    stop: async (): Promise<void> => {
      this.monitorIsActive = false;
    },
    snapshot: async (): Promise<MonitorSnapshot> => {
      const status: MonitorStatus = {
        isActive: this.monitorIsActive,
        startedAt: this.monitorIsActive ? new Date().toISOString() : undefined,
        sourcesActive: ['dumpsys window', 'logcat'],
      };
      return {
        status,
        events: [...this.monitorEvents],
        candidates: [...this.monitorCandidates],
      };
    },
  };

  // Audit Service
  readonly audit: AuditService = {
    run: async (): Promise<AuditItem[]> => {
      return [...MOCK_AUDIT_ITEMS];
    },
    getLatest: async (): Promise<AuditItem[]> => {
      return [...MOCK_AUDIT_ITEMS];
    },
  };

  // Remediate Service
  readonly remediate: RemediateService = {
    createPlan: async (findingIds: string[]): Promise<RemediationPlan> => {
      const actions: RemediationAction[] = [];
      for (const id of findingIds) {
        const finding = MOCK_FINDINGS.find((f) => f.id === id);
        if (finding?.recommendedAction) {
          actions.push(finding.recommendedAction);
        }
      }
      return { actions, requiresBackup: true };
    },
    execute: async (plan: RemediationPlan): Promise<ActionLogEntry[]> => {
      const entries: ActionLogEntry[] = [];
      for (const action of plan.actions) {
        if (action.isBlocked) {
          throw new BridgeError('ActionBlocked', action.blockedReason ?? 'Action is blocked');
        }
        const entry: ActionLogEntry = {
          id: `log-mock-${Date.now()}-${action.id}`,
          timestamp: new Date().toISOString(),
          action,
          status: 'success',
          backupPath: `backups/${action.package}.apk`,
        };
        entries.push(entry);
        this.actionHistory.unshift(entry);
      }
      return entries;
    },
    getHistory: async (): Promise<ActionLogEntry[]> => {
      return [...this.actionHistory];
    },
    undo: async (logEntryId: string): Promise<ActionLogEntry> => {
      const idx = this.actionHistory.findIndex((e) => e.id === logEntryId);
      if (idx === -1) {
        throw new BridgeError('ParseFailure', `Action log entry ${logEntryId} not found`);
      }
      this.actionHistory[idx] = {
        ...this.actionHistory[idx],
        status: 'undone',
      };
      return { ...this.actionHistory[idx] };
    },
  };

  // Report Service
  readonly report: ReportService = {
    list: async (): Promise<ScanSummaryItem[]> => {
      const summaries: ScanSummaryItem[] = [];
      for (const scan of this.scanResults.values()) {
        summaries.push({
          id: scan.id,
          startedAt: scan.startedAt,
          completedAt: scan.completedAt,
          mode: scan.mode,
          verdict: scan.summary.verdict,
          findingsCount: scan.summary.counts,
        });
      }
      return summaries;
    },
    get: async (id: string): Promise<ScanResult | null> => {
      const scan = this.scanResults.get(id);
      return scan ? { ...scan } : null;
    },
    compare: async (scanIdA: string, scanIdB: string): Promise<ScanComparison> => {
      const scanA = this.scanResults.get(scanIdA);
      const scanB = this.scanResults.get(scanIdB);
      if (!scanA || !scanB) {
        throw new BridgeError('ParseFailure', 'One or both scan IDs not found for comparison');
      }

      const summaryA: ScanSummaryItem = {
        id: scanA.id,
        startedAt: scanA.startedAt,
        completedAt: scanA.completedAt,
        mode: scanA.mode,
        verdict: scanA.summary.verdict,
        findingsCount: scanA.summary.counts,
      };

      const summaryB: ScanSummaryItem = {
        id: scanB.id,
        startedAt: scanB.startedAt,
        completedAt: scanB.completedAt,
        mode: scanB.mode,
        verdict: scanB.summary.verdict,
        findingsCount: scanB.summary.counts,
      };

      const idsB = new Set(scanB.findings.map((f) => f.id));
      const idsA = new Set(scanA.findings.map((f) => f.id));

      const newFindings = scanB.findings.filter((f) => !idsA.has(f.id));
      const resolvedFindings = scanA.findings.filter((f) => !idsB.has(f.id));
      const unchangedFindings = scanA.findings.filter((f) => idsB.has(f.id));

      return {
        scanA: summaryA,
        scanB: summaryB,
        newFindings,
        resolvedFindings,
        unchangedFindings,
      };
    },
    export: async (options: ReportExportOptions): Promise<string> => {
      const scan = this.scanResults.get(options.scanId);
      if (!scan) {
        throw new BridgeError('ParseFailure', `Scan ID ${options.scanId} not found`);
      }
      return options.targetPath ?? `reports/scan_${options.scanId}.${options.format}`;
    },
  };

  // Settings Service
  readonly settings: SettingsService = {
    get: async (): Promise<Settings> => {
      return { ...this.settingsState };
    },
    update: async (partial: Partial<Settings>): Promise<Settings> => {
      this.settingsState = { ...this.settingsState, ...partial };
      return { ...this.settingsState };
    },
  };

  // Allowlist Service
  readonly allowlist: AllowlistService = {
    list: async (): Promise<AllowlistEntry[]> => {
      return [...this.allowlistEntries];
    },
    add: async (entry: Omit<AllowlistEntry, 'id' | 'addedAt'>): Promise<AllowlistEntry> => {
      const created: AllowlistEntry = {
        ...entry,
        id: `allow-${Date.now()}`,
        addedAt: new Date().toISOString(),
      };
      this.allowlistEntries.push(created);
      return created;
    },
    remove: async (id: string): Promise<void> => {
      this.allowlistEntries = this.allowlistEntries.filter((e) => e.id !== id);
    },
  };
}
