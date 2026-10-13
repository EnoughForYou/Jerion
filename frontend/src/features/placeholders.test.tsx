import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { HomeView } from './home/HomeView';
import { AppsView } from './apps/AppsView';
import { ScanView } from './scan/ScanView';
import { MonitorView } from './monitor/MonitorView';
import { AuditView } from './audit/AuditView';
import { RemediateView } from './remediate/RemediateView';
import { ReportsView } from './reports/ReportsView';
import { SettingsView } from './settings/SettingsView';
import '../i18n';

describe('Feature Placeholder Views with Mock Bridge Fixtures', () => {
  it('renders HomeView with device specs and security summary', async () => {
    render(
      <MemoryRouter>
        <HomeView />
      </MemoryRouter>
    );

    expect(screen.getByText('Device Overview')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Infinix Note 11S (mock)')).toBeInTheDocument();
      expect(screen.getByText('Security Findings Summary')).toBeInTheDocument();
    });
  });

  it('renders AppsView with realistic inventory package table', async () => {
    render(
      <MemoryRouter>
        <AppsView />
      </MemoryRouter>
    );

    expect(screen.getByText('Installed Applications')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('com.example.mock.adware')).toBeInTheDocument();
    });
  });

  it('renders ScanView with Quick/Deep scan options', async () => {
    render(
      <MemoryRouter>
        <ScanView />
      </MemoryRouter>
    );

    expect(screen.getByText('Quick Scan')).toBeInTheDocument();
    expect(screen.getByText('Deep Scan')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /start scan/i })).toBeInTheDocument();
    });
  });

  it('renders MonitorView for Popup Hunter with signal monitors', async () => {
    render(
      <MemoryRouter>
        <MonitorView />
      </MemoryRouter>
    );

    expect(screen.getByText('Popup Hunter')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /start monitoring/i })).toBeInTheDocument();
    });
  });

  it('renders AuditView with system hardening audit checklist', async () => {
    render(
      <MemoryRouter>
        <AuditView />
      </MemoryRouter>
    );

    expect(screen.getByText('System Security Audit')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Unknown Sources Installation')).toBeInTheDocument();
    });
  });

  it('renders RemediateView with action history and guardrails', async () => {
    render(
      <MemoryRouter>
        <RemediateView />
      </MemoryRouter>
    );

    expect(screen.getByText('Remediation & Cleanup')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Action History')).toBeInTheDocument();
    });
  });

  it('renders ReportsView with historical scans and export options', async () => {
    render(
      <MemoryRouter>
        <ReportsView />
      </MemoryRouter>
    );

    expect(screen.getByText('Scan Reports')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Historical Threat Scans')).toBeInTheDocument();
    });
  });

  it('renders SettingsView with theme toggles and bridge mode', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    expect(screen.getByText('Settings')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /dark theme/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /light theme/i })).toBeInTheDocument();
    });
  });
});
