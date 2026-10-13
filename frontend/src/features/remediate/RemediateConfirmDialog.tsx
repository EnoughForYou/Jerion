import React, { useState } from 'react';
import { AlertOctagon, ShieldCheck, AlertTriangle } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/Dialog';
import { Button } from '../../components/ui/Button';
import type { RemediationAction, RemediationPlan } from '../../lib/bridge/schema';

export interface RemediateConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  actions: RemediationAction[];
  onConfirm: (plan: RemediationPlan) => void;
  isExecuting: boolean;
}

export const RemediateConfirmDialog: React.FC<RemediateConfirmDialogProps> = ({
  open,
  onOpenChange,
  actions,
  onConfirm,
  isExecuting,
}) => {
  const { t } = useTranslation();
  const [requiresBackup, setRequiresBackup] = useState(true);

  const handleExecute = () => {
    onConfirm({
      actions,
      requiresBackup,
    });
  };

  const hasUninstallAction = actions.some((a) => a.type === 'uninstall');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-1 text-sev-high">
            <AlertOctagon className="w-5 h-5" />
            <DialogTitle className="text-base font-semibold text-text-primary">
              {t('cleanup.confirm.title')}
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-text-secondary leading-relaxed">
            {t('cleanup.confirm.description', { count: actions.length })}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3.5 py-2 text-xs">
          {/* Actions summary list */}
          <div className="max-h-48 overflow-y-auto space-y-2 p-3 bg-raised/30 rounded-md border border-separator/20">
            {actions.map((act) => (
              <div
                key={act.id}
                className="flex items-start justify-between gap-2 p-2 rounded bg-surface/70 border border-separator/15"
              >
                <div>
                  <span className="font-semibold text-text-primary block">
                    {act.title}
                  </span>
                  <span className="font-mono text-[11px] text-text-secondary block">
                    {act.package}
                  </span>
                </div>
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-raised text-text-secondary border border-separator/20 uppercase shrink-0">
                  {act.type}
                </span>
              </div>
            ))}
          </div>

          {/* Backup Checkbox Toggle */}
          <label className="flex items-start gap-2.5 p-3 rounded-md bg-accent/5 border border-accent/20 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={requiresBackup}
              onChange={(e) => setRequiresBackup(e.target.checked)}
              className="mt-0.5 rounded border-separator text-accent focus:ring-accent"
            />
            <div>
              <span className="font-semibold text-text-primary block text-xs">
                {t('cleanup.confirm.backupOption')}
              </span>
              <span className="text-[11px] text-text-secondary">
                Backs up existing APK package files to safe storage prior to execution.
              </span>
            </div>
          </label>

          {/* Warning on irreversible actions */}
          {hasUninstallAction && (
            <div className="flex items-start gap-2 p-2.5 rounded bg-sev-high/10 border border-sev-high/20 text-sev-high text-[11px]">
              <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
              <span>{t('cleanup.confirm.irreversibleWarning')}</span>
            </div>
          )}
        </div>

        <DialogFooter className="flex items-center justify-between sm:justify-between w-full">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onOpenChange(false)}
            disabled={isExecuting}
          >
            {t('cleanup.confirm.cancel')}
          </Button>

          <Button
            variant="destructive"
            size="md"
            onClick={handleExecute}
            disabled={isExecuting}
          >
            <ShieldCheck className="w-4 h-4" />
            <span>
              {isExecuting
                ? t('cleanup.executing')
                : t('cleanup.confirm.confirmAction')}
            </span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
