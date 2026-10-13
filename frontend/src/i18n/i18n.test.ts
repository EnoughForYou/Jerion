import { describe, it, expect } from 'vitest';
import i18n from './index';

describe('i18n configuration', () => {
  it('is initialized with language "en"', () => {
    expect(i18n.language).toBe('en');
  });

  it('translates navigation keys correctly', () => {
    expect(i18n.t('nav.home')).toBe('Home');
    expect(i18n.t('nav.apps')).toBe('Applications');
    expect(i18n.t('nav.scan')).toBe('Scanner');
    expect(i18n.t('nav.monitor')).toBe('Popup Hunter');
    expect(i18n.t('nav.audit')).toBe('System Audit');
    expect(i18n.t('nav.cleanup')).toBe('Remediation');
    expect(i18n.t('nav.reports')).toBe('Reports');
    expect(i18n.t('nav.settings')).toBe('Settings');
  });

  it('translates preflight checks and actions', () => {
    expect(i18n.t('preflight.title')).toBe('System Preflight');
    expect(i18n.t('preflight.checks.adb_binary.name')).toBe('ADB Binary');
    expect(i18n.t('preflight.status.missing')).toBe('Action Required');
    expect(i18n.t('common.download')).toBe('Download');
    expect(i18n.t('common.choosePath')).toBe('Choose existing...');
  });

  it('translates interpolation variables properly', () => {
    const text = i18n.t('statusBar.ruleVersion', { version: '2026.03.1' });
    expect(text).toBe('Rules v2026.03.1');

    const runningText = i18n.t('statusBar.running', { job: 'Quick Scan' });
    expect(runningText).toBe('Running: Quick Scan');
  });
});
