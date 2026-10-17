import { create } from 'zustand';

export interface SystemInfo {
  appVersion: string;
  rulesVersion: string;
  rulesCount: number;
  adbVersion?: string;
  contractVersion: string;
}

interface SystemInfoState {
  systemInfo: SystemInfo | null;
  fetchSystemInfo: () => Promise<void>;
  setSystemInfo: (info: SystemInfo) => void;
}

export const useSystemInfoStore = create<SystemInfoState>((set) => ({
  systemInfo: null,
  fetchSystemInfo: async () => {
    try {
      const goApp = (window as unknown as {
        go?: {
          main?: {
            App?: {
              AppGetSystemInfo?: () => Promise<SystemInfo>;
            };
          };
        };
      })?.go?.main?.App;

      if (goApp?.AppGetSystemInfo) {
        const info = await goApp.AppGetSystemInfo();
        if (info) {
          set({ systemInfo: info });
        }
      }
    } catch (err) {
      console.warn('Failed to fetch system info from backend:', err);
    }
  },
  setSystemInfo: (info) => set({ systemInfo: info }),
}));
