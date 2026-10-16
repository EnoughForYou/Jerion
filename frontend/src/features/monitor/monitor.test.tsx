import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { MonitorView } from './MonitorView';
import { useDeviceStore } from '../../app/deviceStore';
import { MOCK_DEVICE } from '../../lib/bridge/mock';
import '../../i18n';

describe('MonitorView (Popup Hunter)', () => {
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
        <MonitorView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No Device Connected')).toBeInTheDocument();
    });

    bridge.device.getConnected = origGet;
  });

  it('renders MonitorView with start button and telemetry signals', async () => {
    render(
      <MemoryRouter>
        <MonitorView />
      </MemoryRouter>
    );

    expect(screen.getByText('Popup Hunter')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Start Monitoring')).toBeInTheDocument();
      expect(screen.getByText('Window Focus')).toBeInTheDocument();
      expect(screen.getByText('Logcat Traces')).toBeInTheDocument();
    });
  });

  it('starts monitoring and enables "Popup Just Appeared!" button', async () => {
    render(
      <MemoryRouter>
        <MonitorView />
      </MemoryRouter>
    );

    const startBtn = screen.getByRole('button', { name: /start monitoring/i });
    fireEvent.click(startBtn);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /popup just appeared/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /stop monitoring/i })).toBeInTheDocument();
    });
  });

  it('records popup marker and updates candidates when "Popup Just Appeared!" is clicked', async () => {
    render(
      <MemoryRouter>
        <MonitorView />
      </MemoryRouter>
    );

    const startBtn = screen.getByRole('button', { name: /start monitoring/i });
    fireEvent.click(startBtn);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /popup just appeared/i })).toBeInTheDocument();
    });

    const markBtn = screen.getByRole('button', { name: /popup just appeared/i });
    fireEvent.click(markBtn);

    await waitFor(() => {
      expect(screen.getByText(/marker recorded/i)).toBeInTheDocument();
      expect(screen.getByText('Cleaner Master')).toBeInTheDocument();
    });
  });
});
