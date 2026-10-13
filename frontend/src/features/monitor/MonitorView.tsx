import React, { useState, useEffect } from 'react';
import { Radar, Play, Square, Activity, Bell } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { getBridge } from '../../lib/bridge';
import { useDeviceStore } from '../../app/deviceStore';
import type { MonitorCandidate, MonitorEvent, MonitorSnapshot } from '../../lib/bridge/schema';

export const MonitorView: React.FC = () => {
  const { t } = useTranslation();
  const { setActiveJob } = useDeviceStore();

  const [snapshot, setSnapshot] = useState<MonitorSnapshot | null>(null);
  const [candidates, setCandidates] = useState<MonitorCandidate[]>([]);
  const [events, setEvents] = useState<MonitorEvent[]>([]);
  const [isMonitoring, setIsMonitoring] = useState(false);

  const bridge = getBridge();

  useEffect(() => {
    let isMounted = true;

    async function loadSnapshot() {
      try {
        const snap = await bridge.monitor.snapshot();
        if (isMounted) {
          setSnapshot(snap);
          setIsMonitoring(snap.status.isActive);
          setCandidates(snap.candidates);
          setEvents(snap.events);
        }
      } catch (err) {
        console.error('Failed to load monitor snapshot:', err);
      }
    }

    loadSnapshot();

    const unsubEvent = bridge.on('monitor:event', (evt) => {
      setEvents((prev) => [evt, ...prev.slice(0, 19)]);
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
  }, [bridge]);

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

  return (
    <div className="space-y-5">
      {/* Header & Monitor Control */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-surface border border-separator/35 rounded-lg">
        <div className="flex items-center gap-3.5">
          <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
            <Radar className={`w-6 h-6 ${isMonitoring ? 'animate-spin' : ''}`} />
          </div>
          <div>
            <h2 className="text-md font-semibold text-text-primary">
              {t('monitor.title')}
            </h2>
            <p className="text-xs text-text-secondary">
              {t('monitor.description')}
            </p>
          </div>
        </div>

        <Button
          variant={isMonitoring ? 'destructive' : 'primary'}
          size="md"
          onClick={toggleMonitoring}
        >
          {isMonitoring ? (
            <>
              <Square className="w-4 h-4 fill-current" />
              {t('monitor.stop')}
            </>
          ) : (
            <>
              <Play className="w-4 h-4 fill-current" />
              {t('monitor.start')}
            </>
          )}
        </Button>
      </div>

      {/* Signal Sources telemetry */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {['appops', 'toast', 'window', 'overlay'].map((sigKey) => {
          const isActive = snapshot?.status.sourcesActive.includes(sigKey);
          return (
            <div
              key={sigKey}
              className="p-3 bg-surface border border-separator/30 rounded-md flex items-center justify-between text-xs"
            >
              <span className="font-medium text-text-secondary capitalize">{sigKey}</span>
              <span
                className={`w-2 h-2 rounded-full ${
                  isActive ? 'bg-ok shadow-xs' : 'bg-separator'
                }`}
              />
            </div>
          );
        })}
      </div>

      {/* Candidates & Live Events Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* Ranked Adware Candidates */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-accent" />
              <CardTitle>{t('monitor.candidates')}</CardTitle>
            </div>
            <CardDescription>
              Packages correlated with popup triggers and fullscreen overlays
            </CardDescription>
          </CardHeader>
          <CardContent>
            {candidates.length === 0 ? (
              <div className="py-8 text-center text-xs text-text-secondary">
                {t('monitor.noEvents')}
              </div>
            ) : (
              <div className="space-y-3">
                {candidates.map((cand) => (
                  <div
                    key={cand.package}
                    className="p-3 rounded-md bg-raised/30 border border-separator/25 space-y-1.5 text-xs"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-text-primary">
                        {cand.appLabel || cand.package}
                      </span>
                      <Badge
                        severity={cand.score >= 80 ? 'critical' : cand.score >= 50 ? 'high' : 'medium'}
                      >
                        Risk {cand.score}/100
                      </Badge>
                    </div>
                    <span className="font-mono text-[11px] text-text-secondary block">
                      {cand.package}
                    </span>
                    <p className="text-[11px] text-text-secondary leading-relaxed">
                      {cand.whyFlagged}
                    </p>
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
              <CardTitle>Trigger Log</CardTitle>
            </div>
            <CardDescription>
              Recent foreground shifts, toast broadcasts, and window overlays
            </CardDescription>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <div className="py-8 text-center text-xs text-text-secondary">
                No events recorded. Start monitoring to capture foreground activity.
              </div>
            ) : (
              <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
                {events.map((evt) => (
                  <div
                    key={evt.id}
                    className="p-2 rounded-sm bg-raised/20 border border-separator/20 text-xs flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2 truncate">
                      <span className="px-1.5 py-0.5 rounded-xs bg-raised text-[10px] font-mono text-text-secondary uppercase">
                        {evt.type}
                      </span>
                      <span className="font-mono text-[11px] text-text-primary truncate">
                        {evt.package}
                      </span>
                    </div>
                    <span className="text-[10px] text-text-secondary shrink-0 font-mono">
                      {evt.timestamp.slice(11, 19)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
