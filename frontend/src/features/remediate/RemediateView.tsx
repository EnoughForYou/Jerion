import React, { useState, useEffect } from 'react';
import { ShieldCheck, RotateCcw, AlertOctagon, History } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { EmptyState } from '../../components/ui/EmptyState';
import { getBridge } from '../../lib/bridge';
import type { ActionLogEntry } from '../../lib/bridge/schema';

export const RemediateView: React.FC = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<ActionLogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const bridge = getBridge();

  const fetchLogs = async () => {
    try {
      setIsLoading(true);
      const history = await bridge.remediate.getHistory();
      setLogs(history);
    } catch (err) {
      console.error('Failed to load action log history:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;

    async function loadInitialLogs() {
      try {
        const history = await bridge.remediate.getHistory();
        if (isMounted) {
          setLogs(history);
        }
      } catch (err) {
        console.error('Failed to load action log history:', err);
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadInitialLogs();

    return () => {
      isMounted = false;
    };
  }, [bridge]);

  const handleUndo = async (actionId: string) => {
    try {
      await bridge.remediate.undo(actionId);
      await fetchLogs();
    } catch (err) {
      console.error('Failed to undo action:', err);
    }
  };

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('cleanup.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('cleanup.description')}
            </p>
          </div>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchLogs}>
          <RotateCcw className="w-3.5 h-3.5" />
          Refresh Log
        </Button>
      </div>

      {/* Action Guardrails Notice */}
      <div className="p-3.5 rounded-lg bg-raised/40 border border-separator/30 text-xs flex items-center gap-3">
        <AlertOctagon className="w-4 h-4 text-sev-high shrink-0" />
        <span className="text-text-secondary leading-relaxed">
          Critical system OEM packages are strictly protected by guardrail denylists to prevent device instability.
        </span>
      </div>

      {/* Historical Remediation Actions Log */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <History className="w-4 h-4 text-accent" />
            <CardTitle>{t('cleanup.actionLog')}</CardTitle>
          </div>
          <CardDescription>
            Audited removal, disable, and permission revocation entries
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3 p-2">
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </div>
          ) : logs.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No Remediation Actions Recorded"
                description={t('cleanup.noActions')}
              />
            </div>
          ) : (
            <div className="space-y-3">
              {logs.map((log) => (
                <div
                  key={log.id}
                  className="p-3.5 rounded-md bg-raised/30 border border-separator/25 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs"
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-text-primary capitalize">
                        {log.action.type} Action
                      </span>
                      <Badge severity={log.status === 'success' ? 'ok' : 'critical'}>
                        {log.status.toUpperCase()}
                      </Badge>
                    </div>
                    <span className="font-mono text-[11px] text-text-secondary block">
                      Target: {log.action.package} · {log.timestamp.slice(0, 19)}
                    </span>
                  </div>

                  {log.action.undoable ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleUndo(log.id)}
                    >
                      <RotateCcw className="w-3 h-3" />
                      {t('common.undo')}
                    </Button>
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
