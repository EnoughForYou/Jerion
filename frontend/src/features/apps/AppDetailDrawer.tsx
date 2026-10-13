import React, { useState } from 'react';
import {
  ShieldAlert,
  PackageCheck,
  Download,
  Search,
  Wrench,
  Check,
  AlertTriangle,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { DetailPanel } from '../../app/DetailPanel';
import { Button } from '../../components/ui/Button';
import { getBridge } from '../../lib/bridge';
import type { AppInfo } from '../../lib/bridge/schema';

export interface AppDetailDrawerProps {
  app: AppInfo | null;
  onClose: () => void;
}

export const AppDetailDrawer: React.FC<AppDetailDrawerProps> = ({
  app,
  onClose,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [isPulling, setIsPulling] = useState(false);
  const [pulledPath, setPulledPath] = useState<string | null>(null);

  const bridge = getBridge();

  if (!app) return null;

  const handlePullApk = async () => {
    try {
      setIsPulling(true);
      setPulledPath(null);
      const path = await bridge.inventory.extractApk(app.package);
      setPulledPath(path);
    } catch (err) {
      console.error('Failed to pull APK:', err);
    } finally {
      setIsPulling(false);
    }
  };

  const handleDeepScan = () => {
    onClose();
    navigate('/scan');
  };

  const handleRemediate = () => {
    onClose();
    navigate('/remediate');
  };

  return (
    <DetailPanel
      open={Boolean(app)}
      onClose={onClose}
      title={app.label || app.package}
      description={app.package}
    >
      <div className="space-y-4 text-xs">
        {/* Action Buttons */}
        <div className="p-3 rounded-md bg-raised/40 border border-separator/20 space-y-2">
          <span className="font-semibold text-text-primary text-[11px] block">
            {t('apps.detail.actions')}
          </span>

          <div className="flex flex-wrap gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={handlePullApk}
              disabled={isPulling}
            >
              {pulledPath ? (
                <>
                  <Check className="w-3.5 h-3.5 text-ok" />
                  <span>Pulled</span>
                </>
              ) : (
                <>
                  <Download className={`w-3.5 h-3.5 ${isPulling ? 'animate-bounce' : ''}`} />
                  <span>{isPulling ? t('apps.detail.pulling') : t('apps.detail.pullApk')}</span>
                </>
              )}
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={handleDeepScan}
            >
              <Search className="w-3.5 h-3.5 text-accent" />
              <span>{t('apps.detail.deepScan')}</span>
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={handleRemediate}
            >
              <Wrench className="w-3.5 h-3.5 text-sev-high" />
              <span>{t('apps.detail.remediate')}</span>
            </Button>
          </div>

          {pulledPath && (
            <p className="text-[11px] font-mono text-ok break-all bg-ok/10 p-2 rounded border border-ok/20">
              {t('apps.detail.pulledSuccess', { path: pulledPath })}
            </p>
          )}
        </div>

        {/* General Attributes */}
        <div className="space-y-2 border-b border-separator/20 pb-3">
          <div className="grid grid-cols-2 gap-2">
            <div>
              <span className="text-text-secondary block font-medium">
                {t('apps.detail.versionCode')}:
              </span>
              <span className="font-mono text-text-primary">{app.versionCode}</span>
            </div>
            <div>
              <span className="text-text-secondary block font-medium">
                {t('apps.detail.targetSdk')}:
              </span>
              <span className="text-text-primary">API {app.targetSdk} (min {app.minSdk})</span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div>
              <span className="text-text-secondary block font-medium">
                {t('apps.detail.firstInstall')}:
              </span>
              <span className="text-text-primary">{app.installedAt.slice(0, 10)}</span>
            </div>
            <div>
              <span className="text-text-secondary block font-medium">
                {t('apps.detail.lastUpdate')}:
              </span>
              <span className="text-text-primary">{app.updatedAt.slice(0, 10)}</span>
            </div>
          </div>

          <div>
            <span className="text-text-secondary block font-medium">
              {t('apps.detail.installer')}:
            </span>
            <span className="font-mono text-text-primary text-[11px]">
              {app.installerPackage || t('apps.detail.installerUnknown')}
            </span>
          </div>
        </div>

        {/* Suspicious Flags / Signals */}
        {app.suspiciousFlags.length > 0 && (
          <div className="p-3 rounded-md bg-sev-high/10 border border-sev-high/25 space-y-1.5">
            <div className="flex items-center gap-1.5 font-semibold text-sev-high">
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>Security Signals ({app.suspiciousFlags.length})</span>
            </div>
            <div className="flex flex-wrap gap-1.5 pt-1">
              {app.suspiciousFlags.map((flag) => (
                <span
                  key={flag}
                  className="px-2 py-0.5 rounded text-[10px] font-mono bg-sev-high/15 text-sev-high border border-sev-high/30"
                >
                  {flag}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Dangerous Permissions */}
        {app.dangerousPermissions.length > 0 && (
          <div>
            <h4 className="font-semibold text-sev-high mb-2 flex items-center gap-1.5">
              <ShieldAlert className="w-3.5 h-3.5" />
              <span>{t('apps.detail.dangerousPermissions', { count: app.dangerousPermissions.length })}</span>
            </h4>
            <div className="max-h-36 overflow-y-auto space-y-1 bg-sev-high/5 p-2 rounded-md border border-sev-high/20 font-mono text-[11px]">
              {app.dangerousPermissions.map((perm) => (
                <div key={perm} className="text-sev-high truncate">
                  {perm}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Requested Permissions */}
        <div>
          <h4 className="font-semibold text-text-primary mb-2 flex items-center gap-1.5">
            <ShieldAlert className="w-3.5 h-3.5 text-accent" />
            <span>{t('apps.detail.permissions', { count: app.permissions.length })}</span>
          </h4>
          <div className="max-h-44 overflow-y-auto space-y-1 bg-raised/30 p-2 rounded-md border border-separator/25 font-mono text-[11px]">
            {app.permissions.length === 0 ? (
              <span className="text-text-secondary">{t('apps.detail.noPermissions')}</span>
            ) : (
              app.permissions.map((perm) => (
                <div key={perm} className="text-text-secondary truncate">
                  {perm}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Signatures */}
        <div>
          <h4 className="font-semibold text-text-primary mb-2 flex items-center gap-1.5">
            <PackageCheck className="w-3.5 h-3.5 text-accent" />
            <span>{t('apps.detail.signatures', { count: app.signatures.length })}</span>
          </h4>
          <div className="space-y-1 bg-raised/30 p-2 rounded-md border border-separator/25 font-mono text-[11px]">
            {app.signatures.map((sig, idx) => (
              <div key={idx} className="break-all text-text-secondary">
                {sig}
              </div>
            ))}
          </div>
        </div>
      </div>
    </DetailPanel>
  );
};
