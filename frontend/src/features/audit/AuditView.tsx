import React, { useState, useEffect } from 'react';
import { ClipboardCheck, RotateCcw, AlertTriangle } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { ErrorState } from '../../components/ui/ErrorState';
import { getBridge } from '../../lib/bridge';
import type { AuditItem } from '../../lib/bridge/schema';

export const AuditView: React.FC = () => {
  const { t } = useTranslation();
  const [items, setItems] = useState<AuditItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const bridge = getBridge();

  const handleManualRun = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const results = await bridge.audit.run();
      setItems(results);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Audit check execution failed');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;

    async function loadInitialAudit() {
      try {
        setError(null);
        const results = await bridge.audit.run();
        if (isMounted) {
          setItems(results);
        }
      } catch (err: unknown) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Audit check execution failed');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadInitialAudit();

    return () => {
      isMounted = false;
    };
  }, [bridge]);

  if (error) {
    return (
      <ErrorState
        title="Failed to Run System Audit"
        remedy="Check ADB connection and verify settings service access."
        onRetry={handleManualRun}
      />
    );
  }

  const getStatusBadge = (status: AuditItem['status']) => {
    switch (status) {
      case 'ok':
        return <Badge severity="ok">Compliant</Badge>;
      case 'warning':
        return <Badge severity="medium">Warning</Badge>;
      case 'issue':
        return <Badge severity="critical">Issue</Badge>;
      case 'unavailable':
      default:
        return <Badge severity="info">Not Available</Badge>;
    }
  };

  return (
    <div className="space-y-5">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <ClipboardCheck className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('audit.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('audit.description')}
            </p>
          </div>
        </div>

        <Button
          variant="secondary"
          size="md"
          onClick={handleManualRun}
          disabled={isLoading}
        >
          <RotateCcw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          {t('audit.run')}
        </Button>
      </div>

      {/* Audit Items List */}
      <Card>
        <CardHeader>
          <CardTitle>System Configuration Hardening</CardTitle>
          <CardDescription>
            Audit results evaluating security options, developer flags, and trust parameters
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3 p-2">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : (
            <div className="space-y-3">
              {items.map((item) => (
                <div
                  key={item.id}
                  className="p-3.5 rounded-md bg-raised/30 border border-separator/25 space-y-2 text-xs"
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2.5">
                      {getStatusBadge(item.status)}
                      <span className="font-semibold text-text-primary text-xs">
                        {item.title}
                      </span>
                    </div>
                    <span className="text-[10px] uppercase font-mono text-text-secondary px-1.5 py-0.5 rounded-xs bg-raised border border-separator/20">
                      {item.category}
                    </span>
                  </div>

                  <p className="text-text-secondary text-[11px] leading-relaxed">
                    {item.summary}
                  </p>

                  {item.remedy ? (
                    <div className="pt-2 border-t border-separator/15 text-[11px] text-accent flex items-center gap-1.5">
                      <AlertTriangle className="w-3.5 h-3.5 text-sev-medium shrink-0" />
                      <span>{item.remedy}</span>
                    </div>
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
