import { create } from 'zustand';
import type { Device } from '../lib/bridge/schema';

interface DeviceState {
  activeDevice: Device | null;
  devices: Device[];
  activeJob: string | null;
  jobProgress: number | null;
  setActiveDevice: (device: Device | null) => void;
  setDevices: (devices: Device[]) => void;
  setActiveJob: (job: string | null, progress?: number | null) => void;
}

export const useDeviceStore = create<DeviceState>((set) => ({
  activeDevice: null,
  devices: [],
  activeJob: null,
  jobProgress: null,
  setActiveDevice: (device) => set({ activeDevice: device }),
  setDevices: (devices) => set({ devices }),
  setActiveJob: (job, progress = null) => set({ activeJob: job, jobProgress: progress }),
}));
