import React, { useState, useEffect } from 'react';
import {
  Radar,
  Play,
  Square,
  Activity,
  Bell,
  Sparkles,
  Search,
  Wrench,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Smartphone,
  AlertTriangle,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { EmptyState } from '../../components/ui/EmptyState';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import type { MonitorCandidate, MonitorEvent, MonitorSnapshot } from '../../lib/bridge/schema';

export const MonitorView: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeDevice, setActiveDevice, setActiveJob } = useDeviceStore();

  const [snapshot, setSnapshot] = useState<MonitorSnapshot | null>(null);
  const [candidates, setCandidates] = useState<MonitorCandidate[]>([]);
  const [events, setEvents] = useState<MonitorEvent[]>([]);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [expandedEventId, setExpandedEventId] = useState<string | null>(null);
  const [markerFeedback, setMarkerFeedback] = useState<string | null>(null);

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadSnapshot() {
      let dev = activeDevice;
      if (!dev) {
        try {
          dev = await bridge.device.getConnected();
          if (dev && isMounted) {
            setActiveDevice(dev);
          }
        } catch {
          dev = null;
        }
      }

      if (!dev || dev.connectionState !== 'ready') {
        if (isMounted) {
          setSnapshot(null);
          setIsMonitoring(false);
          setCandidates([]);
          setEvents([]);
        }
        return;
      }
      try {
        const snap = await bridge.monitor.snapshot();
        if (isMounted) {
          setSnapshot(snap);
          setIsMonitoring(Boolean(snap?.status?.isActive));
          setCandidates(snap?.candidates || []);
          setEvents(snap?.events || []);
        }
      } catch (err) {
        console.error('Failed to load monitor snapshot:', err);
      }
    }

    loadSnapshot();

    const unsubEvent = bridge.on('monitor:event', (evt) => {
      setEvents((prev) => [evt, ...prev.slice(0, 24)]);
    });

    const unsubCandidate = bridge.on('monitor:candidate', (cand) => {
      setCandidates((prev) => {
        const existing = prev.filter((c) => c.package !== cand.package);
        return [cand, ...existing];
      });
    });

    return () => {
      isMounted = false;
      unsubEvent();
      unsubCandidate();
    };
  }, [bridge, activeDevice, setActiveDevice]);

  const toggleMonitoring = async () => {
    try {
      if (isMonitoring) {
        await bridge.monitor.stop();
        setIsMonitoring(false);
        setActiveJob(null);
      } else {
        await bridge.monitor.start();
        setIsMonitoring(true);
        setActiveJob('Popup Hunter (Active)');
      }
    } catch (err) {
      console.error('Failed to toggle monitor:', err);
    }
  };

  const handleMarkPopup = async () => {
    try {
      await bridge.monitor.markPopup();
      setMarkerFeedback(t('monitor.markPopupSuccess'));

      const snap = await bridge.monitor.snapshot();
      setSnapshot(snap);
      setCandidates(snap?.candidates || []);
      setEvents(snap?.events || []);

      setTimeout(() => {
        setMarkerFeedback(null);
      }, 4000);
    } catch (err) {
      console.error('Failed to mark popup:', err);
    }
  };

  if (!activeDevice) {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <EmptyState
          icon={<Smartphone className="w-5 h-5 text-text-secondary" />}
          title={t('device.noDevice')}
          description={t('device.noDeviceHint')}
        />
      </div>
    );
  }

  if (activeDevice.connectionState === 'unauthorized') {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <EmptyState
          icon={<AlertTriangle className="w-5 h-5 text-warning" />}
          title={t('device.unauthorizedTitle')}
          description={`${t('device.unauthorizedPrompt')} ${t('device.unauthorizedAction')}`}
        />
      </div>
    );
  }

  if (activeDevice.connectionState === 'offline') {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <EmptyState
          icon={<AlertTriangle className="w-5 h-5 text-warning" />}
          title={t('device.offlineTitle')}
          description={`${t('device.offlinePrompt')} ${t('device.offlineTip')}`}
        />
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Header & Monitor Control */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg shadow-sm">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <Radar className={`w-6 h-6 ${isMonitoring ? 'animate-spin' : ''}`} />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('monitor.title')}
            </h2>
            <p className="text-xs text-text-secondary leading-relaxed">
              {t('monitor.description')}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2.5 w-full sm:w-auto">
          {isMonitoring && (
            <Button
              variant="secondary"
              size="md"
              onClick={handleMarkPopup}
              className="bg-sev-high/15 border-sev-high/30 text-sev-high hover:bg-sev-high/25"
            >
              <Sparkles className="w-4 h-4" />
              <span>{t('monitor.markPopup')}</span>
            </Button>
          )}

          <Button
            variant={isMonitoring ? 'destructive' : 'primary'}
            size="md"
            onClick={toggleMonitoring}
            disabled={!activeDevice || activeDevice.connectionState !== 'ready'}
            title={
              !activeDevice
                ? t('device.noDeviceHint')
                : activeDevice.connectionState !== 'ready'
                ? t('device.unauthorizedPrompt')
                : undefined
            }
          >
            {isMonitoring ? (
              <>
                <Square className="w-4 h-4 fill-current" />
                <span>{t('monitor.stop')}</span>
              </>
            ) : (
              <>
                <Play className="w-4 h-4 fill-current" />
                <span>{t('monitor.start')}</span>
              </>
            )}
          </Button>
        </div>
      </div>

      {markerFeedback && (
        <div className="flex items-center gap-2 p-3 rounded-md bg-ok/10 border border-ok/25 text-xs text-ok">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{markerFeedback}</span>
        </div>
      )}

      {/* Signal Sources Telemetry */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {[
          { key: 'window', label: 'Window Focus' },
          { key: 'logcat', label: 'Logcat Traces' },
          { key: 'overlay', label: 'AppOps Overlays' },
          { key: 'toast', label: 'Toast Broadcasts' },
        ].map((source) => {
          const isActive = isMonitoring && (
            snapshot?.status.sourcesActive.some((s) => s.toLowerCase().includes(source.key)) ?? true
          );

          return (
            <div
              key={source.key}
              className="p-3 bg-surface border border-separator/30 rounded-md flex items-center justify-between text-xs shadow-xs"
            >
              <span className="font-medium text-text-secondary">{source.label}</span>
              <div className="flex items-center gap-1.5">
                <span
                  className={`w-2 h-2 rounded-full ${
                    isActive ? 'bg-ok animate-pulse' : 'bg-separator/60'
                  }`}
                />
                <span className="text-[10px] uppercase font-mono text-text-secondary">
                  {isActive ? 'Active' : 'Idle'}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Candidates & Live Events Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* Ranked Adware Candidates */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-accent" />
              <CardTitle>{t('monitor.candidates')}</CardTitle>
            </div>
            <CardDescription>
              {t('monitor.candidatesDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {(candidates?.length ?? 0) === 0 ? (
              <div className="py-12 text-center text-xs text-text-secondary">
                {isMonitoring ? t('monitor.noEventsMonitoring') : t('monitor.noEvents')}
              </div>
            ) : (
              <div className="space-y-3">
                {(candidates || []).map((cand) => (
                  <div
                    key={cand.package}
                    className="p-3.5 rounded-md bg-raised/30 border border-separator/25 space-y-2 text-xs transition-colors hover:border-separator/40"
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-semibold text-text-primary text-xs">
                        {cand.appLabel || cand.package}
                      </span>
                      <Badge
                        severity={
                          cand.score >= 80 ? 'critical' : cand.score >= 50 ? 'high' : 'medium'
                        }
                      >
                        {t('monitor.candidateScore', { score: cand.score })}
                      </Badge>
                    </div>

                    <span className="font-mono text-[11px] text-text-secondary block">
                      {cand.package}
                    </span>

                    <p className="text-[11px] text-text-secondary leading-relaxed">
                      {cand.whyFlagged}
                    </p>

                    <div className="pt-2 border-t border-separator/15 flex items-center justify-end gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigate('/scan')}
                      >
                        <Search className="w-3.5 h-3.5 text-accent" />
                        <span>{t('monitor.deepScanCandidate')}</span>
                      </Button>

                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => navigate('/remediate')}
                      >
                        <Wrench className="w-3.5 h-3.5 text-sev-high" />
                        <span>{t('monitor.stageRemediation')}</span>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Real-time Event Stream */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bell className="w-4 h-4 text-accent" />
              <CardTitle>{t('monitor.triggerLog')}</CardTitle>
            </div>
            <CardDescription>
              {t('monitor.triggerLogDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {(events?.length ?? 0) === 0 ? (
              <div className="py-12 text-center text-xs text-text-secondary">
                {t('monitor.noEvents')}
              </div>
            ) : (
              <div className="space-y-2 max-h-[460px] overflow-y-auto pr-1">
                {(events || []).map((evt) => {
                  const isExpanded = expandedEventId === evt.id;

                  return (
                    <div
                      key={evt.id}
                      className="p-2.5 rounded-md bg-raised/20 border border-separator/20 text-xs transition-colors hover:bg-raised/35"
                    >
                      <div
                        className="flex items-center justify-between cursor-pointer select-none"
                        onClick={() => setExpandedEventId(isExpanded ? null : evt.id)}
                      >
                        <div className="flex items-center gap-2 truncate">
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-mono text-text-secondary uppercase bg-raised border border-separator/20">
                            {evt.type}
                          </span>
                          <span className="font-mono text-[11px] text-text-primary truncate font-medium">
                            {evt.package}
                          </span>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          <span className="text-[10px] text-text-secondary font-mono">
                            {(evt.timestamp || '').slice(11, 19)}
                          </span>
                          {isExpanded ? (
                            <ChevronUp className="w-3.5 h-3.5 text-text-secondary" />
                          ) : (
                            <ChevronDown className="w-3.5 h-3.5 text-text-secondary" />
                          )}
                        </div>
                      </div>

                      {isExpanded && (
                        <div className="mt-2 pt-2 border-t border-separator/15 space-y-1 text-[11px] font-mono text-text-secondary bg-base/60 p-2 rounded">
                          {Object.entries(evt.details || {}).map(([key, val]) => (
                            <div key={key} className="flex gap-2">
                              <span className="text-accent">{key}:</span>
                              <span className="text-text-primary break-all">{String(val)}</span>
                            </div>
                          ))}
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
    </div>
  );
};
