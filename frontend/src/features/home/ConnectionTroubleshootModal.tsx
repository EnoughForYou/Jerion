import React, { useState, useEffect, useCallback } from 'react';
import {
  Wrench,
  RotateCcw,
  AlertTriangle,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  Wifi,
  Usb,
  Server,
  ShieldAlert,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/Dialog';
import { Button } from '../../components/ui/Button';
import { Badge } from '../../components/ui/Badge';
import { Card } from '../../components/ui/Card';
import { getBridge } from '../../lib/bridge';
import type { ConnectionDiagnosticResult } from '../../lib/bridge/schema';
import { BrowserOpenURL } from '../../../wailsjs/runtime';

export interface ConnectionTroubleshootModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOpenWireless?: () => void;
}

interface OEMDriverEntry {
  brand: string;
  url: string;
  note: string;
}

const OFFICIAL_OEM_DRIVERS: OEMDriverEntry[] = [
  {
    brand: 'Google (Pixel)',
    url: 'https://developer.android.com/studio/run/win-usb',
    note: 'Official Google USB Driver for Windows (WinUSB)',
  },
  {
    brand: 'Samsung',
    url: 'https://developer.samsung.com/android-usb-driver',
    note: 'Official Samsung Android USB Driver for Windows',
  },
  {
    brand: 'Xiaomi / Redmi / POCO',
    url: 'https://xiaomidriverpress.com/',
    note: 'Official Xiaomi USB / Fastboot Drivers',
  },
  {
    brand: 'Transsion (Infinix, TECNO, itel)',
    url: 'https://www.infinixmobility.com/support',
    note: 'Transsion & MediaTek USB VCOM / ADB Support',
  },
  {
    brand: 'OPPO / Realme',
    url: 'https://support.oppo.com/',
    note: 'OPPO Official Customer Care & USB Driver Support',
  },
  {
    brand: 'vivo / iQOO',
    url: 'https://www.vivo.com/en/support',
    note: 'vivo Official PC Assistant & Driver Suite',
  },
  {
    brand: 'Motorola',
    url: 'https://en-us.support.motorola.com/app/answers/detail/a_id/88481',
    note: 'Motorola Device Manager (Contains USB Drivers)',
  },
  {
    brand: 'ASUS',
    url: 'https://www.asus.com/support/',
    note: 'ASUS Driver & Utilities Support Page',
  },
];

export const ConnectionTroubleshootModal: React.FC<ConnectionTroubleshootModalProps> = ({
  open,
  onOpenChange,
  onOpenWireless,
}) => {
  const { t } = useTranslation();
  const bridge = getBridge();

  const [diagnostics, setDiagnostics] = useState<ConnectionDiagnosticResult | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [isSoftReconnecting, setIsSoftReconnecting] = useState(false);
  const [actionNotice, setActionNotice] = useState<string | null>(null);

  const [isRestartingServer, setIsRestartingServer] = useState(false);
  const [showRestartConfirm, setShowRestartConfirm] = useState(false);

  const runDiagnostics = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setActionNotice(null);
    try {
      const res = await bridge.device.diagnoseConnection();
      setDiagnostics(res);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to diagnose connection');
    } finally {
      setIsLoading(false);
    }
  }, [bridge]);

  useEffect(() => {
    let isMounted = true;
    if (open) {
      const initDiagnostics = async () => {
        try {
          const res = await bridge.device.diagnoseConnection();
          if (isMounted) {
            setDiagnostics(res);
            setIsLoading(false);
          }
        } catch (err: unknown) {
          if (isMounted) {
            setError(err instanceof Error ? err.message : 'Failed to diagnose connection');
            setIsLoading(false);
          }
        }
      };
      void initDiagnostics();
    }
    return () => {
      isMounted = false;
    };
  }, [open, bridge]);

  const handleSoftReconnect = async () => {
    try {
      setIsSoftReconnecting(true);
      setActionNotice(null);
      await bridge.device.softReconnect();
      setActionNotice('Soft reconnect completed. Refreshing diagnostic state...');
      await runDiagnostics();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Soft reconnect failed');
    } finally {
      setIsSoftReconnecting(false);
    }
  };

  const handleRestartServerConfirmed = async () => {
    try {
      setIsRestartingServer(true);
      setShowRestartConfirm(false);
      setActionNotice(null);
      await bridge.device.restartServer();
      setActionNotice(t('device.troubleshoot.restartSuccess', 'ADB server restarted successfully.'));
      await runDiagnostics();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Restart ADB server failed');
    } finally {
      setIsRestartingServer(false);
    }
  };

  const openExternal = (url: string) => {
    try {
      BrowserOpenURL(url);
    } catch {
      window.open(url, '_blank', 'noopener,noreferrer');
    }
  };

  const renderStatusBadge = () => {
    if (!diagnostics) return null;
    switch (diagnostics.likelyCause) {
      case 'ready':
        return <Badge severity="ok">{t('device.status.online')}</Badge>;
      case 'unauthorized':
        return <Badge severity="high">{t('device.status.unauthorized')}</Badge>;
      case 'offline':
        return <Badge severity="medium">{t('device.status.offline')}</Badge>;
      case 'usb_detected_no_adb':
        return <Badge severity="medium">USB Active (No ADB)</Badge>;
      case 'no_usb_hardware':
        return <Badge severity="info">No USB Hardware</Badge>;
      case 'adb_server_failed':
      default:
        return <Badge severity="critical">ADB Server Stopped</Badge>;
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col p-6 overflow-hidden">
          <DialogHeader className="shrink-0 pb-3 border-b border-separator-subtle">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-md bg-accent/15 text-accent flex items-center justify-center shrink-0">
                  <Wrench className="w-4 h-4" />
                </div>
                <div>
                  <DialogTitle>{t('device.troubleshoot.title')}</DialogTitle>
                  <DialogDescription className="text-xs">
                    {t('device.troubleshoot.subtitle')}
                  </DialogDescription>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {renderStatusBadge()}
              </div>
            </div>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto py-4 space-y-4 pr-1">
            {error && (
              <div className="p-3 rounded-md bg-sev-critical/10 border border-sev-critical/30 text-xs text-sev-critical flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{error}</span>
              </div>
            )}

            {actionNotice && (
              <div className="p-3 rounded-md bg-ok/10 border border-ok/30 text-xs text-ok flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 shrink-0" />
                <span>{actionNotice}</span>
              </div>
            )}

            {/* Diagnostic Verdict & Findings Card */}
            {diagnostics && (
              <Card className="p-4 border-separator-subtle bg-surface/80 space-y-3">
                <div className="flex items-start gap-2.5">
                  {diagnostics.likelyCause === 'ready' ? (
                    <CheckCircle2 className="w-5 h-5 text-ok shrink-0 mt-0.5" />
                  ) : diagnostics.likelyCause === 'unauthorized' ? (
                    <AlertTriangle className="w-5 h-5 text-sev-high shrink-0 mt-0.5" />
                  ) : diagnostics.likelyCause === 'usb_detected_no_adb' ? (
                    <AlertTriangle className="w-5 h-5 text-sev-medium shrink-0 mt-0.5" />
                  ) : (
                    <AlertCircle className="w-5 h-5 text-text-secondary shrink-0 mt-0.5" />
                  )}
                  <div className="space-y-1">
                    <h3 className="text-sm font-semibold text-text-primary">
                      {diagnostics.headline}
                    </h3>
                  </div>
                </div>

                {diagnostics.possibleCauses.length > 0 && (
                  <div className="space-y-1 pl-7">
                    <span className="text-[11px] font-semibold text-text-secondary uppercase tracking-wider">
                      {t('device.troubleshoot.causesTitle')}
                    </span>
                    <ul className="list-disc list-inside space-y-0.5 text-xs text-text-primary">
                      {diagnostics.possibleCauses.map((cause, i) => (
                        <li key={i}>{cause}</li>
                      ))}
                    </ul>
                  </div>
                )}

                {diagnostics.recommendedSteps.length > 0 && (
                  <div className="space-y-1.5 pl-7 pt-1">
                    <span className="text-[11px] font-semibold text-accent uppercase tracking-wider">
                      {t('device.troubleshoot.stepsTitle')}
                    </span>
                    <ol className="list-decimal list-inside space-y-1 text-xs text-text-primary">
                      {diagnostics.recommendedSteps.map((step, i) => (
                        <li key={i} className="font-medium leading-relaxed">{step}</li>
                      ))}
                    </ol>
                  </div>
                )}
              </Card>
            )}

            {/* Multi-tier Inspection Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
              {/* Layer 1: OS USB Hardware */}
              <div className="p-3 rounded-md bg-raised/50 border border-separator-subtle space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5 font-semibold text-text-primary">
                    <Usb className="w-4 h-4 text-accent" />
                    <span>{t('device.troubleshoot.hardwareSection')}</span>
                  </div>
                  <Badge severity={diagnostics?.usbHardwarePresent ? 'ok' : 'info'}>
                    {diagnostics?.usbHardwarePresent ? 'Detected' : 'Not Found'}
                  </Badge>
                </div>

                {diagnostics && diagnostics.usbDevices.length > 0 ? (
                  <div className="space-y-1.5 pt-1">
                    {diagnostics.usbDevices.map((dev, idx) => (
                      <div
                        key={idx}
                        className="p-2 rounded bg-surface/80 border border-separator-muted text-[11px] space-y-0.5"
                      >
                        <div className="font-semibold text-text-primary truncate">{dev.name}</div>
                        <div className="text-text-secondary flex items-center justify-between">
                          <span>{dev.manufacturer || 'Generic / Android'}</span>
                          <span className="font-mono text-[10px]">PnP: {dev.pnpClass}</span>
                        </div>
                        {dev.configManagerErrorCode !== 0 && (
                          <div className="text-sev-high font-mono text-[10px]">
                            Warning: ErrorCode {dev.configManagerErrorCode} (Driver Missing/Code 28)
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-[11px] text-text-secondary py-1">
                    {t('device.troubleshoot.noUsbHardware')}
                  </div>
                )}
              </div>

              {/* Layer 2: ADB Daemon State */}
              <div className="p-3 rounded-md bg-raised/50 border border-separator-subtle space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5 font-semibold text-text-primary">
                    <Server className="w-4 h-4 text-accent" />
                    <span>{t('device.troubleshoot.adbSection')}</span>
                  </div>
                  <Badge severity={diagnostics?.adbServerRunning ? 'ok' : 'critical'}>
                    {diagnostics?.adbServerRunning ? 'Running' : 'Stopped'}
                  </Badge>
                </div>

                <div className="text-[11px] text-text-secondary space-y-1 pt-1">
                  <div>
                    ADB Version:{' '}
                    <span className="font-mono text-text-primary">
                      {diagnostics?.adbVersion || 'Unknown'}
                    </span>
                  </div>
                  <div>
                    Attached Devices:{' '}
                    <span className="font-semibold text-text-primary">
                      {diagnostics?.adbDevices.length ?? 0}
                    </span>
                  </div>
                  {diagnostics && diagnostics.adbDevices.length > 0 && (
                    <div className="space-y-1 pt-1">
                      {diagnostics.adbDevices.map((dev, idx) => (
                        <div
                          key={idx}
                          className="p-1.5 rounded bg-surface/80 border border-separator-muted flex items-center justify-between"
                        >
                          <span className="font-mono text-[11px]">{dev.serial}</span>
                          <span className="uppercase text-[10px] font-semibold text-accent">
                            {dev.state}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Official OEM USB Drivers Directory */}
            <div className="p-4 rounded-md bg-raised/40 border border-separator-subtle space-y-3">
              <div>
                <h4 className="text-xs font-semibold text-text-primary">
                  {t('device.troubleshoot.driversTitle')}
                </h4>
                <p className="text-[11px] text-text-secondary leading-relaxed">
                  {t('device.troubleshoot.driversSubtitle')}
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {OFFICIAL_OEM_DRIVERS.map((item, idx) => (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => openExternal(item.url)}
                    className="p-2.5 rounded bg-surface hover:bg-raised/80 border border-separator-subtle transition-colors text-left flex items-center justify-between group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent"
                  >
                    <div className="min-w-0 pr-2">
                      <div className="text-xs font-medium text-text-primary group-hover:text-accent flex items-center gap-1.5">
                        <span>{item.brand}</span>
                      </div>
                      <div className="text-[10px] text-text-secondary truncate">{item.note}</div>
                    </div>
                    <ExternalLink className="w-3.5 h-3.5 text-text-secondary group-hover:text-accent shrink-0" />
                  </button>
                ))}
              </div>

              <div className="text-[10px] text-text-secondary/80 flex items-center gap-1 pt-1">
                <span>{t('device.troubleshoot.driversOfficialNotice')}</span>
              </div>
            </div>
          </div>

          <DialogFooter className="shrink-0 pt-3 border-t border-separator-subtle flex flex-wrap items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleSoftReconnect}
                disabled={isSoftReconnecting || isLoading}
              >
                <RotateCcw className={`w-3.5 h-3.5 ${isSoftReconnecting ? 'animate-spin' : ''}`} />
                <span>
                  {isSoftReconnecting
                    ? t('device.troubleshoot.softReconnecting')
                    : t('device.troubleshoot.softReconnect')}
                </span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowRestartConfirm(true)}
                disabled={isRestartingServer || isLoading}
                className="text-sev-high hover:text-sev-high"
              >
                <ShieldAlert className="w-3.5 h-3.5" />
                <span>{t('device.troubleshoot.restartServer')}</span>
              </Button>

              {onOpenWireless && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    onOpenChange(false);
                    onOpenWireless();
                  }}
                >
                  <Wifi className="w-3.5 h-3.5" />
                  <span>{t('device.wireless.button')}</span>
                </Button>
              )}
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="primary"
                size="sm"
                onClick={runDiagnostics}
                disabled={isLoading}
              >
                <RotateCcw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
                <span>
                  {isLoading
                    ? t('device.troubleshoot.runningCheck')
                    : t('device.troubleshoot.runCheck')}
                </span>
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirmation Dialog for Global ADB Kill-Server */}
      <Dialog open={showRestartConfirm} onOpenChange={setShowRestartConfirm}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-2 text-sev-high">
              <AlertTriangle className="w-5 h-5 shrink-0" />
              <DialogTitle>{t('device.troubleshoot.restartConfirmTitle')}</DialogTitle>
            </div>
            <DialogDescription className="pt-2 text-xs leading-relaxed text-text-primary font-medium">
              {t('device.troubleshoot.restartConfirmDesc')}
            </DialogDescription>
          </DialogHeader>

          <DialogFooter className="pt-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowRestartConfirm(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={handleRestartServerConfirmed}
              disabled={isRestartingServer}
              className="bg-sev-high hover:bg-sev-high/90 text-white"
            >
              {isRestartingServer ? t('common.loading') : t('device.troubleshoot.restartServer')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};
