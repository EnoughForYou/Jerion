import React, { useState } from 'react';
import { Smartphone, Battery, BatteryCharging, RotateCcw, Moon, Sun, Wifi, Wrench } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { SimpleTooltip } from '../components/ui/Tooltip';
import { useDeviceStore } from './deviceStore';
import { useThemeStore } from './theme';
import { ConnectionTroubleshootModal } from '../features/home/ConnectionTroubleshootModal';
import { WirelessConnectModal } from '../features/home/WirelessConnectModal';

export interface DeviceHeaderProps {
  onRefresh?: () => void;
  isRefreshing?: boolean;
}

export const DeviceHeader: React.FC<DeviceHeaderProps> = ({
  onRefresh,
  isRefreshing = false,
}) => {
  const { t } = useTranslation();
  const { activeDevice } = useDeviceStore();
  const { theme, toggleTheme } = useThemeStore();
  const [showTroubleshoot, setShowTroubleshoot] = useState(false);
  const [showWireless, setShowWireless] = useState(false);

  const getStatusBadge = () => {
    if (!activeDevice) return null;
    switch (activeDevice.connectionState) {
      case 'ready':
        return <Badge severity="ok">{t('device.status.online')}</Badge>;
      case 'unauthorized':
        return (
          <SimpleTooltip content={t('device.unauthorizedPrompt')}>
            <Badge severity="high">{t('device.status.unauthorized')}</Badge>
          </SimpleTooltip>
        );
      case 'offline':
      case 'no_device':
      default:
        return <Badge severity="info">{t('device.status.offline')}</Badge>;
    }
  };

  return (
    <>
      <header className="h-12 border-b border-separator-subtle bg-surface/50 backdrop-blur-sm px-4 flex items-center justify-between shrink-0 select-none">
        {/* Device Information section */}
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-md bg-raised flex items-center justify-center text-accent shrink-0">
            <Smartphone className="w-4 h-4" />
          </div>

          {activeDevice ? (
            <div className="flex items-center gap-3">
              <div className="flex flex-col">
                <span className="text-base font-semibold text-text-primary leading-tight">
                  {activeDevice.model}
                </span>
                <span className="text-[11px] text-text-secondary">
                  {activeDevice.connectionState === 'ready'
                    ? `${t('device.android')} ${activeDevice.androidVersion} (${t('device.api')} ${activeDevice.sdkInt}) · ${activeDevice.serial}`
                    : `${t('device.pendingAuth', 'Pending Authorization')} · ${activeDevice.serial}`}
                </span>
              </div>

              <div className="flex items-center gap-2 ml-2">
                {getStatusBadge()}

                {activeDevice.connectionState === 'ready' && typeof activeDevice.batteryLevel === 'number' && activeDevice.batteryLevel > 0 ? (
                  <span className="inline-flex items-center gap-1 text-xs text-text-secondary bg-raised/70 px-2 py-0.5 rounded-sm border border-separator-muted">
                    {activeDevice.isCharging ? (
                      <BatteryCharging className="w-3.5 h-3.5 text-ok" />
                    ) : (
                      <Battery className="w-3.5 h-3.5" />
                    )}
                    <span>{activeDevice.batteryLevel}%</span>
                  </span>
                ) : null}
              </div>
            </div>
          ) : (
            <div className="flex flex-col">
              <span className="text-base font-medium text-text-secondary">
                {t('device.noDevice')}
              </span>
              <span className="text-[11px] text-text-secondary/70">
                {t('device.noDeviceHint')}
              </span>
            </div>
          )}
        </div>

        {/* Actions section */}
        <div className="flex items-center gap-2">
          <SimpleTooltip content={t('device.wireless.button', 'Connect via Wi-Fi')}>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowWireless(true)}
              aria-label={t('device.wireless.button', 'Connect via Wi-Fi')}
            >
              <Wifi className="w-3.5 h-3.5" />
            </Button>
          </SimpleTooltip>

          <SimpleTooltip content={t('device.troubleshoot.button', 'Connection Troubleshoot')}>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowTroubleshoot(true)}
              aria-label={t('device.troubleshoot.button', 'Connection Troubleshoot')}
            >
              <Wrench className="w-3.5 h-3.5" />
            </Button>
          </SimpleTooltip>

          {onRefresh ? (
            <SimpleTooltip content={t('device.refresh')}>
              <Button
                variant="ghost"
                size="sm"
                onClick={onRefresh}
                disabled={isRefreshing}
                aria-label={t('device.refresh')}
              >
                <RotateCcw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
              </Button>
            </SimpleTooltip>
          ) : null}

          <SimpleTooltip content={theme === 'dark' ? t('settings.themeLight') : t('settings.themeDark')}>
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleTheme}
              aria-label="Toggle Theme"
            >
              {theme === 'dark' ? <Sun className="w-3.5 h-3.5" /> : <Moon className="w-3.5 h-3.5" />}
            </Button>
          </SimpleTooltip>
        </div>
      </header>

      <ConnectionTroubleshootModal
        open={showTroubleshoot}
        onOpenChange={setShowTroubleshoot}
        onOpenWireless={() => setShowWireless(true)}
      />

      <WirelessConnectModal
        open={showWireless}
        onOpenChange={setShowWireless}
        onConnected={() => {
          if (onRefresh) {
            onRefresh();
          }
        }}
      />
    </>
  );
};
