import React, { useEffect, useState, useMemo, useRef } from 'react';
import { Search, ShieldAlert, ArrowUpDown } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { EmptyState } from '../../components/ui/EmptyState';
import { ErrorState } from '../../components/ui/ErrorState';
import { AppDetailDrawer } from './AppDetailDrawer';
import { getBridge } from '../../lib/bridge';
import type { AppInfo } from '../../lib/bridge/schema';

type AppCategoryFilter = 'all' | 'user' | 'system' | 'risky';
type SortField = 'name' | 'installedAt' | 'updatedAt' | 'targetSdk';

export const AppsView: React.FC = () => {
  const { t } = useTranslation();
  const [apps, setApps] = useState<AppInfo[]>([]);
  const [selectedApp, setSelectedApp] = useState<AppInfo | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState<AppCategoryFilter>('all');
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortAsc, setSortAsc] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const parentRef = useRef<HTMLDivElement>(null);
  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadApps() {
      try {
        setIsLoading(true);
        setError(null);
        const list = await bridge.inventory.list();
        if (isMounted) {
          setApps(list || []);
        }
      } catch (err: unknown) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Failed to retrieve package inventory');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadApps();

    return () => {
      isMounted = false;
    };
  }, [bridge]);

  const filteredApps = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();

    return (apps || [])
      .filter((app) => {
        // Category filtering
        if (category === 'user' && app.isSystem) return false;
        if (category === 'system' && !app.isSystem) return false;
        if (category === 'risky' && (app.suspiciousFlags?.length ?? 0) === 0 && !app.hasOverlayPermission) {
          return false;
        }

        // Search filtering
        if (!query) return true;
        return (
          (app.package || '').toLowerCase().includes(query) ||
          (app.label || '').toLowerCase().includes(query)
        );
      })
      .sort((a, b) => {
        let cmp = 0;
        if (sortField === 'name') {
          cmp = (a.label || a.package).localeCompare(b.label || b.package);
        } else if (sortField === 'installedAt') {
          cmp = (a.installedAt || '').localeCompare(b.installedAt || '');
        } else if (sortField === 'updatedAt') {
          cmp = (a.updatedAt || '').localeCompare(b.updatedAt || '');
        } else if (sortField === 'targetSdk') {
          cmp = (a.targetSdk || 0) - (b.targetSdk || 0);
        }
        return sortAsc ? cmp : -cmp;
      });
  }, [apps, category, searchQuery, sortField, sortAsc]);

  // eslint-disable-next-line react-hooks/incompatible-library
  const rowVirtualizer = useVirtualizer({
    count: filteredApps.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 54,
    overscan: 8,
    initialRect: { width: 800, height: 600 },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();
  const itemsToRender =
    virtualItems.length > 0
      ? virtualItems.map((v) => ({ index: v.index, start: v.start, size: v.size, app: filteredApps[v.index] }))
      : filteredApps.map((app, index) => ({
          index,
          start: index * 54,
          size: 54,
          app,
        }));

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortAsc((prev) => !prev);
    } else {
      setSortField(field);
      setSortAsc(true);
    }
  };

  if (error) {
    return (
      <ErrorState
        title="Failed to Load Installed Applications"
        remedy="Ensure ADB package manager service is reachable and not restricted."
        onRetry={() => window.location.reload()}
      />
    );
  }

  return (
    <div className="flex h-full gap-4">
      {/* Main Table Area */}
      <div className="flex-1 flex flex-col min-w-0 space-y-4">
        {/* Header & Filter Controls */}
        <div className="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-3 p-4 bg-surface border border-separator/35 rounded-lg shadow-sm">
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('apps.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('apps.packagesCount', { count: filteredApps.length })}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2.5 w-full lg:w-auto">
            {/* Category Filter Pills */}
            <div className="flex items-center gap-1 bg-raised/50 p-1 rounded-md border border-separator/20">
              {(['all', 'user', 'system', 'risky'] as AppCategoryFilter[]).map((cat) => (
                <button
                  key={cat}
                  type="button"
                  onClick={() => setCategory(cat)}
                  className={`px-2.5 py-1 rounded text-[11px] font-medium transition-colors ${
                    category === cat
                      ? 'bg-accent text-accent-foreground'
                      : 'text-text-secondary hover:text-text-primary'
                  }`}
                >
                  {t(`apps.categories.${cat}`)}
                </button>
              ))}
            </div>

            {/* Search Input */}
            <div className="relative flex-1 sm:w-60 min-w-[180px]">
              <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-text-secondary" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('apps.searchPlaceholder')}
                className="w-full h-8 pl-8 pr-3 text-xs bg-raised rounded-md border border-separator/35 text-text-primary placeholder:text-text-secondary/60 focus:outline-none focus:ring-1 focus:ring-accent"
              />
            </div>
          </div>
        </div>

        {/* Virtualized Apps Table */}
        <Card className="flex-1 p-0 overflow-hidden flex flex-col min-h-[420px]">
          {/* Fixed Table Header */}
          <div className="grid grid-cols-12 gap-2 px-4 py-2.5 border-b border-separator/25 bg-raised/30 text-xs font-semibold text-text-secondary select-none">
            <div
              className="col-span-5 flex items-center gap-1 cursor-pointer hover:text-text-primary"
              onClick={() => toggleSort('name')}
            >
              <span>{t('apps.columns.app')}</span>
              <ArrowUpDown className="w-3 h-3" />
            </div>
            <div
              className="col-span-2 flex items-center gap-1 cursor-pointer hover:text-text-primary"
              onClick={() => toggleSort('targetSdk')}
            >
              <span>{t('apps.columns.version')}</span>
              <ArrowUpDown className="w-3 h-3" />
            </div>
            <div className="col-span-3">
              <span>{t('apps.columns.installer')}</span>
            </div>
            <div className="col-span-2 text-right">
              <span>{t('apps.columns.flags')}</span>
            </div>
          </div>

          {isLoading ? (
            <div className="p-4 space-y-3">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : filteredApps.length === 0 ? (
            <div className="p-8 my-auto">
              <EmptyState
                title={t('apps.noAppsTitle')}
                description={
                  searchQuery
                    ? t('apps.noAppsDesc', { query: searchQuery })
                    : t('apps.noAppsEmpty')
                }
                action={
                  searchQuery ? (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => setSearchQuery('')}
                    >
                      {t('apps.resetFilter')}
                    </Button>
                  ) : undefined
                }
              />
            </div>
          ) : (
            <div
              ref={parentRef}
              className="flex-1 overflow-y-auto relative w-full h-[500px]"
            >
              <div
                style={{
                  height: `${rowVirtualizer.getTotalSize()}px`,
                  width: '100%',
                  position: 'relative',
                }}
              >
                {itemsToRender.map((virtualRow) => {
                  const app = virtualRow.app;
                  const isSelected = selectedApp?.package === app.package;

                  return (
                    <div
                      key={app.package}
                      data-index={virtualRow.index}
                      onClick={() => setSelectedApp(app)}
                      style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: `${virtualRow.size}px`,
                        transform: `translateY(${virtualRow.start}px)`,
                      }}
                      className={`grid grid-cols-12 gap-2 px-4 items-center text-xs border-b border-separator/15 transition-colors cursor-pointer select-none ${
                        isSelected
                          ? 'bg-accent/10 border-accent/30'
                          : 'hover:bg-raised/40'
                      }`}
                    >
                      {/* App Name & Package */}
                      <div className="col-span-5 min-w-0 pr-2">
                        <div className="font-medium text-text-primary truncate">
                          {app.label || app.package}
                        </div>
                        <div className="font-mono text-[11px] text-text-secondary truncate">
                          {app.package}
                        </div>
                      </div>

                      {/* Version & Target SDK */}
                      <div className="col-span-2 min-w-0 font-mono text-[11px] text-text-secondary">
                        <div className="truncate text-text-primary">
                          {app.versionName || '-'}
                        </div>
                        <div className="text-[10px]">API {app.targetSdk}</div>
                      </div>

                      {/* Installer */}
                      <div className="col-span-3 min-w-0 text-[11px] text-text-secondary truncate">
                        {app.installerPackage || t('apps.detail.installerUnknown')}
                      </div>

                      {/* Attributes / Badges */}
                      <div className="col-span-2 flex items-center justify-end gap-1 flex-wrap">
                        {app.isSystem ? (
                          <span className="px-1.5 py-0.5 rounded text-[10px] bg-raised text-text-secondary border border-separator/30">
                            {t('apps.flags.system')}
                          </span>
                        ) : (
                          <span className="px-1.5 py-0.5 rounded text-[10px] bg-accent/15 text-accent font-medium">
                            {t('apps.flags.user')}
                          </span>
                        )}

                        {app.hasOverlayPermission && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] bg-sev-high/15 text-sev-high font-medium">
                            {t('apps.flags.overlay')}
                          </span>
                        )}

                        {(app.suspiciousFlags?.length ?? 0) > 0 && (
                          <span className="p-0.5 text-sev-medium" title="Suspicious Flags Detected">
                            <ShieldAlert className="w-3.5 h-3.5" />
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </Card>
      </div>

      {/* Slide-in Detail Drawer */}
      <AppDetailDrawer
        app={selectedApp}
        onClose={() => setSelectedApp(null)}
      />
    </div>
  );
};
