import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  useThemeStore,
  applyTheme,
  getInitialTheme,
  getInitialThemeMode,
  resolveTheme,
} from './theme';

describe('theme store', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('defaults to system mode when localStorage is empty', () => {
    const mode = getInitialThemeMode();
    expect(mode).toBe('system');
    const theme = getInitialTheme();
    expect(['dark', 'light']).toContain(theme);
  });

  it('reads stored theme from localStorage', () => {
    localStorage.setItem('jerion_theme_mode', 'light');
    expect(getInitialThemeMode()).toBe('light');
    expect(resolveTheme('light')).toBe('light');
  });

  it('reads legacy stored theme if mode is not set', () => {
    localStorage.setItem('jerion_theme', 'dark');
    expect(getInitialThemeMode()).toBe('dark');
  });

  it('applies data-theme attribute on document root', () => {
    applyTheme('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');

    applyTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('updates state and DOM when setThemeMode is called', () => {
    useThemeStore.getState().setThemeMode('light');
    expect(useThemeStore.getState().themeMode).toBe('light');
    expect(useThemeStore.getState().theme).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    expect(localStorage.getItem('jerion_theme_mode')).toBe('light');

    useThemeStore.getState().setThemeMode('dark');
    expect(useThemeStore.getState().themeMode).toBe('dark');
    expect(useThemeStore.getState().theme).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(localStorage.getItem('jerion_theme_mode')).toBe('dark');

    useThemeStore.getState().setThemeMode('system');
    expect(useThemeStore.getState().themeMode).toBe('system');
    expect(['dark', 'light']).toContain(useThemeStore.getState().theme);
    expect(localStorage.getItem('jerion_theme_mode')).toBe('system');
  });

  it('supports setTheme alias for backward compatibility', () => {
    useThemeStore.getState().setTheme('light');
    expect(useThemeStore.getState().theme).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('toggles theme correctly', () => {
    useThemeStore.getState().setThemeMode('dark');
    useThemeStore.getState().toggleTheme();
    expect(useThemeStore.getState().theme).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');

    useThemeStore.getState().toggleTheme();
    expect(useThemeStore.getState().theme).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('resolves system mode using matchMedia if available', () => {
    const originalMatchMedia = window.matchMedia;
    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: query.includes('light'),
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    expect(resolveTheme('system')).toBe('light');

    window.matchMedia = originalMatchMedia;
  });
});
