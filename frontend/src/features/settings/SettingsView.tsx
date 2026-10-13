import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Moon, Sun, Shield, Layers } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { useThemeStore } from '../../app/theme';
import { getBridge } from '../../lib/bridge';
import type { AllowlistEntry } from '../../lib/bridge/schema';

export const SettingsView: React.FC = () => {
  const { t } = useTranslation();
  const { theme, setTheme } = useThemeStore();
  const [allowlist, setAllowlist] = useState<AllowlistEntry[]>([]);

  const bridge = getBridge();

  useEffect(() => {
    async function loadSettings() {
      try {
        const list = await bridge.allowlist.list();
        setAllowlist(list);
      } catch (err) {
        console.error('Failed to load allowlist:', err);
      }
    }
    loadSettings();
  }, [bridge]);

  const bridgeMode = import.meta.env.VITE_BRIDGE === 'wails' ? 'WAILS (Backend)' : 'MOCK (Simulated)';

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center gap-3.5 p-5 bg-surface border border-separator/35 rounded-lg">
        <div className="w-11 h-11 rounded-lg bg-accent/15 text-accent flex items-center justify-center shrink-0">
          <SettingsIcon className="w-6 h-6" />
        </div>
        <div>
          <h2 className="text-md font-semibold text-text-primary">
            {t('settings.title')}
          </h2>
          <p className="text-xs text-text-secondary">
            {t('settings.general')}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* Appearance & Theme Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              {theme === 'dark' ? <Moon className="w-4 h-4 text-accent" /> : <Sun className="w-4 h-4 text-accent" />}
              <CardTitle>{t('settings.theme')}</CardTitle>
            </div>
            <CardDescription>
              Toggle between high-contrast Dark and Light interface themes
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-3">
              <Button
                variant={theme === 'dark' ? 'primary' : 'secondary'}
                size="md"
                onClick={() => setTheme('dark')}
              >
                <Moon className="w-4 h-4" />
                {t('settings.themeDark')}
              </Button>
              <Button
                variant={theme === 'light' ? 'primary' : 'secondary'}
                size="md"
                onClick={() => setTheme('light')}
              >
                <Sun className="w-4 h-4" />
                {t('settings.themeLight')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Runtime Bridge & Rules */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-accent" />
              <CardTitle>{t('settings.bridge')}</CardTitle>
            </div>
            <CardDescription>
              Backend transport connection and threat definitions
            </CardDescription>
          </CardHeader>
          <CardContent>
            <dl className="space-y-2 text-xs">
              <div className="flex justify-between py-1 border-b border-separator/15">
                <dt className="text-text-secondary">Active Bridge</dt>
                <dd className="font-mono text-text-primary font-semibold">{bridgeMode}</dd>
              </div>
              <div className="flex justify-between py-1 border-b border-separator/15">
                <dt className="text-text-secondary">{t('settings.rules')}</dt>
                <dd className="font-mono text-text-primary">v2026.03.1 (Embedded)</dd>
              </div>
              <div className="flex justify-between py-1">
                <dt className="text-text-secondary">Contract Spec</dt>
                <dd className="font-mono text-text-primary">v0.1.0 (Draft)</dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </div>

      {/* Package Allowlist */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Layers className="w-4 h-4 text-accent" />
            <CardTitle>Allowed Packages & Exclusions</CardTitle>
          </div>
          <CardDescription>
            Packages explicitly whitelisted from automated threat alerts
          </CardDescription>
        </CardHeader>
        <CardContent>
          {allowlist.length === 0 ? (
            <p className="text-xs text-text-secondary py-2">No packages currently whitelisted.</p>
          ) : (
            <div className="space-y-2">
              {allowlist.map((item) => (
                <div
                  key={item.id}
                  className="p-2.5 rounded-md bg-raised/30 border border-separator/20 flex items-center justify-between text-xs"
                >
                  <div>
                    <span className="font-mono text-text-primary font-medium block">
                      {item.target}
                    </span>
                    <span className="text-[11px] text-text-secondary">{item.reason}</span>
                  </div>
                  <span className="text-[10px] text-text-secondary uppercase px-1.5 py-0.5 rounded-xs bg-raised">
                    {item.type}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
