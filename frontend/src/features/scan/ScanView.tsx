import React, { useState, useEffect } from 'react';
import {
  Search,
  Zap,
  Layers,
  StopCircle,
  CheckCircle2,
  AlertTriangle,
  Shield,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import type { Finding, ScanMode, ScanProgress, ScanResult } from '../../lib/bridge/schema';

export const ScanView: React.FC = () => {
  const { t } = useTranslation();
  const { activeDevice, setActiveJob } = useDeviceStore();

  const [mode, setMode] = useState<ScanMode>('quick');
  const [activeScanId, setActiveScanId] = useState<string | null>(null);
  const [progress, setProgress] = useState<ScanProgress | null>(null);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [findings, setFindings] = useState<Finding[]>([]);
  const [isScanning, setIsScanning] = useState(false);

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

      const scanId = await bridge.scan.start({ mode, includeSystem: mode === 'deep' });
      setActiveScanId(scanId);
      setActiveJob(`Scanner (${mode})`, 0);
    } catch (err) {
      console.error('Failed to start scan:', err);
      setIsScanning(false);
    }
  };

  const handleCancelScan = async () => {
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

      {/* Action Bar */}
      <div className="flex items-center justify-between p-4 bg-surface border border-separator/35 rounded-lg">
        <div className="flex items-center gap-3">
          <Shield className="w-5 h-5 text-accent" />
          <div>
            <span className="font-medium text-xs text-text-primary block">
              {isScanning ? t('scan.running') : `Ready to run ${mode.toUpperCase()} scan`}
            </span>
            <span className="text-[11px] text-text-secondary">
              Target: {activeDevice?.model || 'Infinix Note 11S (mock)'}
            </span>
          </div>
        </div>

        {isScanning ? (
          <Button variant="destructive" size="sm" onClick={handleCancelScan}>
            <StopCircle className="w-3.5 h-3.5" />
            Cancel Scan
          </Button>
        ) : (
          <Button variant="primary" size="md" onClick={handleStartScan}>
            <Search className="w-3.5 h-3.5" />
            Start Scan
          </Button>
        )}
      </div>

      {/* Progress Bar (during scanning) */}
      {isScanning && progress ? (
        <Card className="space-y-3">
          <div className="flex justify-between text-xs text-text-secondary">
            <span>Phase: {progress.step}</span>
            <span>{Math.round(progress.progress)}%</span>
          </div>
          <div className="h-2 w-full bg-raised rounded-full overflow-hidden">
            <div
              className="h-full bg-accent transition-all duration-200"
              style={{ width: `${progress.progress}%` }}
            />
          </div>
          {progress.currentPackage ? (
            <p className="text-[11px] font-mono text-text-secondary truncate">
              Inspecting: {progress.currentPackage}
            </p>
          ) : null}
        </Card>
      ) : null}

      {/* Findings Results */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('scan.results')}</CardTitle>
              <CardDescription>
                {findings.length} findings identified across inspected packages
              </CardDescription>
            </div>
            {scanResult ? (
              <div className="flex items-center gap-1.5">
                <Badge severity="critical" count={scanResult.summary.counts.critical} />
                <Badge severity="high" count={scanResult.summary.counts.high} />
                <Badge severity="medium" count={scanResult.summary.counts.medium} />
              </div>
            ) : null}
          </div>
        </CardHeader>
        <CardContent>
          {findings.length === 0 ? (
            <div className="py-8 text-center text-xs text-text-secondary">
              <CheckCircle2 className="w-8 h-8 text-ok mx-auto mb-2 opacity-80" />
              {t('scan.noFindings')}
            </div>
          ) : (
            <div className="space-y-3">
              {findings.map((finding) => (
                <div
                  key={finding.id}
                  className="p-3 rounded-md bg-raised/30 border border-separator/25 space-y-2 text-xs"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <Badge severity={finding.severity} />
                        <span className="font-semibold text-text-primary">
                          {finding.ruleId}: {finding.title}
                        </span>
                      </div>
                      {finding.package ? (
                        <span className="font-mono text-[11px] text-text-secondary block">
                          Package: {finding.package}
                        </span>
                      ) : null}
                    </div>
                    <span className="text-[10px] text-text-secondary font-mono">
                      Confidence: {finding.confidence.toUpperCase()}
                    </span>
                  </div>

                  <p className="text-text-secondary text-[11px] leading-relaxed">
                    {finding.whyFlagged}
                  </p>

                  {finding.recommendedAction ? (
                    <div className="pt-2 border-t border-separator/15 text-[11px] text-accent flex items-center gap-1">
                      <AlertTriangle className="w-3 h-3 text-sev-medium shrink-0" />
                      <span>Remedy: {finding.recommendedAction.title} ({finding.recommendedAction.impact})</span>
                    </div>
                  ) : null}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
