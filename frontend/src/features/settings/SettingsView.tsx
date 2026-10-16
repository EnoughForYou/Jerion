import React, { useState, useEffect } from 'react';
import {
  Settings as SettingsIcon,
  Moon,
  Sun,
  Laptop,
  Shield,
  Layers,
  Plus,
  Trash2,
  CheckCircle2,
  Terminal,
  Server,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { useThemeStore } from '../../app/theme';
import { getBridge, bridgeKind } from '../../lib/bridge';
import type { AllowlistEntry, AllowlistType, Settings } from '../../lib/bridge/schema';

export const SettingsView: React.FC = () => {
  const { t } = useTranslation();
  const { themeMode, setThemeMode } = useThemeStore();
  const [allowlist, setAllowlist] = useState<AllowlistEntry[]>([]);
  const [settings, setSettings] = useState<Settings>({
    theme: 'dark',
    adbPath: 'C:\\platform-tools\\adb.exe',
    customAdbPort: 5037,
    deepScanTimeoutSec: 120,
    autoCheckPreflight: true,
  });
  const [adbPathInput, setAdbPathInput] = useState('');
  const [clamavEnabled, setClamavEnabled] = useState(false);
  const [clamavHost, setClamavHost] = useState('127.0.0.1:3310');
  const [clamavStatus, setClamavStatus] = useState<{ testing: boolean; connected?: boolean; version?: string; error?: string } | null>(null);
  const [newTarget, setNewTarget] = useState('');
  const [newType, setNewType] = useState<AllowlistType>('package');
  const [newReason, setNewReason] = useState('');
  const [feedback, setFeedback] = useState<string | null>(null);

  const bridge = getBridge();

  useEffect(() => {
    async function loadData() {
      try {
        const [loadedSettings, list] = await Promise.all([
          bridge.settings.get(),
          bridge.allowlist.list(),
        ]);
        setSettings(loadedSettings);
        setAdbPathInput(loadedSettings.adbPath || '');
        setAllowlist(list || []);

        const goApp = (window as unknown as { go?: { main?: { App?: { ClamAVGetConfig?: () => Promise<{ enabled: boolean; host: string }> } } } })?.go?.main?.App;
        if (goApp?.ClamAVGetConfig) {
          const cfg = await goApp.ClamAVGetConfig();
          setClamavEnabled(cfg.enabled);
          setClamavHost(cfg.host || '127.0.0.1:3310');
        }
      } catch (err) {
        console.error('Failed to load settings:', err);
      }
    }
    loadData();
  }, [bridge]);

  const handleTestClamAV = async () => {
    setClamavStatus({ testing: true });
    try {
      const goApp = (window as unknown as { go?: { main?: { App?: { ClamAVTestConnection?: (h: string) => Promise<{ connected: boolean; version?: string; error?: string }> } } } })?.go?.main?.App;
      if (goApp?.ClamAVTestConnection) {
        const res = await goApp.ClamAVTestConnection(clamavHost.trim() || '127.0.0.1:3310');
        setClamavStatus({ testing: false, connected: res.connected, version: res.version, error: res.error });
      } else {
        setClamavStatus({ testing: false, connected: true, version: 'ClamAV 1.4.1 (Simulated)' });
      }
    } catch (err: unknown) {
      setClamavStatus({ testing: false, connected: false, error: err instanceof Error ? err.message : 'Unknown error' });
    }
  };

  const handleSaveClamAV = async () => {
    try {
      const goApp = (window as unknown as { go?: { main?: { App?: { ClamAVSaveConfig?: (e: boolean, h: string) => Promise<void> } } } })?.go?.main?.App;
      if (goApp?.ClamAVSaveConfig) {
        await goApp.ClamAVSaveConfig(clamavEnabled, clamavHost.trim() || '127.0.0.1:3310');
      }
      setFeedback(t('settings.clamavSaved'));
      setTimeout(() => setFeedback(null), 4000);
    } catch (err) {
      console.error('Failed to save ClamAV config:', err);
    }
  };

  const handleSaveAdbPath = async () => {
    try {
      if (adbPathInput.trim()) {
        await bridge.preflight.selectAdbPath(adbPathInput.trim());
        const updated = await bridge.settings.update({ adbPath: adbPathInput.trim() });
        setSettings(updated);
        setFeedback(t('settings.pathSaved'));
        setTimeout(() => setFeedback(null), 4000);
      }
    } catch (err) {
      console.error('Failed to save ADB path:', err);
    }
  };

  const handleToggleAutoCheck = async (checked: boolean) => {
    try {
      const updated = await bridge.settings.update({ autoCheckPreflight: checked });
      setSettings(updated);
    } catch (err) {
      console.error('Failed to update auto-check:', err);
    }
  };

  const handleAddAllowlist = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTarget.trim()) return;

    try {
      const created = await bridge.allowlist.add({
        type: newType,
        target: newTarget.trim(),
        reason: newReason.trim() || undefined,
      });
      setAllowlist((prev) => [...prev, created]);
      setNewTarget('');
      setNewReason('');
      setFeedback('Exclusion added to allowlist.');
      setTimeout(() => setFeedback(null), 4000);
    } catch (err) {
      console.error('Failed to add to allowlist:', err);
    }
  };

  const handleRemoveAllowlist = async (id: string) => {
    try {
      await bridge.allowlist.remove(id);
      setAllowlist((prev) => prev.filter((item) => item.id !== id));
      setFeedback('Exclusion removed from allowlist.');
      setTimeout(() => setFeedback(null), 4000);
    } catch (err) {
      console.error('Failed to remove allowlist entry:', err);
    }
  };

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center gap-3.5 p-5 bg-surface border border-separator-subtle rounded-lg shadow-sm">
        <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
          <SettingsIcon className="w-6 h-6" />
        </div>
        <div>
          <h2 className="text-md font-semibold text-text-primary">
            {t('settings.title')}
          </h2>
          <p className="text-xs text-text-secondary leading-relaxed">
            {t('settings.general')}
          </p>
        </div>
      </div>

      {feedback && (
        <div className="flex items-center gap-2 p-3 rounded-md bg-ok/10 border border-ok/25 text-xs text-ok">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{feedback}</span>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* Appearance & Theme Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Sun className="w-4 h-4 text-accent" />
              <CardTitle>{t('settings.theme')}</CardTitle>
            </div>
            <CardDescription>{t('settings.themeDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap items-center gap-2.5">
              <Button
                variant={themeMode === 'system' ? 'primary' : 'secondary'}
                size="md"
                onClick={() => setThemeMode('system')}
              >
                <Laptop className="w-4 h-4" />
                <span>{t('settings.themeSystem')}</span>
              </Button>
              <Button
                variant={themeMode === 'dark' ? 'primary' : 'secondary'}
                size="md"
                onClick={() => setThemeMode('dark')}
              >
                <Moon className="w-4 h-4" />
                <span>{t('settings.themeDark')}</span>
              </Button>
              <Button
                variant={themeMode === 'light' ? 'primary' : 'secondary'}
                size="md"
                onClick={() => setThemeMode('light')}
              >
                <Sun className="w-4 h-4" />
                <span>{t('settings.themeLight')}</span>
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Diagnostics & Rules Meta */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-accent" />
              <CardTitle>{t('settings.diagnostics', 'Diagnostics & Rules')}</CardTitle>
            </div>
            <CardDescription>
              {t('settings.diagnosticsDesc', 'System specifications and threat definitions')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <dl className="space-y-2 text-xs">
              {import.meta.env.DEV && (
                <div className="flex justify-between py-1 border-b border-separator-muted">
                  <dt className="text-text-secondary">Active Bridge</dt>
                  <dd className="font-mono text-text-primary font-semibold">
                    {bridgeKind === 'wails' ? 'WAILS (Backend)' : `${bridgeKind.toUpperCase()} (Simulated)`}
                  </dd>
                </div>
              )}
              <div className="flex justify-between py-1 border-b border-separator-muted">
                <dt className="text-text-secondary">App Version</dt>
                <dd className="font-mono text-text-primary">v1.1.0</dd>
              </div>
              <div className="flex justify-between py-1 border-b border-separator-muted">
                <dt className="text-text-secondary">{t('settings.rules')}</dt>
                <dd className="font-mono text-text-primary">v2026.03.1 (32 rules)</dd>
              </div>
              <div className="flex justify-between py-1">
                <dt className="text-text-secondary">Contract Schema</dt>
                <dd className="font-mono text-text-primary">v1.0.0 (Frozen)</dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </div>

      {/* ADB Engine Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-accent" />
            <CardTitle>{t('settings.adbConfig')}</CardTitle>
          </div>
          <CardDescription>{t('settings.adbConfigDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-xs">
          {/* Custom ADB Path input */}
          <div className="space-y-1.5">
            <label className="font-medium text-text-primary block">
              {t('settings.adbPath')}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={adbPathInput}
                onChange={(e) => setAdbPathInput(e.target.value)}
                placeholder={t('settings.adbPathPlaceholder')}
                className="flex-1 h-9 px-3 text-xs font-mono bg-raised rounded-md border border-separator-subtle text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
              />
              <Button variant="secondary" size="md" onClick={handleSaveAdbPath}>
                {t('settings.saveAdbPath')}
              </Button>
            </div>
          </div>

          <div className="pt-2 border-t border-separator-muted flex items-center justify-between">
            <div>
              <span className="font-medium text-text-primary block">
                {t('settings.autoCheck')}
              </span>
              <span className="text-[11px] text-text-secondary">
                Verifies ADB binary and daemon prerequisites every time Jerion starts.
              </span>
            </div>

            <input
              type="checkbox"
              checked={settings.autoCheckPreflight}
              onChange={(e) => handleToggleAutoCheck(e.target.checked)}
              className="rounded border-separator text-accent focus:ring-accent h-4 w-4"
            />
          </div>
        </CardContent>
      </Card>

      {/* ClamAV Local Antivirus Engine Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Server className="w-4 h-4 text-accent" />
            <CardTitle>{t('settings.clamavTitle')}</CardTitle>
          </div>
          <CardDescription>{t('settings.clamavDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-xs">
          {/* Enable ClamAV toggle */}
          <div className="flex items-center justify-between">
            <div>
              <span className="font-medium text-text-primary block">
                {t('settings.clamavEnabled')}
              </span>
              <span className="text-[11px] text-text-secondary">
                Stream APK payloads to local clamd daemon (TCP) for offline signature-based threat detection.
              </span>
            </div>
            <input
              type="checkbox"
              checked={clamavEnabled}
              onChange={(e) => setClamavEnabled(e.target.checked)}
              className="rounded border-separator text-accent focus:ring-accent h-4 w-4"
            />
          </div>

          {/* ClamAV Host & Port Input */}
          <div className="space-y-1.5 pt-2 border-t border-separator-muted">
            <label className="font-medium text-text-primary block">
              {t('settings.clamavHost')}
            </label>
            <div className="flex flex-col sm:flex-row gap-2">
              <input
                type="text"
                value={clamavHost}
                onChange={(e) => setClamavHost(e.target.value)}
                placeholder={t('settings.clamavHostPlaceholder')}
                className="flex-1 h-9 px-3 text-xs font-mono bg-raised rounded-md border border-separator-subtle text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
              />
              <Button
                variant="outline"
                size="md"
                onClick={handleTestClamAV}
                disabled={clamavStatus?.testing}
              >
                {clamavStatus?.testing ? t('settings.clamavTesting') : t('settings.clamavTest')}
              </Button>
              <Button
                variant="secondary"
                size="md"
                onClick={handleSaveClamAV}
              >
                {t('settings.clamavSave')}
              </Button>
            </div>

            {/* Test Connection Result Feedback */}
            {clamavStatus && !clamavStatus.testing && (
              <div
                className={`mt-2 p-2.5 rounded text-[11px] font-mono flex items-center gap-2 ${
                  clamavStatus.connected
                    ? 'bg-ok/10 text-ok border border-ok/30'
                    : 'bg-sev-critical/10 text-sev-critical border border-sev-critical/30'
                }`}
              >
                {clamavStatus.connected ? (
                  <>
                    <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />
                    <span>{t('settings.clamavConnected', { version: clamavStatus.version })}</span>
                  </>
                ) : (
                  <span>{t('settings.clamavError', { error: clamavStatus.error })}</span>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Package Allowlist & Exclusions */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Layers className="w-4 h-4 text-accent" />
            <CardTitle>{t('settings.allowlistTitle')}</CardTitle>
          </div>
          <CardDescription>{t('settings.allowlistDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Add Exclusion Form */}
          <form
            onSubmit={handleAddAllowlist}
            className="p-3 rounded-md bg-raised/30 border border-separator-subtle space-y-2.5 text-xs"
          >
            <span className="font-semibold text-text-primary text-[11px] block">
              {t('settings.addAllowlist')}
            </span>

            <div className="grid grid-cols-1 sm:grid-cols-12 gap-2">
              <input
                type="text"
                value={newTarget}
                onChange={(e) => setNewTarget(e.target.value)}
                placeholder={t('settings.targetPlaceholder')}
                className="sm:col-span-6 h-8 px-2.5 font-mono text-[11px] bg-raised rounded border border-separator-subtle text-text-primary placeholder:text-text-secondary/60 focus:outline-none focus:ring-1 focus:ring-accent"
              />

              <select
                value={newType}
                onChange={(e) => setNewType(e.target.value as AllowlistType)}
                className="sm:col-span-3 h-8 px-2 text-[11px] bg-raised rounded border border-separator-subtle text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
              >
                <option value="package">Package</option>
                <option value="rule_package">Rule + Package</option>
              </select>

              <Button
                type="submit"
                variant="primary"
                size="sm"
                className="sm:col-span-3"
                disabled={!newTarget.trim()}
              >
                <Plus className="w-3.5 h-3.5 mr-1" />
                <span>{t('settings.add')}</span>
              </Button>
            </div>

            <input
              type="text"
              value={newReason}
              onChange={(e) => setNewReason(e.target.value)}
              placeholder={t('settings.reasonPlaceholder')}
              className="w-full h-8 px-2.5 text-[11px] bg-raised rounded border border-separator-subtle text-text-primary placeholder:text-text-secondary/60 focus:outline-none focus:ring-1 focus:ring-accent"
            />
          </form>

          {/* List of Entries */}
          {(allowlist?.length ?? 0) === 0 ? (
            <p className="text-xs text-text-secondary py-4 text-center">
              {t('settings.noAllowlist')}
            </p>
          ) : (
            <div className="space-y-2">
              {(allowlist || []).map((item) => (
                <div
                  key={item.id}
                  className="p-3 rounded-md bg-raised/20 border border-separator-subtle flex items-center justify-between gap-3 text-xs transition-colors hover:bg-raised/30"
                >
                  <div className="space-y-0.5 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-text-primary font-medium truncate">
                        {item.target}
                      </span>
                      <span className="text-[10px] text-text-secondary uppercase px-1.5 py-0.2 rounded bg-raised border border-separator-muted">
                        {item.type}
                      </span>
                    </div>
                    {item.reason && (
                      <span className="text-[11px] text-text-secondary block truncate">
                        {item.reason}
                      </span>
                    )}
                  </div>

                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveAllowlist(item.id)}
                    className="text-sev-critical hover:text-sev-critical shrink-0 h-7 px-2"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                    <span className="sr-only">{t('settings.remove')}</span>
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
