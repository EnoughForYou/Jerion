import type {
  Device,
  Finding,
  MonitorCandidate,
  MonitorEvent,
  PreflightCheck,
  ScanProgress,
  ScanResult,
} from './schema';

export type Unsubscribe = () => void;

export interface BridgeEventMap {
  'preflight:update': PreflightCheck;
  'device:changed': Device | null;
  'scan:progress': ScanProgress;
  'scan:finding': Finding;
  'scan:done': ScanResult;
  'monitor:event': MonitorEvent;
  'monitor:candidate': MonitorCandidate;
}

export type BridgeEventKey = keyof BridgeEventMap;
export type BridgeEventCallback<E extends BridgeEventKey> = (
  payload: BridgeEventMap[E]
) => void;
