import React, { useState, useEffect, useMemo } from 'react';
import {
  Search,
  Zap,
  Layers,
  StopCircle,
  CheckCircle2,
  AlertTriangle,
  Shield,
  Filter,
  Smartphone,
  Clock,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/Dialog';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import { useSystemInfoStore } from '../../app/systemInfoStore';
import { getCanonicalSecurityState } from '../../app/deviceSecurityState';
import { ScanTimeline } from './ScanTimeline';
import { EvidenceDrawer } from './EvidenceDrawer';
import type { Finding, ScanMode, ScanProgress, ScanResult, Severity } from '../../lib/bridge/schema';

export const ScanView: React.FC = () => {
  const { t } = useTranslation();
  const { activeDevice, setActiveDevice, setActiveJob } = useDeviceStore();
  const { systemInfo, fetchSystemInfo } = useSystemInfoStore();

  useEffect(() => {
    fetchSystemInfo();
  }, [fetchSystemInfo]);

  const [mode, setMode] = useState<ScanMode>('quick');
  const [includeSystem, setIncludeSystem] = useState(false);
  const [activeScanId, setActiveScanId] = useState<string | null>(null);
  const [progress, setProgress] = useState<ScanProgress | null>(null);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [findings, setFindings] = useState<Finding[]>([]);
  const [isScanning, setIsScanning] = useState(false);
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [selectedFinding, setSelectedFinding] = useState<Finding | null>(null);
  const [selectedSeverity, setSelectedSeverity] = useState<Severity | 'all'>('all');

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    const unsubProgress = bridge.on('scan:progress', (data) => {
      setProgress(data);
      setActiveJob(data.step, data.progress);
    });

    const unsubFinding = bridge.on('scan:finding', (finding) => {
      setFindings((prev) => [finding, ...prev]);
    });

    const unsubDone = bridge.on('scan:done', (result) => {
      setScanResult(result);
      setFindings(result?.findings || []);
      setIsScanning(false);
      setActiveScanId(null);
      setActiveJob(null);
    });

    async function loadInitial() {
      try {
        let device = activeDevice;
        if (!device) {
          try {
            device = await bridge.device.getConnected();
            if (device && isMounted) {
              setActiveDevice(device);
            }
          } catch {
            device = null;
          }
        }

        if (device && device.connectionState === 'ready') {
          const list = await bridge.report.list();
          const devId = device.serial;
          const altId = device.id;
          if (list && list.length > 0) {
            const res = await bridge.report.get(list[0].id);
            if (
              res &&
              ((devId && res.deviceId === devId) || (altId && res.deviceId === altId))
            ) {
              setScanResult(res);
              setFindings(res.findings || []);
            } else {
              setScanResult(null);
              setFindings([]);
            }
          } else {
            setScanResult(null);
            setFindings([]);
          }
        } else {
          setScanResult(null);
          setFindings([]);
        }
      } catch (err) {
        console.error('Failed to load initial scan:', err);
      }
    }

    loadInitial();

    return () => {
      isMounted = false;
      unsubProgress();
      unsubFinding();
      unsubDone();
    };
  }, [bridge, setActiveJob, activeDevice, setActiveDevice]);

  const handleStartScan = async () => {
    try {
      setIsScanning(true);
      setFindings([]);
      setScanResult(null);
      setProgress(null);

      const scanId = await bridge.scan.start({
        mode,
        includeSystem: mode === 'deep' ? true : includeSystem,
      });
      setActiveScanId(scanId);
      setActiveJob(`Scanner (${mode})`, 0);
    } catch (err) {
      console.error('Failed to start scan:', err);
      setIsScanning(false);
    }
  };

  const handleConfirmCancel = async () => {
    setShowCancelDialog(false);
    if (!activeScanId) return;
    try {
      await bridge.scan.cancel(activeScanId);
      setIsScanning(false);
      setActiveScanId(null);
      setActiveJob(null);
    } catch (err) {
      console.error('Failed to cancel scan:', err);
    }
  };

  const filteredFindings = useMemo(() => {
    if (selectedSeverity === 'all') return findings;
    return findings.filter((f) => f.severity === selectedSeverity);
  }, [findings, selectedSeverity]);

  const securityContext = getCanonicalSecurityState(activeDevice, scanResult, isScanning, systemInfo?.rulesVersion);

  return (
    <div className="space-y-5">
      {/* Scan Mode Selection Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Quick Scan Card */}
        <div
          onClick={() => !isScanning && setMode('quick')}
          className={`p-4 rounded-lg border transition-all cursor-pointer ${
            mode === 'quick'
              ? 'border-accent bg-accent/5 ring-1 ring-accent'
              : 'border-separator-subtle bg-surface hover:bg-raised/40'
          }`}
        >
          <div className="flex items-center gap-2.5 mb-2">
            <Zap className="w-4 h-4 text-accent" />
            <span className="font-semibold text-sm text-text-primary">
              {t('scan.quickScan')}
            </span>
          </div>
          <p className="text-xs text-text-secondary leading-relaxed">
            {t('scan.quickScanDesc')}
          </p>
        </div>

        {/* Deep Scan Card */}
        <div
          onClick={() => !isScanning && setMode('deep')}
          className={`p-4 rounded-lg border transition-all cursor-pointer ${
            mode === 'deep'
              ? 'border-accent bg-accent/5 ring-1 ring-accent'
              : 'border-separator-subtle bg-surface hover:bg-raised/40'
          }`}
        >
          <div className="flex items-center gap-2.5 mb-2">
            <Layers className="w-4 h-4 text-accent" />
            <span className="font-semibold text-sm text-text-primary">
              {t('scan.deepScan')}
            </span>
          </div>
          <p className="text-xs text-text-secondary leading-relaxed">
            {t('scan.deepScanDesc')}
          </p>
        </div>
      </div>

      {/* Options & Action Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 bg-surface border border-separator-subtle rounded-lg shadow-sm">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Shield className="w-4 h-4 text-accent" />
            <span className="font-semibold text-xs text-text-primary">
              {isScanning
                ? t('scan.running')
                : t('scan.ready', { mode: mode.toLowerCase() })}
            </span>
          </div>
          <p className="text-[11px] text-text-secondary">
            {t('scan.targetDevice', {
              device: activeDevice?.model || 'No device connected',
            })}
          </p>
        </div>

        <div className="flex items-center gap-4">
          {mode === 'quick' && !isScanning && (
            <label className="flex items-center gap-2 text-xs text-text-secondary cursor-pointer select-none">
              <input
                type="checkbox"
                checked={includeSystem}
                onChange={(e) => setIncludeSystem(e.target.checked)}
                className="rounded border-separator text-accent focus:ring-accent"
              />
              <span>{t('scan.includeSystem')}</span>
            </label>
          )}

          {isScanning ? (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setShowCancelDialog(true)}
            >
              <StopCircle className="w-3.5 h-3.5" />
              {t('scan.cancelScan')}
            </Button>
          ) : (
            <div className="flex flex-col items-end gap-1">
              <Button
                variant="primary"
                size="md"
                onClick={handleStartScan}
                disabled={!securityContext.canStartScan}
                aria-disabled={!securityContext.canStartScan}
              >
                <Search className="w-3.5 h-3.5" />
                {t('scan.startScan')}
              </Button>
              {!securityContext.canStartScan && securityContext.disabledReason && (
                <span className="text-[10px] text-text-secondary">
                  {securityContext.disabledReason}
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Progress & Live Timeline */}
      {(isScanning || (progress && progress.progress > 0 && progress.progress < 100)) && (
        <ScanTimeline progress={progress} isScanning={isScanning} />
      )}

      {/* Findings Results */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('scan.results')}</CardTitle>
              <CardDescription>
                {t('scan.findingsCount', { count: filteredFindings?.length ?? 0 })}
                {scanResult?.completedAt &&
                  scanResult.status === 'completed' &&
                  scanResult.summary?.verdict &&
                  ` · Verdict: ${scanResult.summary.verdict.toUpperCase()}`}
                {scanResult?.completedAt &&
                  scanResult.status === 'partial' &&
                  scanResult.summary?.verdict &&
                  ` · Verdict: ${scanResult.summary.verdict.toUpperCase()} (PARTIAL)`}
              </CardDescription>
            </div>

            {/* Filter pills - only displayed when findings exist */}
            {(findings || []).length > 0 && (
              <div className="flex items-center gap-1.5 flex-wrap">
                <Filter className="w-3.5 h-3.5 text-text-secondary mr-1" />
                <button
                  type="button"
                  onClick={() => setSelectedSeverity('all')}
                  className={`px-2 py-0.5 rounded text-[11px] font-medium border transition-colors ${
                    selectedSeverity === 'all'
                      ? 'bg-accent text-accent-foreground border-accent'
                      : 'bg-raised text-text-secondary border-separator-subtle hover:text-text-primary'
                  }`}
                >
                  {t('scan.allSeverities')} ({(findings || []).length})
                </button>

                {(['critical', 'high', 'medium', 'low', 'info'] as Severity[]).map((sev) => {
                  const count = (findings || []).filter((f) => f.severity === sev).length;

                  return (
                    <button
                      key={sev}
                      type="button"
                      onClick={() => setSelectedSeverity(sev)}
                      className={`px-2 py-0.5 rounded text-[11px] font-medium border transition-colors ${
                        selectedSeverity === sev
                          ? 'bg-accent text-accent-foreground border-accent'
                          : 'bg-raised text-text-secondary border-separator-subtle hover:text-text-primary'
                      }`}
                    >
                      {sev.toUpperCase()} ({count})
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {(filteredFindings?.length ?? 0) === 0 ? (
            securityContext.state === 'no-device' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <Smartphone className="w-8 h-8 text-text-secondary mx-auto mb-2 opacity-50" />
                <p className="font-medium text-text-primary">
                  {t('device.noDevice')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('device.noDeviceHint')}
                </p>
              </div>
            ) : securityContext.state === 'unauthorized' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <AlertTriangle className="w-8 h-8 text-sev-high mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {t('device.unauthorizedTitle', 'Authorization Required')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('device.unauthorizedPrompt', "Unlock your phone screen and tap 'Allow USB debugging'.")}
                </p>
              </div>
            ) : securityContext.state === 'offline' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <AlertTriangle className="w-8 h-8 text-sev-medium mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {t('device.offlineTitle', 'Device Offline')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('device.offlinePrompt', 'Device connection is offline. Reconnect USB.')}
                </p>
              </div>
            ) : securityContext.state === 'never-scanned' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <Clock className="w-8 h-8 text-text-secondary mx-auto mb-2 opacity-60" />
                <p className="font-medium text-text-primary">
                  {t('scan.noScanYetTitle', 'No Scan Performed Yet')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.noScanYetDesc', 'Select a scan mode above and start scanning to inspect this device.')}
                </p>
              </div>
            ) : scanResult && (scanResult.status === 'failed' || scanResult.status === 'cancelled') ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <AlertTriangle className="w-8 h-8 text-warning mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {scanResult.status === 'cancelled' ? t('scan.scanCancelled') : t('scan.scanFailed')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.incompleteNotice')}
                </p>
              </div>
            ) : scanResult && scanResult.status === 'partial' && (findings?.length ?? 0) === 0 ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <AlertTriangle className="w-8 h-8 text-warning mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {t('scan.scanPartial')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.partialNoFindingsNotice')}
                </p>
              </div>
            ) : selectedSeverity !== 'all' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <Filter className="w-8 h-8 text-text-secondary mx-auto mb-2 opacity-60" />
                <p className="font-medium text-text-primary">
                  {t('scan.noFindingsMatchFilter')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.noFindingsMatchFilterDesc')}
                </p>
              </div>
            ) : securityContext.state === 'scanned-clean' ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <CheckCircle2 className="w-8 h-8 text-ok mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {t('scan.noFindingsTitle', 'No findings in the last scan')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto font-mono">
                  {securityContext.contextText}
                </p>
              </div>
            ) : (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <Clock className="w-8 h-8 text-text-secondary mx-auto mb-2 opacity-60" />
                <p className="font-medium text-text-primary">
                  {t('scan.noScanYetTitle', 'No Scan Performed Yet')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.noScanYetDesc')}
                </p>
              </div>
            )
          ) : (
            <div className="space-y-3">
              {filteredFindings.map((finding) => (
                <div
                  key={finding.id}
                  className="p-3.5 rounded-md bg-raised/30 border border-separator-subtle space-y-2 text-xs transition-colors hover:border-separator"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <Badge severity={finding.severity} />
                        <span className="font-semibold text-text-primary">
                          {finding.ruleId}: {finding.title}
                        </span>
                      </div>
                      {finding.package && (
                        <span className="font-mono text-[11px] text-text-secondary block">
                          {t('scan.packageLabel', { package: finding.package })}
                        </span>
                      )}
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      <span className="text-[10px] text-text-secondary font-mono">
                        {t('scan.confidenceLabel', {
                          confidence: finding.confidence.toUpperCase(),
                        })}
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedFinding(finding)}
                      >
                        {t('scan.viewEvidence')}
                      </Button>
                    </div>
                  </div>

                  <p className="text-text-secondary text-[11px] leading-relaxed">
                    {finding.whyFlagged}
                  </p>

                  {finding.recommendedAction && (
                    <div className="pt-2 border-t border-separator-muted text-[11px] text-accent flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 truncate">
                        <AlertTriangle className="w-3.5 h-3.5 text-sev-medium shrink-0" />
                        <span className="truncate">
                          Remedy: {finding.recommendedAction.title}
                        </span>
                      </div>
                      <span className="text-[10px] text-text-secondary shrink-0 font-mono">
                        {finding.recommendedAction.impact}
                      </span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Confirmation Dialog on Cancel */}
      <Dialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('scan.confirmCancelTitle')}</DialogTitle>
            <DialogDescription>{t('scan.confirmCancelDesc')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowCancelDialog(false)}
            >
              {t('scan.continueScanning')}
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleConfirmCancel}
            >
              {t('scan.confirmCancelAction')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Evidence Inspector Drawer */}
      <EvidenceDrawer
        finding={selectedFinding}
        open={Boolean(selectedFinding)}
        onOpenChange={(open) => !open && setSelectedFinding(null)}
      />
    </div>
  );
};
