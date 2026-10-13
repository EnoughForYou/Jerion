import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Smartphone,
  Shield,
  Search,
  CheckCircle2,
  Calendar,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { ErrorState } from '../../components/ui/ErrorState';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import { StaticFallback } from '../../three/StaticFallback';
import type { ScanResult } from '../../lib/bridge/schema';

const DeviceScene = React.lazy(() => import('../../three/DeviceScene'));

export const HomeView: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeDevice, setActiveDevice } = useDeviceStore();

  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadData() {
      try {
        setIsLoading(true);
        setError(null);

        const device = await bridge.device.getConnected();
        if (isMounted) {
          setActiveDevice(device);
        }

        const scanList = await bridge.report.list();
        if (scanList.length > 0 && isMounted) {
          const latest = await bridge.report.get(scanList[0].id);
          setScanResult(latest);
        }
      } catch (err: unknown) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Failed to load device overview');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadData();

    return () => {
      isMounted = false;
    };
  }, [bridge, setActiveDevice]);

  if (error) {
    return (
      <ErrorState
        title="Failed to Load Device Overview"
        remedy="Check ADB connection and verify device authorization."
        onRetry={() => window.location.reload()}
      />
    );
  }

  const device = activeDevice;

  return (
    <div className="space-y-5">
      {/* Top Banner & Quick Scan Action */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 p-5 rounded-lg bg-surface border border-separator/35 shadow-sm">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <Shield className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('home.title')}
            </h2>
            <p className="text-xs text-text-secondary leading-relaxed">
              {t('home.scanCtaSubtitle')}
            </p>
          </div>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => navigate('/scan')}
        >
          <Search className="w-4 h-4" />
          {t('home.scanCta')}
        </Button>
      </div>

      {/* Grid: Device Specs & Security Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* Hardware & OS Card (3uTools Dense Grid) */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Smartphone className="w-4 h-4 text-accent" />
              <CardTitle>{t('home.deviceInfo')}</CardTitle>
            </div>
            <CardDescription>
              Hardware attributes and firmware specifications
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading || !device ? (
              <div className="space-y-2.5 py-1">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-4/6" />
                <Skeleton className="h-4 w-full" />
              </div>
            ) : (
              <dl className="grid grid-cols-2 gap-x-4 gap-y-2.5 text-xs">
                <div>
                  <dt className="text-text-secondary font-medium">Model</dt>
                  <dd className="text-text-primary font-semibold">{device.model}</dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">Manufacturer</dt>
                  <dd className="text-text-primary">{device.manufacturer}</dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">Android Version</dt>
                  <dd className="text-text-primary">
                    Android {device.androidVersion} (API {device.sdkInt})
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">Serial</dt>
                  <dd className="text-text-primary font-mono text-[11px]">{device.serial}</dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">Battery</dt>
                  <dd className="text-text-primary">
                    {device.batteryLevel}% {device.isCharging ? '(Charging)' : ''}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">Device Type</dt>
                  <dd className="text-text-primary">
                    {device.isEmulator ? 'Emulator (AVD)' : 'Physical Device'}
                  </dd>
                </div>
              </dl>
            )}
          </CardContent>
        </Card>

        {/* Security Findings Summary Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-accent" />
              <CardTitle>{t('home.securitySummary')}</CardTitle>
            </div>
            <CardDescription>
              Threat detections categorized by severity
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-2.5 py-1">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-3/4" />
              </div>
            ) : scanResult ? (
              <div className="space-y-3.5">
                <div className="flex flex-wrap gap-2">
                  <Badge severity="critical" count={scanResult.summary.counts.critical} />
                  <Badge severity="high" count={scanResult.summary.counts.high} />
                  <Badge severity="medium" count={scanResult.summary.counts.medium} />
                  <Badge severity="low" count={scanResult.summary.counts.low} />
                  <Badge severity="info" count={scanResult.summary.counts.info} />
                </div>

                <div className="pt-2 border-t border-separator/20 grid grid-cols-2 gap-2 text-xs text-text-secondary">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="w-3.5 h-3.5" />
                    <span>Mode: {scanResult.mode.toUpperCase()}</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-ok" />
                    <span>Findings: {scanResult.findings.length} detected</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="py-4 text-center text-xs text-text-secondary">
                No previous scan recorded. Run a scan to evaluate device threats.
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 3D Scene / Visual Area Container */}
      <Card className="min-h-[220px] p-0 overflow-hidden flex items-center justify-center bg-surface border border-separator/35 shadow-sm">
        <React.Suspense fallback={<StaticFallback findingsCount={scanResult?.findings.length || 0} />}>
          <DeviceScene
            status={scanResult ? 'completed' : 'idle'}
            findingsCount={scanResult?.findings.length || 0}
          />
        </React.Suspense>
      </Card>
    </div>
  );
};
