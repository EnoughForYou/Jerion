import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { AuditView } from './AuditView';
import { useDeviceStore } from '../../app/deviceStore';
import { MOCK_DEVICE } from '../../lib/bridge/mock';
import '../../i18n';

describe('AuditView (System Security Audit)', () => {
  beforeEach(() => {
    useDeviceStore.getState().setActiveDevice(MOCK_DEVICE);
  });

  it('renders No Device Connected empty state when activeDevice is null', async () => {
    useDeviceStore.getState().setActiveDevice(null);
    const bridge = (await import('../../lib/bridge')).getBridge();
    const origGet = bridge.device.getConnected;
    bridge.device.getConnected = async () => null;

    render(
      <MemoryRouter>
        <AuditView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No Device Connected')).toBeInTheDocument();
    });

    bridge.device.getConnected = origGet;
  });

  it('renders audit checklist items with status badges and categories', async () => {
    render(
      <MemoryRouter>
        <AuditView />
      </MemoryRouter>
    );

    expect(screen.getByText('System Security Audit')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('USB Debugging Authorization')).toBeInTheDocument();
      expect(screen.getByText('ADB over Wi-Fi Enabled')).toBeInTheDocument();
      expect(screen.getByText('Compliant')).toBeInTheDocument();
      expect(screen.getByText('Warning')).toBeInTheDocument();
    });
  });

  it('filters audit items by search query', async () => {
    render(
      <MemoryRouter>
        <AuditView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('USB Debugging Authorization')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText(/search audit items/i);
    fireEvent.change(searchInput, { target: { value: 'Wi-Fi' } });

    await waitFor(() => {
      expect(screen.getByText('ADB over Wi-Fi Enabled')).toBeInTheDocument();
      expect(screen.queryByText('USB Debugging Authorization')).not.toBeInTheDocument();
    });
  });

  it('filters audit items by category pills', async () => {
    render(
      <MemoryRouter>
        <AuditView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('USB Debugging Authorization')).toBeInTheDocument();
    });

    const securityPill = screen.getByRole('button', { name: /^security$/i });
    fireEvent.click(securityPill);

    await waitFor(() => {
      expect(screen.getByText('Unknown Sources Installation')).toBeInTheDocument();
      expect(screen.queryByText('USB Debugging Authorization')).not.toBeInTheDocument();
    });
  });
});
