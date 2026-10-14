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

export interface PreflightService {
  run(): Promise<PreflightCheck[]>;
  status(): Promise<PreflightCheck[]>;
  downloadAdb(): Promise<void>;
  selectAdbPath(path: string): Promise<void>;
  retry(): Promise<PreflightCheck[]>;
}

export interface DeviceService {
  getConnected(): Promise<Device | null>;
  list(): Promise<Device[]>;
  select(serial: string): Promise<void>;
}

export interface InventoryService {
  list(filter?: InventoryFilter): Promise<AppInfo[]>;
  get(packageName: string): Promise<AppInfo | null>;
  extractApk(packageName: string, targetPath?: string): Promise<string>;
}

export interface ScanService {
  start(options: ScanOptions): Promise<string>;
  status(scanId: string): Promise<ScanResult | null>;
  cancel(scanId: string): Promise<void>;
  get(scanId: string): Promise<ScanResult | null>;
  quickScan(): Promise<ScanResult>;
}

export interface MonitorService {
  start(): Promise<void>;
  stop(): Promise<void>;
  snapshot(): Promise<MonitorSnapshot>;
  markPopup(): Promise<void>;
}

export interface AuditService {
  run(): Promise<AuditItem[]>;
  getLatest(): Promise<AuditItem[]>;
}

export interface RemediateService {
  createPlan(findingIds: string[]): Promise<RemediationPlan>;
  execute(plan: RemediationPlan): Promise<ActionLogEntry[]>;
  getHistory(): Promise<ActionLogEntry[]>;
  undo(logEntryId: string): Promise<ActionLogEntry>;
}

export interface ReportService {
  list(): Promise<ScanSummaryItem[]>;
  get(id: string): Promise<ScanResult | null>;
  compare(scanIdA: string, scanIdB: string): Promise<ScanComparison>;
  export(options: ReportExportOptions): Promise<string>;
}

export interface SettingsService {
  get(): Promise<Settings>;
  update(partial: Partial<Settings>): Promise<Settings>;
}

export interface AllowlistService {
  list(): Promise<AllowlistEntry[]>;
  add(entry: Omit<AllowlistEntry, 'id' | 'addedAt'>): Promise<AllowlistEntry>;
  remove(id: string): Promise<void>;
}

export interface Bridge {
  readonly preflight: PreflightService;
  readonly device: DeviceService;
  readonly inventory: InventoryService;
  readonly scan: ScanService;
  readonly monitor: MonitorService;
  readonly audit: AuditService;
  readonly remediate: RemediateService;
  readonly report: ReportService;
  readonly settings: SettingsService;
  readonly allowlist: AllowlistService;

  on<E extends BridgeEventKey>(
    event: E,
    callback: BridgeEventCallback<E>
  ): Unsubscribe;
}
