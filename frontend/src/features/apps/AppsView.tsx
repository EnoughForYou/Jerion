import React, { useEffect, useState, useMemo } from 'react';
import { Search, ShieldAlert, PackageCheck } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card } from '../../components/ui/Card';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../components/ui/Table';
import { Button } from '../../components/ui/Button';
import { Skeleton } from '../../components/ui/Skeleton';
import { EmptyState } from '../../components/ui/EmptyState';
import { ErrorState } from '../../components/ui/ErrorState';
import { DetailPanel } from '../../app/DetailPanel';
import { getBridge } from '../../lib/bridge';
import type { AppInfo } from '../../lib/bridge/schema';

export const AppsView: React.FC = () => {
  const { t } = useTranslation();
  const [apps, setApps] = useState<AppInfo[]>([]);
  const [selectedApp, setSelectedApp] = useState<AppInfo | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [showSystemApps, setShowSystemApps] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadApps() {
      try {
        setIsLoading(true);
        setError(null);
        const list = await bridge.inventory.list();
        if (isMounted) {
          setApps(list);
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
    return apps.filter((app) => {
      if (!showSystemApps && app.isSystem) {
        return false;
      }
      if (!searchQuery) return true;
      const query = searchQuery.toLowerCase();
      return (
        app.package.toLowerCase().includes(query) ||
        (app.label && app.label.toLowerCase().includes(query))
      );
    });
  }, [apps, searchQuery, showSystemApps]);

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
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-4 bg-surface border border-separator/35 rounded-lg">
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('apps.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('apps.description')} ({filteredApps.length} packages listed)
            </p>
          </div>

          <div className="flex items-center gap-2.5 w-full sm:w-auto">
            {/* Search Input */}
            <div className="relative flex-1 sm:w-64">
              <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-text-secondary" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('apps.searchPlaceholder')}
                className="w-full h-8 pl-8 pr-3 text-xs bg-raised rounded-md border border-separator/35 text-text-primary placeholder:text-text-secondary/60 focus:outline-none focus:ring-1 focus:ring-accent"
              />
            </div>

            {/* Toggle System Packages */}
            <label className="flex items-center gap-1.5 text-xs text-text-secondary cursor-pointer select-none">
              <input
                type="checkbox"
                checked={showSystemApps}
                onChange={(e) => setShowSystemApps(e.target.checked)}
                className="rounded-xs border-separator text-accent focus:ring-0"
              />
              <span>System</span>
            </label>
          </div>
        </div>

        {/* Apps Table */}
        <Card className="flex-1 p-0 overflow-hidden flex flex-col">
          {isLoading ? (
            <div className="p-4 space-y-3">
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
            </div>
          ) : filteredApps.length === 0 ? (
            <div className="p-8">
              <EmptyState
                title="No Applications Found"
                description={
                  searchQuery
                    ? `No packages match "${searchQuery}". Check spelling or include system apps.`
                    : 'No packages detected in inventory.'
                }
                action={
                  searchQuery ? (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => setSearchQuery('')}
                    >
                      Reset Filter
                    </Button>
                  ) : undefined
                }
              />
            </div>
          ) : (
            <div className="flex-1 overflow-y-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('apps.columns.app')}</TableHead>
                    <TableHead>{t('apps.columns.version')}</TableHead>
                    <TableHead>{t('apps.columns.installer')}</TableHead>
                    <TableHead>{t('apps.columns.flags')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredApps.map((app) => (
                    <TableRow
                      key={app.package}
                      className="cursor-pointer"
                      onClick={() => setSelectedApp(app)}
                    >
                      <TableCell>
                        <div className="flex flex-col">
                          <span className="font-medium text-text-primary">
                            {app.label || app.package}
                          </span>
                          <span className="font-mono text-[11px] text-text-secondary truncate">
                            {app.package}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {app.versionName || '-'}
                      </TableCell>
                      <TableCell className="text-xs text-text-secondary truncate">
                        {app.installerPackage || 'Sideload / Unknown'}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          {app.isSystem ? (
                            <span className="px-1.5 py-0.5 rounded-sm bg-raised text-[10px] text-text-secondary border border-separator/30">
                              System
                            </span>
                          ) : (
                            <span className="px-1.5 py-0.5 rounded-sm bg-accent/15 text-[10px] text-accent font-medium">
                              User
                            </span>
                          )}
                          {app.hasOverlayPermission ? (
                            <span className="px-1.5 py-0.5 rounded-sm bg-sev-high/15 text-[10px] text-sev-high font-medium">
                              Overlay
                            </span>
                          ) : null}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </Card>
      </div>

      {/* 420px Detail Slide-in Panel */}
      <DetailPanel
        open={Boolean(selectedApp)}
        onClose={() => setSelectedApp(null)}
        title={selectedApp?.label || selectedApp?.package || 'App Details'}
        description={selectedApp?.package}
      >
        {selectedApp ? (
          <div className="space-y-4 text-xs">
            {/* General Specs */}
            <div className="space-y-2 border-b border-separator/20 pb-3">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-text-secondary block">Version Code:</span>
                  <span className="font-mono text-text-primary">{selectedApp.versionCode}</span>
                </div>
                <div>
                  <span className="text-text-secondary block">Target SDK:</span>
                  <span className="text-text-primary">API {selectedApp.targetSdk}</span>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-text-secondary block">First Installed:</span>
                  <span className="text-text-primary">{selectedApp.installedAt.slice(0, 10)}</span>
                </div>
                <div>
                  <span className="text-text-secondary block">Last Updated:</span>
                  <span className="text-text-primary">{selectedApp.updatedAt.slice(0, 10)}</span>
                </div>
              </div>
            </div>

            {/* Requested Permissions */}
            <div>
              <h4 className="font-semibold text-text-primary mb-2 flex items-center gap-1.5">
                <ShieldAlert className="w-3.5 h-3.5 text-accent" />
                Requested Permissions ({selectedApp.permissions.length})
              </h4>
              <div className="max-h-48 overflow-y-auto space-y-1 bg-raised/30 p-2 rounded-md border border-separator/25 font-mono text-[11px]">
                {selectedApp.permissions.length === 0 ? (
                  <span className="text-text-secondary">No declared permissions.</span>
                ) : (
                  selectedApp.permissions.map((perm) => (
                    <div key={perm} className="text-text-secondary truncate">
                      {perm}
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Signatures / Certs */}
            <div>
              <h4 className="font-semibold text-text-primary mb-2 flex items-center gap-1.5">
                <PackageCheck className="w-3.5 h-3.5 text-accent" />
                Signatures ({selectedApp.signatures.length})
              </h4>
              <div className="space-y-1 bg-raised/30 p-2 rounded-md border border-separator/25 font-mono text-[11px]">
                {selectedApp.signatures.map((sig, idx) => (
                  <div key={idx} className="break-all text-text-secondary">
                    {sig}
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : null}
      </DetailPanel>
    </div>
  );
};
