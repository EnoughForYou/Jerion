import React from 'react';
import { Activity } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useDeviceStore } from './deviceStore';

import { bridgeKind } from '../lib/bridge';

export const StatusBar: React.FC = () => {
  const { t } = useTranslation();
  const { activeJob, jobProgress } = useDeviceStore();

  const ruleVersion = '2026.03.1';

  return (
    <footer className="h-7 border-t border-separator-subtle bg-surface/70 backdrop-blur-xs px-3 flex items-center justify-between text-[11px] text-text-secondary select-none shrink-0 z-10">
      {/* Active Job & Progress */}
      <div className="flex items-center gap-2">
        <Activity className={`w-3 h-3 ${activeJob ? 'text-accent animate-pulse' : 'text-text-secondary'}`} />
        <span>
          {activeJob
            ? t('statusBar.running', { job: activeJob })
            : t('statusBar.idle')}
        </span>
        {typeof jobProgress === 'number' ? (
          <div className="w-16 h-1.5 bg-raised rounded-full overflow-hidden ml-1">
            <div
              className="h-full bg-accent transition-all duration-200"
              style={{ width: `${Math.min(100, Math.max(0, jobProgress))}%` }}
            />
          </div>
        ) : null}
      </div>

      {/* Rules version & Dev Bridge mode */}
      <div className="flex items-center gap-3">
        <span>{t('statusBar.ruleVersion', { version: ruleVersion })}</span>
        {import.meta.env.DEV && (
          <>
            <span className="text-separator">|</span>
            <span className="font-mono text-[10px] uppercase px-1.5 py-0.2 rounded-xs bg-raised border border-separator text-text-primary">
              {t('statusBar.bridgeMode', { mode: bridgeKind.toUpperCase() })}
            </span>
          </>
        )}
      </div>
    </footer>
  );
};
