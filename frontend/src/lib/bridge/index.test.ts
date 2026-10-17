import { describe, expect, it } from 'vitest';
import { CONTRACT_VERSION, bridge, getBridge } from './index';

describe('Bridge Selector & Version', () => {
  it('exposes CONTRACT_VERSION 1.1.0', () => {
    expect(CONTRACT_VERSION).toBe('1.1.0');
  });

  it('exports singleton bridge instance and getter', () => {
    expect(bridge).toBeDefined();
    expect(getBridge()).toBe(bridge);
    expect(typeof bridge.preflight.run).toBe('function');
    expect(typeof bridge.device.getConnected).toBe('function');
    expect(typeof bridge.scan.quickScan).toBe('function');
    expect(typeof bridge.monitor.start).toBe('function');
    expect(typeof bridge.monitor.markPopup).toBe('function');
    expect(typeof bridge.audit.run).toBe('function');
    expect(typeof bridge.remediate.execute).toBe('function');
    expect(typeof bridge.report.list).toBe('function');
    expect(typeof bridge.settings.get).toBe('function');
    expect(typeof bridge.allowlist.list).toBe('function');
  });
});
