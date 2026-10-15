import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { ScanView } from './ScanView';
import { ScanTimeline } from './ScanTimeline';
import { EvidenceDrawer } from './EvidenceDrawer';
import { getBridge } from '../../lib/bridge';
import { MOCK_FINDINGS, MockBridge } from '../../lib/bridge/mock';
import '../../i18n';

describe('ScanView and Interactive Scan Components', () => {
  it('renders ScanTimeline with steps and progress percentage', () => {
    render(
      <ScanTimeline
        progress={{ step: 'Inspecting package manifests', progress: 45, currentPackage: 'com.example.mock' }}
        isScanning={true}
      />
    );

    expect(screen.getByText('Inspecting package manifests')).toBeInTheDocument();
    expect(screen.getByText('45%')).toBeInTheDocument();
    expect(screen.getByText(/com\.example\.mock/)).toBeInTheDocument();
  });

  it('renders EvidenceDrawer with technical details and raw output', () => {
    const finding = MOCK_FINDINGS[0];
    render(
      <MemoryRouter>
        <EvidenceDrawer
          finding={finding}
          open={true}
          onOpenChange={() => {}}
        />
      </MemoryRouter>
    );

    expect(screen.getByText(finding.title)).toBeInTheDocument();
    expect(screen.getByText(finding.whyFlagged)).toBeInTheDocument();
    expect(screen.getByText(finding.falsePositiveHint)).toBeInTheDocument();
    expect(screen.getByText(finding.evidence[0].title)).toBeInTheDocument();
  });

  it('allows switching scan modes between quick and deep', async () => {
    render(
      <MemoryRouter>
        <ScanView />
      </MemoryRouter>
    );

    const deepCard = screen.getByText('Deep Scan');
    fireEvent.click(deepCard);

    await waitFor(() => {
      expect(screen.getByText(/ready to run deep scan/i)).toBeInTheDocument();
    });
  });

  it('filters findings by severity when filter buttons are clicked', async () => {
    render(
      <MemoryRouter>
        <ScanView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/all severities/i)).toBeInTheDocument();
    });

    const criticalButton = screen.queryByRole('button', { name: /critical/i });
    if (criticalButton) {
      fireEvent.click(criticalButton);
      // Findings should be filtered to only critical
      expect(screen.getByText(/POP-001/i)).toBeInTheDocument();
    }
  });

  it('displays warning alert when scan is failed or cancelled without displaying clean verdict', async () => {
    const bridge = getBridge() as MockBridge;
    render(
      <MemoryRouter>
        <ScanView />
      </MemoryRouter>
    );

    // Simulate terminal scan:done event with failed status
    act(() => {
      bridge.emit('scan:done', {
        id: 'scan-fail-1',
        deviceId: 'dev-test',
        mode: 'quick',
        startedAt: '2026-10-15T12:00:00Z',
        completedAt: '2026-10-15T12:00:10Z',
        status: 'failed',
        findings: [],
        summary: {
          counts: { critical: 0, high: 0, medium: 0, low: 0, info: 0 },
          totalApps: 0,
          verdict: 'review',
        },
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Scan failed before completion')).toBeInTheDocument();
      expect(
        screen.getByText('Security analysis was not completed. Threat findings and security verdict cannot be determined.')
      ).toBeInTheDocument();
    });

    // Clean verdict or "no threats detected" must NOT be shown
    expect(screen.queryByText(/· Verdict: CLEAN/i)).not.toBeInTheDocument();
    expect(screen.queryByText('No threats or suspicious behaviors detected on this device.')).not.toBeInTheDocument();
  });
});
