import { describe, it, expect } from 'vitest';

describe('No Mock Literals in Non-Dev Production Code', () => {
  // Vite's built-in ?raw import lets us inspect source code directly as strings
  const sourceModules = import.meta.glob<string>(
    [
      '../StatusBar.tsx',
      '../features/settings/SettingsView.tsx',
      '../features/home/HomeView.tsx',
      '../features/scan/ScanView.tsx',
      '../features/apps/AppsView.tsx',
      '../features/monitor/MonitorView.tsx',
      '../features/audit/AuditView.tsx',
      '../features/remediate/RemediateView.tsx',
      '../features/reports/ReportsView.tsx',
      '../three/StaticFallback.tsx',
      '../three/DeviceModel.tsx',
      '../three/DeviceScene.tsx',
    ],
    { query: '?raw', import: 'default', eager: true }
  );

  const checkScriptModule = import.meta.glob<string>(
    ['../../scripts/check-no-mock.js'],
    { query: '?raw', import: 'default', eager: true }
  );

  const forbiddenLiterals = [
    "'BRIDGE: MOCK'",
    '"BRIDGE: MOCK"',
    "'MOCK (Simulated)'",
    '"MOCK (Simulated)"',
    "'Bridge: MOCK'",
    '"Bridge: MOCK"',
    "'MOCK'",
    '"MOCK"',
  ];

  it('verifies that no production UI component contains literal "MOCK" strings or tokens', () => {
    const fileEntries = Object.entries(sourceModules);
    expect(fileEntries.length).toBeGreaterThan(0);

    for (const [filePath, content] of fileEntries) {
      for (const forbidden of forbiddenLiterals) {
        const found = content.includes(forbidden);
        expect(
          found,
          `File ${filePath} must NOT contain literal ${forbidden} (must use bridgeKind or runtime diagnostics instead)`
        ).toBe(false);
      }
    }
  });

  it('ensures check-no-mock script includes forbidden patterns for UI bridge literals', () => {
    const scriptEntries = Object.entries(checkScriptModule);
    expect(scriptEntries.length).toBe(1);
    const [, scriptContent] = scriptEntries[0];

    expect(scriptContent).toContain("'BRIDGE: MOCK'");
    expect(scriptContent).toContain("'MOCK (Simulated)'");
  });
});
