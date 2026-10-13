import React, { useEffect, useState, useCallback } from 'react';
import { ShieldCheck } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { PreflightCheckItem } from './PreflightCheckItem';
import { getBridge } from '../../lib/bridge';
import type { PreflightCheck } from '../../lib/bridge/schema';

export interface PreflightViewProps {
  onComplete: () => void;
}

export const PreflightView: React.FC<PreflightViewProps> = ({ onComplete }) => {
  const { t } = useTranslation();
  const [checks, setChecks] = useState<PreflightCheck[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const checkCompletion = useCallback((currentChecks: PreflightCheck[]) => {
    if (currentChecks.length === 0) return;
    const allPassed = currentChecks.every(
      (c) => c.status === 'ok' || c.status === 'warning'
    );
    if (allPassed) {
      onComplete();
    }
  }, [onComplete]);

  useEffect(() => {
    let isMounted = true;
    const bridge = getBridge();

    async function runChecks() {
      try {
        setIsLoading(true);
        const results = await bridge.preflight.run();
        if (isMounted) {
          setChecks(results);
          checkCompletion(results);
        }
      } catch (err) {
        console.error('Failed to run preflight checks:', err);
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    const unsubUpdate = bridge.on('preflight:update', (updated: PreflightCheck) => {
      setChecks((prev) => {
        const next = prev.map((c) => (c.id === updated.id ? updated : c));
        checkCompletion(next);
        return next;
      });
    });

    runChecks();

    return () => {
      isMounted = false;
      unsubUpdate();
    };
  }, [checkCompletion]);

  const handleDownload = async () => {
    try {
      const bridge = getBridge();
      await bridge.preflight.downloadAdb();
    } catch (err) {
      console.error('Download resolution failed:', err);
    }
  };

  const handleChoosePath = async () => {
    try {
      const bridge = getBridge();
      await bridge.preflight.selectAdbPath('C:\\platform-tools\\adb.exe');
    } catch (err) {
      console.error('Choose path resolution failed:', err);
    }
  };

  const handleRetry = async () => {
    try {
      const bridge = getBridge();
      const results = await bridge.preflight.retry();
      setChecks(results);
      checkCompletion(results);
    } catch (err) {
      console.error('Retry resolution failed:', err);
    }
  };

  return (
    <div className="flex h-screen w-screen items-center justify-center bg-base p-4 select-none">
      <div className="w-full max-w-[480px] bg-surface border border-separator/35 rounded-lg p-6 shadow-xl text-text-primary">
        {/* Header */}
        <div className="flex items-center gap-3 mb-4 pb-3 border-b border-separator/20">
          <div className="w-10 h-10 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('preflight.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('preflight.description')}
            </p>
          </div>
        </div>

        {/* Checks List */}
        <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
          {isLoading && checks.length === 0 ? (
            <div className="py-8 text-center text-xs text-text-secondary animate-pulse">
              {t('preflight.status.checking')}
            </div>
          ) : (
            checks.map((check) => (
              <PreflightCheckItem
                key={check.id}
                check={check}
                onDownload={handleDownload}
                onChoosePath={handleChoosePath}
                onRetry={handleRetry}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
};
