import { describe, expect, it } from 'vitest';
import { BridgeError, WIRE_ERROR_PREFIX } from './errors';
import {
  AppInfoSchema,
  DeviceSchema,
  PreflightCheckSchema,
  CONTRACT_VERSION,
  ConnectionDiagnosticResultSchema,
  USBDeviceSchema,
} from './schema';

describe('Contract Schemas & Version', () => {
  it('has CONTRACT_VERSION set to 1.1.0', () => {
    expect(CONTRACT_VERSION).toBe('1.1.0');
  });

  it('strictly validates PreflightCheck and rejects unknown fields', () => {
    const valid = {
      id: 'adb_binary',
      title: 'ADB Binary',
      status: 'ok',
      detail: 'Installed at /usr/bin/adb',
      remedy: 'None needed',
    };
    expect(PreflightCheckSchema.parse(valid)).toEqual(valid);

    const withUnknownField = {
      ...valid,
      extraField: 'should fail',
    };
    expect(() => PreflightCheckSchema.parse(withUnknownField)).toThrow();
  });

  it('strictly validates Device and handles connection states', () => {
    const validDevice = {
      id: 'dev-001',
      serial: 'MOCK-0001',
      model: 'Infinix Note 11S (mock)',
      manufacturer: 'Infinix',
      androidVersion: '12',
      sdkInt: 31,
      batteryLevel: 85,
      isCharging: false,
      connectionState: 'ready',
      isEmulator: false,
    };
    expect(DeviceSchema.parse(validDevice)).toEqual(validDevice);
  });

  it('validates AppInfo with string decimal versionCode (int64 support)', () => {
    const validApp = {
      package: 'com.example.mock.test',
      label: 'Test App',
      versionName: '1.0.0',
      versionCode: '9223372036854775807', // max int64 as decimal string
      isSystem: false,
      isSideloaded: true,
      installerPackage: 'com.android.packageinstaller',
      targetSdk: 31,
      minSdk: 21,
      installedAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-02T00:00:00Z',
      permissions: ['android.permission.INTERNET'],
      dangerousPermissions: ['android.permission.SYSTEM_ALERT_WINDOW'],
      hasOverlayPermission: true,
      hasAccessibilityPermission: false,
      isDeviceAdmin: false,
      isBatteryOptimized: false,
      signatures: ['d41d8cd98f00b204e9800998ecf8427e'],
      suspiciousFlags: ['sideloaded', 'overlay'],
    };
    expect(AppInfoSchema.parse(validApp)).toEqual(validApp);

    // Reject unknown fields
    expect(() =>
      AppInfoSchema.parse({ ...validApp, unknownField: 'bad' })
    ).toThrow();
  });
});

describe('BridgeError & Wire Protocol', () => {
  it('serializes to wire format JERION_ERR:JSON', () => {
    const err = new BridgeError(
      'AdbMissing',
      'ADB binary not found',
      'Download platform-tools'
    );
    const wire = err.toWire();
    expect(wire.startsWith(WIRE_ERROR_PREFIX)).toBe(true);
    expect(wire).toBe(
      'JERION_ERR:{"code":"AdbMissing","message":"ADB binary not found","remedy":"Download platform-tools"}'
    );
  });

  it('deserializes from structured wire format correctly', () => {
    const wire =
      'JERION_ERR:{"code":"DeviceUnauthorized","message":"User denied USB RSA prompt","remedy":"Tap allow on device"}' ;
    const err = BridgeError.fromWire(wire);
    expect(err).toBeInstanceOf(BridgeError);
    expect(err.code).toBe('DeviceUnauthorized');
    expect(err.message).toBe('User denied USB RSA prompt');
    expect(err.remedy).toBe('Tap allow on device');
  });

  it('falls back to InternalError for standard non-wire errors', () => {
    const rawError = new Error('Random system crash');
    const err = BridgeError.fromWire(rawError);
    expect(err.code).toBe('InternalError');
    expect(err.message).toBe('Random system crash');
  });

  it('returns self if already a BridgeError instance', () => {
    const existing = new BridgeError('Timeout', 'ADB command timed out');
    expect(BridgeError.fromWire(existing)).toBe(existing);
  });

  it('handles ValidationFailed and CommandFailed error codes', () => {
    const valErr = new BridgeError('ValidationFailed', 'invalid pairing code');
    expect(valErr.code).toBe('ValidationFailed');
    const cmdErr = new BridgeError('CommandFailed', 'command rejected');
    expect(cmdErr.code).toBe('CommandFailed');
  });
});

describe('Connection Diagnostics & Hardware Schemas', () => {
  it('validates USBDevice and ConnectionDiagnosticResult schemas', () => {
    const usbDev = {
      deviceId: 'USB\\VID_2717&PID_FF40\\0789456',
      name: 'Infinix Note 11S',
      manufacturer: 'Infinix',
      pnpClass: 'WPD',
      status: 'OK',
      configManagerErrorCode: 0,
    };
    expect(USBDeviceSchema.parse(usbDev)).toEqual(usbDev);

    const diagResult = {
      adbServerRunning: true,
      adbVersion: '1.0.41',
      usbHardwarePresent: true,
      usbDevices: [usbDev],
      adbDevices: [
        {
          serial: 'MOCK-0001',
          state: 'ready',
          product: 'XOS',
          model: 'Infinix Note 11S',
        },
      ],
      likelyCause: 'ready' as const,
      headline: 'Device is connected and authorized.',
      possibleCauses: [],
      recommendedSteps: ['Device ready.'],
    };
    expect(ConnectionDiagnosticResultSchema.parse(diagResult)).toEqual(diagResult);
  });
});

