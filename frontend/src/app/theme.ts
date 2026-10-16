import { create } from 'zustand';

export type Theme = 'dark' | 'light';
export type ThemeMode = 'system' | 'dark' | 'light';

const STORAGE_KEY = 'jerion_theme_mode';
const LEGACY_STORAGE_KEY = 'jerion_theme';

export function getSystemTheme(): Theme {
  if (typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    return 'light';
  }
  return 'dark';
}

export function getInitialThemeMode(): ThemeMode {
  if (typeof window === 'undefined') {
    return 'system';
  }

  const stored = localStorage.getItem(STORAGE_KEY) || localStorage.getItem(LEGACY_STORAGE_KEY);
  if (stored === 'dark' || stored === 'light' || stored === 'system') {
    return stored as ThemeMode;
  }

  return 'system';
}

export function resolveTheme(mode: ThemeMode): Theme {
  if (mode === 'system') {
    return getSystemTheme();
  }
  return mode;
}

export function getInitialTheme(): Theme {
  return resolveTheme(getInitialThemeMode());
}

export function applyTheme(theme: Theme): void {
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', theme);
  }
}

interface ThemeState {
  themeMode: ThemeMode;
  theme: Theme;
  setThemeMode: (mode: ThemeMode) => void;
  setTheme: (theme: Theme | ThemeMode) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>((set, get) => {
  const initialMode = getInitialThemeMode();
  const initialTheme = resolveTheme(initialMode);
  applyTheme(initialTheme);

  return {
    themeMode: initialMode,
    theme: initialTheme,
    setThemeMode: (mode: ThemeMode) => {
      const resolved = resolveTheme(mode);
      applyTheme(resolved);
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, mode);
      }
      set({ themeMode: mode, theme: resolved });
    },
    setTheme: (target: Theme | ThemeMode) => {
      get().setThemeMode(target);
    },
    toggleTheme: () => {
      const current = get().theme;
      const next: Theme = current === 'dark' ? 'light' : 'dark';
      get().setThemeMode(next);
    },
  };
});

// Listen to OS system color-scheme changes if currently on system theme
if (typeof window !== 'undefined' && window.matchMedia) {
  try {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');
    const handleSystemChange = () => {
      const state = useThemeStore.getState();
      if (state.themeMode === 'system') {
        const nextTheme = getSystemTheme();
        applyTheme(nextTheme);
        useThemeStore.setState({ theme: nextTheme });
      }
    };

    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleSystemChange);
    } else if (mediaQuery.addListener) {
      // Legacy browser fallback
      mediaQuery.addListener(handleSystemChange);
    }
  } catch {
    // Ignore environments where matchMedia is mocked or unsupported
  }
}
