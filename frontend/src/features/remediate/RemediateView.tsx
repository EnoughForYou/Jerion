import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldCheck,
  RotateCcw,
  AlertOctagon,
  AlertTriangle,
  History,
  CheckCircle2,
  Undo2,
  Trash2,
  Ban,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { RemediateConfirmDialog } from './RemediateConfirmDialog';
import { getBridge } from '../../lib/bridge';
import type { ActionLogEntry, RemediationAction, RemediationPlan } from '../../lib/bridge/schema';

interface ExecutionFeedback {
  type: 'ok' | 'warning' | 'error';
  message: string;
  auditWarning?: string;
}

export const RemediateView: React.FC = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<ActionLogEntry[]>([]);
  const [stagedActions, setStagedActions] = useState<RemediationAction[]>([]);
  const [selectedActionIds, setSelectedActionIds] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);
  const [isExecuting, setIsExecuting] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [feedback, setFeedback] = useState<ExecutionFeedback | null>(null);

  const bridge = getBridge();

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true);

      // Load action history
      const history = await bridge.remediate.getHistory();
      setLogs(history || []);

      // Load actionable findings from latest report to generate staged remediation plan
      const reports = await bridge.report.list();
      if (reports && reports.length > 0) {
        const latestScan = await bridge.report.get(reports[0].id);
        if (latestScan && latestScan.findings) {
          const findingIds = latestScan.findings.map((f) => f.id);
          const plan = await bridge.remediate.createPlan(findingIds);
          if (plan && plan.actions) {
            setStagedActions(plan.actions);
            // Select all non-blocked actions by default
            setSelectedActionIds(
              new Set(plan.actions.filter((a) => !a.isBlocked).map((a) => a.id))
            );
          }
        }
      }
    } catch (err) {
      console.error('Failed to load remediation data:', err);
    } finally {
      setIsLoading(false);
    }
  }, [bridge]);

  useEffect(() => {
    let isMounted = true;

    async function init() {
      await loadData();
      if (!isMounted) return;
    }

    init();

    return () => {
      isMounted = false;
    };
  }, [loadData]);

  const handleToggleSelect = (actionId: string) => {
    setSelectedActionIds((prev) => {
      const next = new Set(prev);
      if (next.has(actionId)) {
        next.delete(actionId);
      } else {
        next.add(actionId);
      }
      return next;
    });
  };

  const handleExecutePlan = async (plan: RemediationPlan) => {
    try {
      setIsExecuting(true);
      const newLogs = await bridge.remediate.execute(plan);
      const logsList = newLogs || [];

      // Remove executed actions from staged list
      const executedIds = new Set(plan.actions.map((a) => a.id));
      setStagedActions((prev) => prev.filter((a) => !executedIds.has(a.id)));
      setSelectedActionIds(new Set());

      // Update history log
      setLogs((prev) => [...logsList, ...prev]);
      setShowConfirmDialog(false);

      if (logsList.length === 0) {
        return;
      }

      const successCount = logsList.filter((l) => l.status === 'success').length;
      const failedCount = logsList.filter((l) => l.status === 'failed').length;
      const auditFailures = logsList.filter((l) => l.auditSaved === false).length;

      let type: 'ok' | 'warning' | 'error' = 'ok';
      let message = '';

      if (failedCount === 0) {
        type = auditFailures > 0 ? 'warning' : 'ok';
        message = t('cleanup.executionSuccess', { count: successCount });
      } else if (successCount > 0) {
        type = 'warning';
        message = t('cleanup.executionPartial', {
          success: successCount,
          total: logsList.length,
          failed: failedCount,
        });
      } else {
        type = 'error';
        message = t('cleanup.executionFailed', { count: failedCount });
      }

      let auditWarning: string | undefined;
      if (auditFailures > 0) {
        auditWarning = t('cleanup.auditSaveWarning', { count: auditFailures });
      }

      setFeedback({ type, message, auditWarning });
      setTimeout(() => setFeedback(null), 5000);
    } catch (err) {
      console.error('Failed to execute remediation plan:', err);
      setShowConfirmDialog(false);
      setFeedback({
        type: 'error',
        message:
          (err as Error)?.message ||
          t('cleanup.executionFailed', { count: plan.actions.length }),
      });
      setTimeout(() => setFeedback(null), 5000);
    } finally {
      setIsExecuting(false);
    }
  };

  const handleUndo = async (actionId: string) => {
    try {
      const updatedLog = await bridge.remediate.undo(actionId);
      setLogs((prev) =>
        prev.map((log) => (log.id === updatedLog.id ? updatedLog : log))
      );
      setFeedback({
        type: 'ok',
        message: t('cleanup.undoSuccess'),
      });
      setTimeout(() => setFeedback(null), 4000);
    } catch (err) {
      console.error('Failed to undo action:', err);
      setFeedback({
        type: 'error',
        message: (err as Error)?.message || 'Failed to undo action',
      });
      setTimeout(() => setFeedback(null), 5000);
    }
  };

  const selectedActionsToExecute = stagedActions.filter((a) =>
    selectedActionIds.has(a.id)
  );

  return (
    <div className="space-y-5">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg shadow-sm">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('cleanup.title')}
            </h2>
            <p className="text-xs text-text-secondary leading-relaxed">
              {t('cleanup.description')}
            </p>
          </div>
        </div>

        <Button
          variant="secondary"
          size="sm"
          onClick={loadData}
          disabled={isLoading}
        >
          <RotateCcw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          <span>{t('cleanup.refresh')}</span>
        </Button>
      </div>

      {feedback && (
        <div
          className={`flex flex-col gap-1.5 p-3 rounded-md text-xs border ${
            feedback.type === 'ok'
              ? 'bg-ok/10 border-ok/25 text-ok'
              : feedback.type === 'warning'
              ? 'bg-sev-medium/10 border-sev-medium/25 text-sev-medium'
              : 'bg-sev-high/10 border-sev-high/25 text-sev-high'
          }`}
        >
          <div className="flex items-center gap-2">
            {feedback.type === 'ok' ? (
              <CheckCircle2 className="w-4 h-4 shrink-0" />
            ) : feedback.type === 'warning' ? (
              <AlertTriangle className="w-4 h-4 shrink-0" />
            ) : (
              <AlertOctagon className="w-4 h-4 shrink-0" />
            )}
            <span className="font-medium">{feedback.message}</span>
          </div>
          {feedback.auditWarning && (
            <div className="flex items-center gap-2 pl-6 text-[11px] opacity-90 font-mono">
              <span>{feedback.auditWarning}</span>
            </div>
          )}
        </div>
      )}

      {/* Action Guardrails Notice */}
      <div className="p-3.5 rounded-lg bg-raised/40 border border-separator/30 text-xs flex items-center gap-3">
        <AlertOctagon className="w-4 h-4 text-sev-high shrink-0" />
        <span className="text-text-secondary leading-relaxed">
          {t('cleanup.guardrailNotice')}
        </span>
      </div>

      {/* Grid: Pending Staged Actions vs History Log */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* Pending Remediation Actions */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>{t('cleanup.pendingActions')}</CardTitle>
                <CardDescription>{t('cleanup.pendingDesc')}</CardDescription>
              </div>

              {selectedActionsToExecute.length > 0 && (
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => setShowConfirmDialog(true)}
                  disabled={isExecuting}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>
                    {t('cleanup.executeSelected', {
                      count: selectedActionsToExecute.length,
                    })}
                  </span>
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-3 py-2">
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
              </div>
            ) : stagedActions.length === 0 ? (
              <div className="py-12 text-center text-xs text-text-secondary">
                {t('cleanup.noActions')}
              </div>
            ) : (
              <div className="space-y-3">
                {stagedActions.map((action) => {
                  const isSelected = selectedActionIds.has(action.id);

                  return (
                    <div
                      key={action.id}
                      className={`p-3.5 rounded-md border text-xs transition-colors ${
                        action.isBlocked
                          ? 'bg-raised/20 border-separator/20 opacity-70'
                          : isSelected
                          ? 'bg-accent/5 border-accent/40'
                          : 'bg-raised/30 border-separator/25 hover:border-separator/40'
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        <input
                          type="checkbox"
                          checked={isSelected}
                          disabled={action.isBlocked}
                          onChange={() => handleToggleSelect(action.id)}
                          className="mt-1 rounded border-separator text-accent focus:ring-accent"
                        />

                        <div className="flex-1 space-y-1 min-w-0">
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-semibold text-text-primary text-xs">
                              {action.title}
                            </span>
                            <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-raised text-text-secondary border border-separator/25">
                              {action.type}
                            </span>
                          </div>

                          <span className="font-mono text-[11px] text-text-secondary block truncate">
                            {action.package}
                          </span>

                          <p className="text-[11px] text-text-secondary leading-relaxed">
                            Impact: {action.impact}
                          </p>

                          {action.isBlocked && (
                            <div className="flex items-center gap-1 text-[11px] text-sev-high font-medium pt-1">
                              <Ban className="w-3.5 h-3.5" />
                              <span>Blocked: {action.blockedReason || 'Protected by safety guardrail'}</span>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Action History Log */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <History className="w-4 h-4 text-accent" />
              <CardTitle>{t('cleanup.actionLog')}</CardTitle>
            </div>
            <CardDescription>{t('cleanup.actionLogDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            {logs.length === 0 ? (
              <div className="py-12 text-center text-xs text-text-secondary">
                {t('cleanup.noHistory')}
              </div>
            ) : (
              <div className="space-y-2.5 max-h-[460px] overflow-y-auto pr-1">
                {logs.map((entry) => {
                  const isUndone = entry.status === 'undone';

                  return (
                    <div
                      key={entry.id}
                      className="p-3 rounded-md bg-raised/20 border border-separator/20 text-xs space-y-1.5"
                    >
                      <div className="flex items-center justify-between gap-2 flex-wrap">
                        <div className="flex items-center gap-2 truncate">
                          <Badge
                            severity={
                              isUndone
                                ? 'info'
                                : entry.status === 'success'
                                ? 'ok'
                                : 'critical'
                            }
                          >
                            {isUndone
                              ? t('cleanup.undone')
                              : entry.status.toUpperCase()}
                          </Badge>

                          {entry.auditSaved !== undefined && entry.auditSaved !== null && (
                            <Badge
                              severity={entry.auditSaved ? 'info' : 'critical'}
                              className="text-[10px]"
                            >
                              {entry.auditSaved
                                ? t('cleanup.auditSavedBadge')
                                : t('cleanup.auditFailedBadge')}
                            </Badge>
                          )}

                          <span className="font-semibold text-text-primary text-xs truncate">
                            {entry.action.title}
                          </span>
                        </div>

                        {entry.action.undoable && !isUndone && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleUndo(entry.id)}
                            className="h-6 text-[11px] px-2"
                          >
                            <Undo2 className="w-3 h-3 mr-1" />
                            <span>{t('cleanup.undo')}</span>
                          </Button>
                        )}
                      </div>

                      <div className="flex items-center justify-between text-[11px] text-text-secondary font-mono">
                        <span className="truncate">{entry.action.package}</span>
                        <span>{entry.timestamp.slice(0, 19).replace('T', ' ')}</span>
                      </div>

                      {entry.error && (
                        <div className="flex items-start gap-1.5 text-[11px] text-sev-high bg-sev-high/10 p-2 rounded border border-sev-high/20">
                          <AlertOctagon className="w-3.5 h-3.5 shrink-0 mt-0.5" />
                          <div>
                            <span className="font-semibold">{t('cleanup.actionErrorLabel')}: </span>
                            <span>{entry.error}</span>
                          </div>
                        </div>
                      )}

                      {entry.auditSaved === false && (
                        <div className="flex items-start gap-1.5 text-[11px] text-sev-medium bg-sev-medium/10 p-2 rounded border border-sev-medium/20">
                          <AlertTriangle className="w-3.5 h-3.5 shrink-0 mt-0.5" />
                          <div>
                            <span className="font-semibold">{t('cleanup.auditErrorLabel')}: </span>
                            <span>{entry.auditError || t('cleanup.auditFailedNotice')}</span>
                          </div>
                        </div>
                      )}

                      {entry.backupPath && (
                        <div className="text-[10px] text-text-secondary/80 font-mono truncate">
                          {t('cleanup.backupSaved', { path: entry.backupPath })}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Confirmation Dialog before Execution */}
      <RemediateConfirmDialog
        open={showConfirmDialog}
        onOpenChange={setShowConfirmDialog}
        actions={selectedActionsToExecute}
        onConfirm={handleExecutePlan}
        isExecuting={isExecuting}
      />
    </div>
  );
};
