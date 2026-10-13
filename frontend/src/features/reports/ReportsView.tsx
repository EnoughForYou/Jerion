import React, { useState, useEffect } from 'react';
import { FileText, Download, GitCompare, Calendar, CheckCircle2, ChevronDown, ChevronUp } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { EmptyState } from '../../components/ui/EmptyState';
import { getBridge } from '../../lib/bridge';
import type { ScanComparison, ScanSummaryItem } from '../../lib/bridge/schema';

export const ReportsView: React.FC = () => {
  const { t } = useTranslation();
  const [scans, setScans] = useState<ScanSummaryItem[]>([]);
  const [comparison, setComparison] = useState<ScanComparison | null>(null);
  const [showDiffDetails, setShowDiffDetails] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [exportFeedback, setExportFeedback] = useState<string | null>(null);

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadReports() {
      try {
        setIsLoading(true);
        const list = await bridge.report.list();
        if (isMounted) {
          setScans(list);
        }

        if (list.length >= 2) {
          const comp = await bridge.report.compare(list[0].id, list[1].id);
          if (isMounted) {
            setComparison(comp);
          }
        }
      } catch (err) {
        console.error('Failed to load reports:', err);
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadReports();

    return () => {
      isMounted = false;
    };
  }, [bridge]);

  const handleExport = async (format: 'json' | 'html', scanId: string) => {
    try {
      const exportedPath = await bridge.report.export({
        scanId,
        format,
      });
      setExportFeedback(t('reports.exportSuccess', { path: exportedPath }));
      setTimeout(() => setExportFeedback(null), 5000);
    } catch (err) {
      console.error('Export failed:', err);
    }
  };

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg shadow-sm">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('reports.title')}
            </h2>
            <p className="text-xs text-text-secondary leading-relaxed">
              {t('reports.description')}
            </p>
          </div>
        </div>
      </div>

      {exportFeedback && (
        <div className="flex items-center gap-2 p-3 rounded-md bg-ok/10 border border-ok/25 text-xs text-ok">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{exportFeedback}</span>
        </div>
      )}

      {/* Comparison Delta Preview (if >= 2 scans exist) */}
      {comparison && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <GitCompare className="w-4 h-4 text-accent" />
                <CardTitle>{t('reports.compare')}</CardTitle>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowDiffDetails((prev) => !prev)}
              >
                <span>{showDiffDetails ? 'Hide Diff Details' : 'View Diff Details'}</span>
                {showDiffDetails ? (
                  <ChevronUp className="w-3.5 h-3.5 ml-1" />
                ) : (
                  <ChevronDown className="w-3.5 h-3.5 ml-1" />
                )}
              </Button>
            </div>
            <CardDescription>
              {t('reports.compareDesc')} ({comparison.scanA.id} vs {comparison.scanB.id})
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div className="p-3 bg-sev-critical/5 rounded-md border border-sev-critical/20">
                <span className="text-md font-bold text-sev-critical block">
                  +{comparison.newFindings.length}
                </span>
                <span className="text-[11px] text-text-secondary font-medium">
                  {t('reports.newFindings')}
                </span>
              </div>
              <div className="p-3 bg-ok/5 rounded-md border border-ok/20">
                <span className="text-md font-bold text-ok block">
                  -{comparison.resolvedFindings.length}
                </span>
                <span className="text-[11px] text-text-secondary font-medium">
                  {t('reports.resolvedFindings')}
                </span>
              </div>
              <div className="p-3 bg-raised/30 rounded-md border border-separator/20">
                <span className="text-md font-bold text-text-primary block">
                  {comparison.unchangedFindings.length}
                </span>
                <span className="text-[11px] text-text-secondary font-medium">
                  {t('reports.unchanged')}
                </span>
              </div>
            </div>

            {/* Expanded Detailed Delta */}
            {showDiffDetails && (
              <div className="space-y-3 pt-2 border-t border-separator/20 text-xs">
                {comparison.newFindings.length > 0 && (
                  <div className="space-y-1.5">
                    <span className="font-semibold text-sev-critical text-xs block">
                      New Threats Detected:
                    </span>
                    {comparison.newFindings.map((f) => (
                      <div
                        key={f.id}
                        className="p-2 rounded bg-sev-critical/5 border border-sev-critical/15 flex items-center justify-between"
                      >
                        <div className="flex items-center gap-2">
                          <Badge severity={f.severity} />
                          <span className="font-medium text-text-primary">{f.title}</span>
                        </div>
                        <span className="font-mono text-[11px] text-text-secondary">{f.ruleId}</span>
                      </div>
                    ))}
                  </div>
                )}

                {comparison.resolvedFindings.length > 0 && (
                  <div className="space-y-1.5">
                    <span className="font-semibold text-ok text-xs block">
                      Threats Resolved:
                    </span>
                    {comparison.resolvedFindings.map((f) => (
                      <div
                        key={f.id}
                        className="p-2 rounded bg-ok/5 border border-ok/15 flex items-center justify-between"
                      >
                        <div className="flex items-center gap-2">
                          <Badge severity={f.severity} />
                          <span className="font-medium text-text-primary">{f.title}</span>
                        </div>
                        <span className="font-mono text-[11px] text-text-secondary">{f.ruleId}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Historical Scans List */}
      <Card>
        <CardHeader>
          <CardTitle>Historical Threat Scans</CardTitle>
          <CardDescription>
            Archived evaluations stored in local SQLite database
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3 p-2">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : scans.length === 0 ? (
            <div className="py-12">
              <EmptyState
                title="No Reports Found"
                description={t('reports.noScans')}
              />
            </div>
          ) : (
            <div className="space-y-3">
              {scans.map((scan) => {
                const totalCount = Object.values(scan.findingsCount).reduce((a, b) => a + b, 0);

                return (
                  <div
                    key={scan.id}
                    className="p-3.5 rounded-md bg-raised/30 border border-separator/25 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs transition-colors hover:border-separator/40"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-text-primary font-mono text-xs">
                          {scan.id}
                        </span>
                        <span className="px-1.5 py-0.5 rounded text-[10px] text-text-secondary font-mono bg-raised uppercase border border-separator/20">
                          {scan.mode}
                        </span>
                        <Badge
                          severity={
                            scan.verdict === 'high_risk'
                              ? 'critical'
                              : scan.verdict === 'review'
                              ? 'medium'
                              : 'ok'
                          }
                        >
                          {scan.verdict.toUpperCase()}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2 text-[11px] text-text-secondary">
                        <Calendar className="w-3.5 h-3.5" />
                        <span>{scan.startedAt.slice(0, 19).replace('T', ' ')}</span>
                        <span>·</span>
                        <span className="font-medium text-text-primary">{totalCount} findings</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleExport('html', scan.id)}
                      >
                        <Download className="w-3.5 h-3.5" />
                        <span>{t('reports.exportHtml')}</span>
                      </Button>
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => handleExport('json', scan.id)}
                      >
                        <Download className="w-3.5 h-3.5" />
                        <span>{t('reports.exportJson')}</span>
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
