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
import type { BridgeEventCallback, BridgeEventKey, Unsubscribe } from './events';
import type {
  ActionLogEntry,
  AllowlistEntry,
  AppInfo,
  AuditItem,
  Device,
  InventoryFilter,
  MonitorSnapshot,
  PreflightCheck,
  RemediationPlan,
  ReportExportOptions,
  ScanComparison,
  ScanOptions,
  ScanResult,
  ScanSummaryItem,
  Settings,
} from './schema';

async function callWails<T>(fnName: string, ...args: unknown[]): Promise<T> {
  const app = (
    window as unknown as {
      go?: {
        main?: {
          App?: Record<string, (...a: unknown[]) => Promise<T>>;
        };
      };
    }
  )?.go?.main?.App;

  if (!app || typeof app[fnName] !== 'function') {
    throw new BridgeError(
      'NotSupportedOnThisOS',
      `Wails backend binding '${fnName}' is not available in current window runtime.`,
      'Ensure the desktop application is running inside Wails webview.'
    );
  }

  try {
    return await app[fnName](...args);
  } catch (err) {
    throw BridgeError.fromWire(err);
  }
}

export class WailsBridge implements Bridge {
  readonly preflight: PreflightService = {
    run: () => callWails<PreflightCheck[]>('PreflightRun'),
    status: () => callWails<PreflightCheck[]>('PreflightStatus'),
    downloadAdb: () => callWails<void>('PreflightDownloadAdb'),
    selectAdbPath: (path: string) => callWails<void>('PreflightSelectAdbPath', path),
    retry: () => callWails<PreflightCheck[]>('PreflightRetry'),
  };

  readonly device: DeviceService = {
    getConnected: () => callWails<Device | null>('DeviceGetConnected'),
    list: () => callWails<Device[]>('DeviceList'),
    select: (serial: string) => callWails<void>('DeviceSelect', serial),
  };

  readonly inventory: InventoryService = {
    list: (filter?: InventoryFilter) =>
      callWails<AppInfo[]>('InventoryList', filter ?? { category: 'all', search: '' }),
    get: (packageName: string) => callWails<AppInfo | null>('InventoryGet', packageName),
    extractApk: (packageName: string, targetPath?: string) =>
      callWails<string>('InventoryExtractApk', packageName, targetPath ?? ''),
  };

  readonly scan: ScanService = {
    start: (options: ScanOptions) => callWails<string>('ScanStart', options),
    status: (scanId: string) => callWails<ScanResult | null>('ScanStatus', scanId),
    cancel: (scanId: string) => callWails<void>('ScanCancel', scanId),
    get: (scanId: string) => callWails<ScanResult | null>('ScanGet', scanId),
    quickScan: () => callWails<ScanResult>('ScanQuick'),
  };

  readonly monitor: MonitorService = {
    start: () => callWails<void>('MonitorStart'),
    stop: () => callWails<void>('MonitorStop'),
    snapshot: () => callWails<MonitorSnapshot>('MonitorSnapshot'),
    markPopup: () => callWails<void>('MonitorMarkPopup'),
  };

  readonly audit: AuditService = {
    run: () => callWails<AuditItem[]>('AuditRun'),
    getLatest: () => callWails<AuditItem[]>('AuditGetLatest'),
  };

  readonly remediate: RemediateService = {
    createPlan: (findingIds: string[]) => callWails<RemediationPlan>('RemediateCreatePlan', findingIds),
    execute: (plan: RemediationPlan) => callWails<ActionLogEntry[]>('RemediateExecute', plan),
    getHistory: () => callWails<ActionLogEntry[]>('RemediateGetHistory'),
    undo: (logEntryId: string) => callWails<ActionLogEntry>('RemediateUndo', logEntryId),
  };

  readonly report: ReportService = {
    list: () => callWails<ScanSummaryItem[]>('ReportList'),
    get: (id: string) => callWails<ScanResult | null>('ReportGet', id),
    compare: (scanIdA: string, scanIdB: string) =>
      callWails<ScanComparison>('ReportCompare', scanIdA, scanIdB),
    export: (options: ReportExportOptions) => callWails<string>('ReportExport', options),
  };

  readonly settings: SettingsService = {
    get: () => callWails<Settings>('SettingsGet'),
    update: (partial: Partial<Settings>) => callWails<Settings>('SettingsUpdate', partial),
  };

  readonly allowlist: AllowlistService = {
    list: () => callWails<AllowlistEntry[]>('AllowlistList'),
    add: (entry: Omit<AllowlistEntry, 'id' | 'addedAt'>) =>
      callWails<AllowlistEntry>('AllowlistAdd', entry),
    remove: (id: string) => callWails<void>('AllowlistRemove', id),
  };

  on<E extends BridgeEventKey>(
    event: E,
    callback: BridgeEventCallback<E>
  ): Unsubscribe {
    const runtime = (
      window as unknown as {
        runtime?: {
          EventsOn?: (evt: string, cb: (...d: unknown[]) => void) => void;
          EventsOff?: (evt: string) => void;
        };
      }
    )?.runtime;

    if (!runtime || typeof runtime.EventsOn !== 'function') {
      return () => {};
    }

    const handler = (...data: unknown[]) => {
      callback(data[0] as Parameters<BridgeEventCallback<E>>[0]);
    };

    runtime.EventsOn(event, handler);
    return () => {
      runtime.EventsOff?.(event);
    };
  }
}
