import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Smartphone,
  Shield,
  Search,
  CheckCircle2,
  Calendar,
  HelpCircle,
  AlertTriangle,
  RotateCcw,
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
import { EmptyState } from '../../components/ui/EmptyState';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import { StaticFallback } from '../../three/StaticFallback';
import { OnboardingModal } from './OnboardingModal';
import type { ScanResult } from '../../lib/bridge/schema';

const DeviceScene = React.lazy(() => import('../../three/DeviceScene'));

export const HomeView: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeDevice, setActiveDevice } = useDeviceStore();

  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showOnboarding, setShowOnboarding] = useState(false);

  const bridge = getBridge();

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const device = await bridge.device.getConnected();
      setActiveDevice(device);

      const scanList = await bridge.report.list();
      if (scanList && scanList.length > 0) {
        const latest = await bridge.report.get(scanList[0].id);
        setScanResult(latest);
      }
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'Failed to load device overview'
      );
    } finally {
      setIsLoading(false);
    }
  }, [bridge, setActiveDevice]);

  useEffect(() => {
    let isMounted = true;

    async function init() {
      await loadData();
      if (!isMounted) return;
    }

    init();

    // Subscribe to device changes for live updates
    const unsub = bridge.on('device:changed', (device) => {
      if (isMounted) {
        setActiveDevice(device);
      }
    });

    return () => {
      isMounted = false;
      unsub();
    };
  }, [bridge, loadData, setActiveDevice]);

  if (error) {
    return (
      <ErrorState
        title={t('common.error', 'Failed to Load Device Overview')}
        remedy={t(
          'common.errorRemedy',
          'Check ADB connection and verify device authorization.'
        )}
        onRetry={loadData}
      />
    );
  }

  const device = activeDevice;

  // Determine device connection state for conditional rendering
  const isDeviceReady = device?.connectionState === 'ready';
  const isDeviceUnauthorized = device?.connectionState === 'unauthorized';

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

        <div className="flex items-center gap-2">
          {!device && (
            <Button
              variant="secondary"
              size="md"
              onClick={() => setShowOnboarding(true)}
            >
              <HelpCircle className="w-4 h-4" />
              {t('onboarding.title')}
            </Button>
          )}
          <Button
            variant="primary"
            size="md"
            onClick={() => navigate('/scan')}
            disabled={!isDeviceReady}
          >
            <Search className="w-4 h-4" />
            {t('home.scanCta')}
          </Button>
        </div>
      </div>

      {/* Prominent Actionable Notice when Device is Connected but Unauthorized */}
      {isDeviceUnauthorized && (
        <div className="p-4 rounded-lg bg-sev-high/10 border border-sev-high/30 text-xs text-text-primary space-y-2.5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sev-high font-semibold text-sm">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              <span>{t('device.unauthorizedTitle', 'Authorization Required (RSA Key)')}</span>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={loadData}
                disabled={isLoading}
              >
                <RotateCcw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
                <span>{t('device.checkAgain', 'Check Device Again')}</span>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowOnboarding(true)}
              >
                {t('onboarding.title')}
              </Button>
            </div>
          </div>
          <p className="text-xs text-text-primary font-medium leading-relaxed">
            {t('device.unauthorizedPrompt', "Unlock your phone screen and look for the 'Allow USB debugging?' prompt.")}{' '}
            <span className="text-sev-high font-semibold">{t('device.unauthorizedAction', "Check 'Always allow from this computer' and tap 'Allow'.")}</span>
          </p>
          <div className="flex items-start gap-2 p-2.5 rounded bg-surface/60 border border-separator/20 text-[11px] text-text-secondary leading-relaxed">
            <HelpCircle className="w-3.5 h-3.5 text-accent shrink-0 mt-0.5" />
            <span>{t('device.unauthorizedTip', "Prompt not showing? Swipe down your Android notification shade, switch USB mode from 'Charging only' to 'File Transfer (MTP)', or toggle USB Debugging in Developer Options.")}</span>
          </div>
        </div>
      )}

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
              {t('home.deviceInfoDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-2.5 py-1">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-4/6" />
                <Skeleton className="h-4 w-full" />
              </div>
            ) : device ? (
              <dl className="grid grid-cols-2 gap-x-4 gap-y-2.5 text-xs">
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.model')}
                  </dt>
                  <dd className="text-text-primary font-semibold">
                    {device.model || t('device.status.unauthorized')}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.manufacturer')}
                  </dt>
                  <dd className="text-text-primary">
                    {isDeviceReady
                      ? device.manufacturer
                      : t('device.pendingAuth', 'Pending Authorization')}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.androidVersion')}
                  </dt>
                  <dd className="text-text-primary">
                    {isDeviceReady
                      ? t('home.deviceLabels.androidVersionValue', {
                          version: device.androidVersion,
                          sdk: device.sdkInt,
                        })
                      : t('device.pendingAuth', 'Pending Authorization')}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.serial')}
                  </dt>
                  <dd className="text-text-primary font-mono text-[11px]">
                    {device.serial}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.battery')}
                  </dt>
                  <dd className="text-text-primary">
                    {isDeviceReady
                      ? device.isCharging
                        ? t('home.deviceLabels.batteryCharging', {
                            level: device.batteryLevel,
                          })
                        : t('home.deviceLabels.batteryValue', {
                            level: device.batteryLevel,
                          })
                      : device.isCharging
                      ? 'Charging (' + t('device.pendingAuth', 'Pending Authorization') + ')'
                      : t('device.pendingAuth', 'Pending Authorization')}
                  </dd>
                </div>
                <div>
                  <dt className="text-text-secondary font-medium">
                    {t('home.deviceLabels.deviceType')}
                  </dt>
                  <dd className="text-text-primary">
                    {device.isEmulator
                      ? t('home.deviceLabels.emulator')
                      : t('home.deviceLabels.physicalDevice')}
                  </dd>
                </div>
              </dl>
            ) : (
              <EmptyState
                icon={<Smartphone className="w-5 h-5" />}
                title={t('device.noDevice')}
                description={t('device.noDeviceHint')}
                action={
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setShowOnboarding(true)}
                  >
                    {t('onboarding.title')}
                  </Button>
                }
              />
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
              {t('home.securitySummaryDesc')}
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
                  <Badge severity="critical" count={scanResult.summary?.counts?.critical ?? 0} />
                  <Badge severity="high" count={scanResult.summary?.counts?.high ?? 0} />
                  <Badge severity="medium" count={scanResult.summary?.counts?.medium ?? 0} />
                  <Badge severity="low" count={scanResult.summary?.counts?.low ?? 0} />
                  <Badge severity="info" count={scanResult.summary?.counts?.info ?? 0} />
                </div>

                <div className="pt-2 border-t border-separator/20 grid grid-cols-2 gap-2 text-xs text-text-secondary">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="w-3.5 h-3.5" />
                    <span>
                      {t('home.scanMode', {
                        mode: (scanResult.mode || 'quick').toUpperCase(),
                      })}
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-ok" />
                    <span>
                      {t('home.findingsDetected', {
                        count: scanResult.findings?.length ?? 0,
                      })}
                    </span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="py-4 text-center text-xs text-text-secondary">
                {t('home.noScan')}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 3D Scene / Visual Area Container */}
      <Card className="min-h-[220px] p-0 overflow-hidden flex items-center justify-center bg-surface border border-separator/35 shadow-sm">
        <React.Suspense
          fallback={
            <StaticFallback
              findingsCount={scanResult?.findings?.length ?? 0}
            />
          }
        >
          <DeviceScene
            status={scanResult ? 'completed' : 'idle'}
            findingsCount={scanResult?.findings?.length ?? 0}
          />
        </React.Suspense>
      </Card>

      {/* Onboarding Modal */}
      <OnboardingModal
        open={showOnboarding}
        onOpenChange={setShowOnboarding}
      />
    </div>
  );
};
