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

function notImplemented(methodName: string): never {
  throw new BridgeError(
    'NotSupportedOnThisOS',
    `Wails backend binding '${methodName}' is not implemented yet (scheduled for Phase T4/T5).`,
    'Wait for backend implementation in Phase T4'
  );
}

export class WailsBridge implements Bridge {
  readonly preflight: PreflightService = {
    run: async () => notImplemented('preflight.run'),
    status: async () => notImplemented('preflight.status'),
    downloadAdb: async () => notImplemented('preflight.downloadAdb'),
    selectAdbPath: async () => notImplemented('preflight.selectAdbPath'),
    retry: async () => notImplemented('preflight.retry'),
  };

  readonly device: DeviceService = {
    getConnected: async () => notImplemented('device.getConnected'),
    list: async () => notImplemented('device.list'),
    select: async () => notImplemented('device.select'),
  };

  readonly inventory: InventoryService = {
    list: async () => notImplemented('inventory.list'),
    get: async () => notImplemented('inventory.get'),
    extractApk: async () => notImplemented('inventory.extractApk'),
  };

  readonly scan: ScanService = {
    start: async () => notImplemented('scan.start'),
    status: async () => notImplemented('scan.status'),
    cancel: async () => notImplemented('scan.cancel'),
    get: async () => notImplemented('scan.get'),
    quickScan: async () => notImplemented('scan.quickScan'),
  };

  readonly monitor: MonitorService = {
    start: async () => notImplemented('monitor.start'),
    stop: async () => notImplemented('monitor.stop'),
    snapshot: async () => notImplemented('monitor.snapshot'),
  };

  readonly audit: AuditService = {
    run: async () => notImplemented('audit.run'),
    getLatest: async () => notImplemented('audit.getLatest'),
  };

  readonly remediate: RemediateService = {
    createPlan: async () => notImplemented('remediate.createPlan'),
    execute: async () => notImplemented('remediate.execute'),
    getHistory: async () => notImplemented('remediate.getHistory'),
    undo: async () => notImplemented('remediate.undo'),
  };

  readonly report: ReportService = {
    list: async () => notImplemented('report.list'),
    get: async () => notImplemented('report.get'),
    compare: async () => notImplemented('report.compare'),
    export: async () => notImplemented('report.export'),
  };

  readonly settings: SettingsService = {
    get: async () => notImplemented('settings.get'),
    update: async () => notImplemented('settings.update'),
  };

  readonly allowlist: AllowlistService = {
    list: async () => notImplemented('allowlist.list'),
    add: async () => notImplemented('allowlist.add'),
    remove: async () => notImplemented('allowlist.remove'),
  };

  on<E extends BridgeEventKey>(
    _event: E,
    _callback: BridgeEventCallback<E>
  ): Unsubscribe {
    // In production Wails v2, events will bind to runtime.EventsOn / EventsOff
    return () => {};
  }
}
