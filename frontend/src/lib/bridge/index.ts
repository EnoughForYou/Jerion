import type { Bridge } from './bridge';
import { WailsBridge } from './wails';

export const CONTRACT_VERSION = '1.0.0';
export { BridgeError } from './errors';
export type { Bridge } from './bridge';

function hasWailsRuntime(): boolean {
  return (
    typeof window !== 'undefined' &&
    Boolean(
      (window as unknown as { go?: { main?: { App?: unknown } } })?.go?.main?.App
    )
  );
}

let bridgeInstance: Bridge;

if (import.meta.env.DEV && import.meta.env.VITE_BRIDGE !== 'wails' && !hasWailsRuntime()) {
  const { MockBridge } = await import('./mock');
  bridgeInstance = new MockBridge();
} else {
  bridgeInstance = new WailsBridge();
}

export const bridge = bridgeInstance;

export function getBridge(): Bridge {
  return bridgeInstance;
}
