import type { Device, ScanResult } from '../lib/bridge/schema';

export type CanonicalSecurityState =
  | 'no-device'
  | 'unauthorized'
  | 'offline'
  | 'never-scanned'
  | 'scanning'
  | 'scanned-clean'
  | 'scanned-findings'
  | 'partial-failed';

export interface CanonicalSecurityContext {
  state: CanonicalSecurityState;
  device: Device | null;
  scanResult: ScanResult | null;
  findingsCount: number;
  totalApps: number;
  scanMode?: 'quick' | 'deep';
  scanTime?: string;
  ruleVersion?: string;
  contextText: string;
  allowsCleanVerdict: boolean;
  canStartScan: boolean;
  disabledReason?: string;
}

function formatScanTime(isoString?: string): string {
  if (!isoString) return '';
  try {
    const d = new Date(isoString);
    if (isNaN(d.getTime())) return '';
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
  } catch {
    return '';
  }
}

/**
 * Derives the canonical security state for the currently active session.
 * Enforces strict device-serial binding: scan results belonging to a different device
 * are ignored so they never falsely mark a newly connected device as scanned or clean.
 */
export function getCanonicalSecurityState(
  device: Device | null,
  scanResult: ScanResult | null,
  isScanning: boolean = false,
  activeRuleVersion?: string
): CanonicalSecurityContext {
  const ruleVer =
    activeRuleVersion ||
    (scanResult as unknown as { rulesVersion?: string })?.rulesVersion ||
    '';
  const rulePart = ruleVer
    ? ` · rules ${ruleVer.startsWith('v') ? ruleVer : `v${ruleVer}`}`
    : '';

  // 1. No device connected
  if (!device) {
    return {
      state: 'no-device',
      device: null,
      scanResult: null,
      findingsCount: 0,
      totalApps: 0,
      ruleVersion: ruleVer || undefined,
      contextText: 'No device connected',
      allowsCleanVerdict: false,
      canStartScan: false,
      disabledReason: 'Connect an Android device via USB to start scanning',
    };
  }

  // 2. Device connected but unauthorized (RSA prompt on phone pending)
  if (device.connectionState === 'unauthorized') {
    return {
      state: 'unauthorized',
      device,
      scanResult: null,
      findingsCount: 0,
      totalApps: 0,
      ruleVersion: ruleVer || undefined,
      contextText: 'Authorization required · Allow USB debugging on device screen',
      allowsCleanVerdict: false,
      canStartScan: false,
      disabledReason: 'Authorize USB debugging on your Android phone to start scanning',
    };
  }

  // 3. Device connected but offline in ADB daemon
  if (device.connectionState === 'offline') {
    return {
      state: 'offline',
      device,
      scanResult: null,
      findingsCount: 0,
      totalApps: 0,
      ruleVersion: ruleVer || undefined,
      contextText: 'Device offline · Check USB connection',
      allowsCleanVerdict: false,
      canStartScan: false,
      disabledReason: 'Device is offline in ADB daemon',
    };
  }

  // 4. Device is ready, and a scan is currently running
  if (isScanning) {
    return {
      state: 'scanning',
      device,
      scanResult: null,
      findingsCount: 0,
      totalApps: 0,
      ruleVersion: ruleVer || undefined,
      contextText: 'Scanning in progress...',
      allowsCleanVerdict: false,
      canStartScan: false,
      disabledReason: 'A scan is currently running',
    };
  }

  // 5. Device is ready, check whether the scanResult strictly matches this device's serial or id
  const hasMatchingScan = Boolean(
    scanResult &&
      scanResult.deviceId &&
      scanResult.deviceId.trim() !== '' &&
      (scanResult.deviceId === device.serial ||
        (device.id && scanResult.deviceId === device.id))
  );

  if (!hasMatchingScan || !scanResult) {
    return {
      state: 'never-scanned',
      device,
      scanResult: null,
      findingsCount: 0,
      totalApps: 0,
      ruleVersion: ruleVer || undefined,
      contextText: `Device ready · No scan performed yet${rulePart}`,
      allowsCleanVerdict: false,
      canStartScan: true,
    };
  }

  // 6. Terminal but non-completed scan states (failed, cancelled, partial)
  if (
    scanResult.status === 'failed' ||
    scanResult.status === 'cancelled' ||
    scanResult.status === 'partial'
  ) {
    const statusLabel =
      scanResult.status === 'cancelled'
        ? 'Scan cancelled'
        : scanResult.status === 'failed'
        ? 'Scan failed'
        : 'Partial scan (incomplete)';
    return {
      state: 'partial-failed',
      device,
      scanResult,
      findingsCount: scanResult.findings?.length ?? 0,
      totalApps: scanResult.summary?.totalApps ?? 0,
      scanMode: scanResult.mode,
      scanTime: formatScanTime(scanResult.completedAt || scanResult.startedAt),
      ruleVersion: ruleVer || undefined,
      contextText: `${statusLabel}${rulePart}`,
      allowsCleanVerdict: false,
      canStartScan: true,
    };
  }

  // 7. Completed scan: evaluate findings
  const findingsCount = scanResult.findings?.length ?? 0;
  const totalApps = scanResult.summary?.totalApps ?? 0;
  const scanMode = scanResult.mode || 'quick';
  const timeFormatted = formatScanTime(scanResult.completedAt || scanResult.startedAt);
  const timePart = timeFormatted ? ` · ${timeFormatted}` : '';
  const appsPart = totalApps > 0 ? ` · ${totalApps} apps` : '';
  const modePart = `${scanMode.charAt(0).toUpperCase() + scanMode.slice(1)} scan`;

  if (findingsCount === 0) {
    return {
      state: 'scanned-clean',
      device,
      scanResult,
      findingsCount: 0,
      totalApps,
      scanMode,
      scanTime: timeFormatted,
      ruleVersion: ruleVer || undefined,
      contextText: `No findings in the last scan · ${modePart}${appsPart}${timePart}${rulePart}`,
      allowsCleanVerdict: true, // Only state allowed to render clean checkmark
      canStartScan: true,
    };
  }

  return {
    state: 'scanned-findings',
    device,
    scanResult,
    findingsCount,
    totalApps,
    scanMode,
    scanTime: timeFormatted,
    ruleVersion: ruleVer || undefined,
    contextText: `${findingsCount} findings · ${modePart}${appsPart}${timePart}${rulePart}`,
    allowsCleanVerdict: false,
    canStartScan: true,
  };
}
