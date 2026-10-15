import React, { useState, useEffect, useMemo } from 'react';
import { ClipboardCheck, RotateCcw, AlertTriangle, Search, CheckCircle2, AlertOctagon, HelpCircle, ArrowRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { ErrorState } from '../../components/ui/ErrorState';
import { EmptyState } from '../../components/ui/EmptyState';
import { getBridge } from '../../lib/bridge';
import type { AuditItem } from '../../lib/bridge/schema';

export const AuditView: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [items, setItems] = useState<AuditItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');

  const bridge = getBridge();

  const handleManualRun = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const results = await bridge.audit.run();
      setItems(results || []);
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
          setItems(results || []);
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

  const filteredItems = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();

    return (items || []).filter((item) => {
      if (selectedCategory !== 'all' && item.category !== selectedCategory) return false;
      if (selectedStatus !== 'all' && item.status !== selectedStatus) return false;
      if (!q) return true;
      return (
        (item.title || '').toLowerCase().includes(q) ||
        (item.summary || '').toLowerCase().includes(q) ||
        (item.category || '').toLowerCase().includes(q)
      );
    });
  }, [items, searchQuery, selectedCategory, selectedStatus]);

  const categories = useMemo(() => {
    const set = new Set<string>();
    (items || []).forEach((item) => set.add(item.category));
    return ['all', ...Array.from(set)];
  }, [items]);

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
        return (
          <Badge severity="ok">
            <CheckCircle2 className="w-3 h-3 mr-1" />
            {t('audit.status.ok')}
          </Badge>
        );
      case 'warning':
        return (
          <Badge severity="medium">
            <AlertTriangle className="w-3 h-3 mr-1" />
            {t('audit.status.warning')}
          </Badge>
        );
      case 'issue':
        return (
          <Badge severity="critical">
            <AlertOctagon className="w-3 h-3 mr-1" />
            {t('audit.status.issue')}
          </Badge>
        );
      case 'unavailable':
      default:
        return (
          <Badge severity="info">
            <HelpCircle className="w-3 h-3 mr-1" />
            {t('audit.status.unavailable')}
          </Badge>
        );
    }
  };

  return (
    <div className="space-y-5">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg shadow-sm">
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
          variant="primary"
          size="md"
          onClick={handleManualRun}
          disabled={isLoading}
        >
          <RotateCcw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          <span>{isLoading ? t('audit.running') : t('audit.run')}</span>
        </Button>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-3 bg-surface border border-separator/35 rounded-lg text-xs">
        {/* Category Pills */}
        <div className="flex items-center gap-1.5 flex-wrap">
          {categories.map((cat) => (
            <button
              key={cat}
              type="button"
              onClick={() => setSelectedCategory(cat)}
              className={`px-2.5 py-1 rounded text-[11px] font-medium border transition-colors ${
                selectedCategory === cat
                  ? 'bg-accent text-accent-foreground border-accent'
                  : 'bg-raised text-text-secondary border-separator/30 hover:text-text-primary'
              }`}
            >
              {cat === 'all' ? t('audit.allCategories') : cat.toUpperCase()}
            </button>
          ))}
        </div>

        {/* Search */}
        <div className="relative w-full sm:w-64">
          <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-text-secondary" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('audit.searchPlaceholder')}
            className="w-full h-8 pl-8 pr-3 text-xs bg-raised rounded-md border border-separator/35 text-text-primary placeholder:text-text-secondary/60 focus:outline-none focus:ring-1 focus:ring-accent"
          />
        </div>
      </div>

      {/* Audit Items List */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>System Security Checklist</CardTitle>
              <CardDescription>
                {filteredItems?.length ?? 0} configuration policies verified
              </CardDescription>
            </div>

            {/* Status Filter Buttons */}
            <div className="flex items-center gap-1">
              {['all', 'issue', 'warning', 'ok'].map((st) => (
                <button
                  key={st}
                  type="button"
                  onClick={() => setSelectedStatus(st)}
                  className={`px-2 py-0.5 rounded text-[11px] font-medium border transition-colors ${
                    selectedStatus === st
                      ? 'bg-accent text-accent-foreground border-accent'
                      : 'bg-raised text-text-secondary border-separator/30 hover:text-text-primary'
                  }`}
                >
                  {st.toUpperCase()}
                </button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3 py-2">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : (filteredItems?.length ?? 0) === 0 ? (
            <div className="py-12">
              <EmptyState
                title="No Checklist Items"
                description={t('audit.noItems')}
                action={
                  searchQuery || selectedCategory !== 'all' || selectedStatus !== 'all' ? (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => {
                        setSearchQuery('');
                        setSelectedCategory('all');
                        setSelectedStatus('all');
                      }}
                    >
                      Reset Filters
                    </Button>
                  ) : undefined
                }
              />
            </div>
          ) : (
            <div className="space-y-3.5">
              {(filteredItems || []).map((item) => (
                <div
                  key={item.id}
                  className="p-3.5 rounded-md bg-raised/30 border border-separator/25 space-y-2 text-xs transition-colors hover:border-separator/40"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-text-primary text-xs">
                          {item.title}
                        </span>
                        <span className="px-1.5 py-0.5 rounded text-[10px] uppercase font-mono bg-raised text-text-secondary border border-separator/25">
                          {item.category}
                        </span>
                      </div>
                      <p className="text-text-secondary text-[11px] leading-relaxed">
                        {item.summary}
                      </p>
                    </div>

                    <div className="shrink-0">
                      {getStatusBadge(item.status)}
                    </div>
                  </div>

                  {/* Evidence traces */}
                  {item.evidence.length > 0 && (
                    <div className="space-y-1 pt-1 border-t border-separator/15">
                      <span className="text-[10px] text-text-secondary font-medium block">
                        {t('audit.evidence')}:
                      </span>
                      <div className="space-y-0.5">
                        {item.evidence.map((ev, idx) => (
                          <div
                            key={idx}
                            className="font-mono text-[11px] text-text-secondary bg-base/70 px-2 py-1 rounded border border-separator/20 break-all select-all"
                          >
                            {ev}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Remediation Action Suggestion */}
                  {item.remedy && item.status !== 'ok' && (
                    <div className="pt-2 border-t border-separator/15 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-[11px] bg-accent/5 p-2 rounded border border-accent/20">
                      <div className="flex items-center gap-1.5 text-accent">
                        <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                        <span>
                          {t('audit.remedy')}: {item.remedy}
                        </span>
                      </div>

                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => navigate('/remediate')}
                        className="shrink-0"
                      >
                        <span>{t('audit.fixAction')}</span>
                        <ArrowRight className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
