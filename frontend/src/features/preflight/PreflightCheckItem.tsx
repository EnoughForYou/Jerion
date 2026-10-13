import React from 'react';
import {
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  XCircle,
  Download,
  Loader2,
  FolderOpen,
  RotateCcw,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from '../../components/ui/Button';
import type { PreflightCheck } from '../../lib/bridge/schema';

export interface PreflightCheckItemProps {
  check: PreflightCheck;
  onDownload?: (checkId: string) => void;
  onChoosePath?: (checkId: string) => void;
  onRetry?: (checkId: string) => void;
}

export const PreflightCheckItem: React.FC<PreflightCheckItemProps> = ({
  check,
  onDownload,
  onChoosePath,
  onRetry,
}) => {
  const { t } = useTranslation();

  const getStatusIcon = () => {
    switch (check.status) {
      case 'ok':
        return <CheckCircle2 className="w-4 h-4 text-ok shrink-0" />;
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-sev-medium shrink-0" />;
      case 'missing':
        return <AlertCircle className="w-4 h-4 text-sev-high shrink-0" />;
      case 'downloading':
        return <Download className="w-4 h-4 text-accent shrink-0 animate-pulse" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-sev-critical shrink-0" />;
      case 'checking':
      default:
        return <Loader2 className="w-4 h-4 text-accent animate-spin shrink-0" />;
    }
  };

  const isActionable = check.status === 'missing' || check.status === 'failed';

  return (
    <div className="flex flex-col border border-separator/25 rounded-md bg-raised/30 p-3 transition-all duration-150">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2.5 truncate">
          {getStatusIcon()}
          <span className="font-medium text-xs text-text-primary truncate">
            {check.title}
          </span>
        </div>
        <span className="text-[11px] text-text-secondary truncate text-right">
          {check.detail}
        </span>
      </div>

      {/* Progress bar during download */}
      {check.status === 'downloading' && typeof check.downloadProgress === 'number' ? (
        <div className="mt-2.5">
          <div className="flex justify-between text-[10px] text-text-secondary mb-1">
            <span>{t('preflight.status.downloading')}</span>
            <span>{Math.round(check.downloadProgress)}%</span>
          </div>
          <div className="h-1.5 w-full bg-raised rounded-full overflow-hidden">
            <div
              className="h-full bg-accent transition-all duration-150"
              style={{ width: `${check.downloadProgress}%` }}
            />
          </div>
        </div>
      ) : null}

      {/* Actionable expanded drawer on missing/failed requirements */}
      {isActionable ? (
        <div className="mt-3 pt-2.5 border-t border-separator/20 text-xs text-text-secondary">
          {check.remedy ? (
            <p className="mb-3 text-[11px] leading-relaxed text-text-secondary">
              {check.remedy}
            </p>
          ) : null}

          <div className="flex flex-wrap items-center gap-2">
            {check.sourceUrl && onDownload ? (
              <Button
                variant="primary"
                size="sm"
                onClick={() => onDownload(check.id)}
              >
                <Download className="w-3.5 h-3.5" />
                {t('common.download')}
              </Button>
            ) : null}

            {onChoosePath ? (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => onChoosePath(check.id)}
              >
                <FolderOpen className="w-3.5 h-3.5" />
                {t('common.choosePath')}
              </Button>
            ) : null}

            {onRetry ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRetry(check.id)}
              >
                <RotateCcw className="w-3.5 h-3.5" />
                {t('common.retry')}
              </Button>
            ) : null}
          </div>
        </div>
      ) : null}
    </div>
  );
};
