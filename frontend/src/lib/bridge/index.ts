import type { Bridge } from './bridge';
import { WailsBridge } from './wails';

export const CONTRACT_VERSION = '0.1.0';
export { BridgeError } from './errors';
export type { Bridge } from './bridge';

let bridgeInstance: Bridge;

if (import.meta.env.DEV && import.meta.env.VITE_BRIDGE !== 'wails') {
  const { MockBridge } = await import('./mock');
  bridgeInstance = new MockBridge();
} else {
  bridgeInstance = new WailsBridge();
}

export const bridge = bridgeInstance;

export function getBridge(): Bridge {
  return bridgeInstance;
}
