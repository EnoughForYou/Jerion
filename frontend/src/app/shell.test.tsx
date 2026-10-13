import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { DeviceHeader } from './DeviceHeader';
import { StatusBar } from './StatusBar';
import { DetailPanel } from './DetailPanel';
import { AppLayout } from './AppLayout';
import { useDeviceStore } from './deviceStore';
import type { Device } from '../lib/bridge/schema';
import '../i18n';

describe('App Shell Components', () => {
  beforeEach(() => {
    useDeviceStore.setState({
      activeDevice: null,
      devices: [],
      activeJob: null,
      jobProgress: null,
    });
  });

  describe('DeviceHeader', () => {
    it('renders "No device connected" when activeDevice is null', () => {
      render(<DeviceHeader />);
      expect(screen.getByText('No Device Connected')).toBeInTheDocument();
    });

    it('renders active device details and status badge', () => {
      const mockDevice: Device = {
        id: 'mock-device-1',
        serial: 'MOCK-0001',
        model: 'Infinix Note 11S (mock)',
        manufacturer: 'Infinix',
        androidVersion: '12',
        sdkInt: 31,
        batteryLevel: 84,
        isCharging: true,
        connectionState: 'ready',
        isEmulator: false,
      };

      useDeviceStore.setState({ activeDevice: mockDevice });
      render(<DeviceHeader />);

      expect(screen.getByText('Infinix Note 11S (mock)')).toBeInTheDocument();
      expect(screen.getByText(/Android 12/)).toBeInTheDocument();
      expect(screen.getByText('84%')).toBeInTheDocument();
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
  });

  describe('StatusBar', () => {
    it('renders idle status and rule package version', () => {
      render(<StatusBar />);
      expect(screen.getByText('System Idle')).toBeInTheDocument();
      expect(screen.getByText('Rules v2026.03.1')).toBeInTheDocument();
    });

    it('renders active job when set', () => {
      useDeviceStore.setState({ activeJob: 'Quick Scan', jobProgress: 45 });
      render(<StatusBar />);
      expect(screen.getByText('Running: Quick Scan')).toBeInTheDocument();
    });
  });

  describe('DetailPanel', () => {
    it('renders when open and closes on close button click', () => {
      let open = true;
      const handleClose = () => {
        open = false;
      };

      const { rerender } = render(
        <DetailPanel open={open} onClose={handleClose} title="Package Info">
          <div>Package content</div>
        </DetailPanel>
      );

      expect(screen.getByText('Package Info')).toBeInTheDocument();
      expect(screen.getByText('Package content')).toBeInTheDocument();

      const closeBtn = screen.getByRole('button', { name: /close detail panel/i });
      fireEvent.click(closeBtn);

      rerender(
        <DetailPanel open={open} onClose={handleClose} title="Package Info">
          <div>Package content</div>
        </DetailPanel>
      );

      expect(screen.queryByText('Package Info')).not.toBeInTheDocument();
    });
  });

  describe('AppLayout Navigation', () => {
    it('renders sidebar navigation links and switches active route', () => {
      const memoryRouter = createMemoryRouter(
        [
          {
            path: '/',
            element: <AppLayout />,
            children: [
              { index: true, element: <div>Home Content</div> },
              { path: 'apps', element: <div>Applications Content</div> },
            ],
          },
        ],
        { initialEntries: ['/'] }
      );

      render(<RouterProvider router={memoryRouter} />);

      expect(screen.getByText('Home Content')).toBeInTheDocument();

      const appsLink = screen.getByRole('button', { name: 'Applications' });
      fireEvent.click(appsLink);

      expect(screen.getByText('Applications Content')).toBeInTheDocument();
    });
  });
});
