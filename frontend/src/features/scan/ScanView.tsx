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
import { ScanTimeline } from './ScanTimeline';
import { EvidenceDrawer } from './EvidenceDrawer';
import type { Finding, ScanMode, ScanProgress, ScanResult, Severity } from '../../lib/bridge/schema';

export const ScanView: React.FC = () => {
  const { t } = useTranslation();
  const { activeDevice, setActiveJob } = useDeviceStore();

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
    const unsubProgress = bridge.on('scan:progress', (data) => {
      setProgress(data);
      setActiveJob(data.step, data.progress);
    });

    const unsubFinding = bridge.on('scan:finding', (finding) => {
      setFindings((prev) => [finding, ...prev]);
    });

    const unsubDone = bridge.on('scan:done', (result) => {
      setScanResult(result);
      setFindings(result.findings);
      setIsScanning(false);
      setActiveScanId(null);
      setActiveJob(null);
    });

    async function loadInitial() {
      try {
        const list = await bridge.report.list();
        if (list.length > 0) {
          const res = await bridge.report.get(list[0].id);
          if (res) {
            setScanResult(res);
            setFindings(res.findings);
          }
        }
      } catch (err) {
        console.error('Failed to load initial scan:', err);
      }
    }

    loadInitial();

    return () => {
      unsubProgress();
      unsubFinding();
      unsubDone();
    };
  }, [bridge, setActiveJob]);

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

  return (
    <div className="space-y-5">
      {/* Scan Mode Selection Header */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Quick Scan Card */}
        <div
          onClick={() => !isScanning && setMode('quick')}
          className={`p-4 rounded-lg border transition-all cursor-pointer ${
            mode === 'quick'
              ? 'border-accent bg-accent/5 ring-1 ring-accent'
              : 'border-separator/35 bg-surface hover:bg-raised/40'
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
              : 'border-separator/35 bg-surface hover:bg-raised/40'
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
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 bg-surface border border-separator/35 rounded-lg shadow-sm">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Shield className="w-4 h-4 text-accent" />
            <span className="font-semibold text-xs text-text-primary">
              {isScanning
                ? t('scan.running')
                : t('scan.ready', { mode: mode.toUpperCase() })}
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
            <Button
              variant="primary"
              size="md"
              onClick={handleStartScan}
              disabled={activeDevice?.connectionState !== 'ready'}
            >
              <Search className="w-3.5 h-3.5" />
              {t('scan.startScan')}
            </Button>
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
                {t('scan.findingsCount', { count: filteredFindings.length })}
                {scanResult?.completedAt &&
                  (scanResult.status === 'completed' || scanResult.status === 'partial') &&
                  ` · Verdict: ${scanResult.summary.verdict.toUpperCase()}`}
              </CardDescription>
            </div>

            {/* Filter pills */}
            <div className="flex items-center gap-1.5 flex-wrap">
              <Filter className="w-3.5 h-3.5 text-text-secondary mr-1" />
              <button
                type="button"
                onClick={() => setSelectedSeverity('all')}
                className={`px-2 py-0.5 rounded text-[11px] font-medium border transition-colors ${
                  selectedSeverity === 'all'
                    ? 'bg-accent text-accent-foreground border-accent'
                    : 'bg-raised text-text-secondary border-separator/30 hover:text-text-primary'
                }`}
              >
                {t('scan.allSeverities')} ({findings.length})
              </button>

              {(['critical', 'high', 'medium', 'low', 'info'] as Severity[]).map((sev) => {
                const count = findings.filter((f) => f.severity === sev).length;
                if (count === 0 && selectedSeverity !== sev) return null;

                return (
                  <button
                    key={sev}
                    type="button"
                    onClick={() => setSelectedSeverity(sev)}
                    className={`px-2 py-0.5 rounded text-[11px] font-medium border transition-colors ${
                      selectedSeverity === sev
                        ? 'bg-accent text-accent-foreground border-accent'
                        : 'bg-raised text-text-secondary border-separator/30 hover:text-text-primary'
                    }`}
                  >
                    {sev.toUpperCase()} ({count})
                  </button>
                );
              })}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredFindings.length === 0 ? (
            scanResult && (scanResult.status === 'failed' || scanResult.status === 'cancelled') ? (
              <div className="py-8 text-center text-xs text-text-secondary space-y-2">
                <AlertTriangle className="w-8 h-8 text-warning mx-auto mb-2 opacity-80" />
                <p className="font-medium text-text-primary">
                  {scanResult.status === 'cancelled' ? t('scan.scanCancelled') : t('scan.scanFailed')}
                </p>
                <p className="text-[11px] text-text-secondary max-w-sm mx-auto">
                  {t('scan.incompleteNotice')}
                </p>
              </div>
            ) : (
              <div className="py-8 text-center text-xs text-text-secondary">
                <CheckCircle2 className="w-8 h-8 text-ok mx-auto mb-2 opacity-80" />
                {t('scan.noFindings')}
              </div>
            )
          ) : (
            <div className="space-y-3">
              {filteredFindings.map((finding) => (
                <div
                  key={finding.id}
                  className="p-3.5 rounded-md bg-raised/30 border border-separator/25 space-y-2 text-xs transition-colors hover:border-separator/40"
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
                    <div className="pt-2 border-t border-separator/15 text-[11px] text-accent flex items-center justify-between gap-2">
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
