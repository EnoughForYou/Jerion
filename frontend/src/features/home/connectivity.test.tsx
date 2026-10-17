import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ConnectionTroubleshootModal } from './ConnectionTroubleshootModal';
import { WirelessConnectModal } from './WirelessConnectModal';
import { getBridge } from '../../lib/bridge';
import '../../i18n';

describe('ConnectionTroubleshootModal', () => {
  const bridge = getBridge();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders diagnostic information when opened', async () => {
    render(
      <ConnectionTroubleshootModal
        open={true}
        onOpenChange={() => {}}
      />
    );

    // Header title
    expect(screen.getByText(/Connection Diagnostics & Troubleshooting/i)).toBeInTheDocument();

    // Wait for diagnostics to load from mock bridge
    await waitFor(() => {
      expect(screen.getByText(/Device MOCK-0001 is connected and authorized/i)).toBeInTheDocument();
    });

    // Layer sections
    expect(screen.getByText(/USB Hardware Layer/i)).toBeInTheDocument();
    expect(screen.getByText(/ADB Daemon Layer/i)).toBeInTheDocument();
    expect(screen.getByText(/Official OEM USB Drivers Directory/i)).toBeInTheDocument();

    // Check OEM driver entries
    expect(screen.getByText(/Google \(Pixel\)/i)).toBeInTheDocument();
    expect(screen.getByText('Samsung')).toBeInTheDocument();
    expect(screen.getByText(/Xiaomi \/ Redmi \/ POCO/i)).toBeInTheDocument();
    expect(screen.getByText(/Transsion \(Infinix, TECNO, itel\)/i)).toBeInTheDocument();
  });

  it('triggers soft reconnect when clicking Soft Reconnect button', async () => {
    const softReconnectSpy = vi.spyOn(bridge.device, 'softReconnect').mockResolvedValue(undefined);

    render(
      <ConnectionTroubleshootModal
        open={true}
        onOpenChange={() => {}}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Device MOCK-0001 is connected and authorized/i)).toBeInTheDocument();
    });

    const softReconnectBtn = screen.getByRole('button', { name: /Soft Reconnect/i });
    fireEvent.click(softReconnectBtn);

    await waitFor(() => {
      expect(softReconnectSpy).toHaveBeenCalledTimes(1);
    });
  });

  it('shows confirmation dialog before restarting ADB server', async () => {
    const restartSpy = vi.spyOn(bridge.device, 'restartServer').mockResolvedValue(undefined);

    render(
      <ConnectionTroubleshootModal
        open={true}
        onOpenChange={() => {}}
      />
    );

    const restartBtn = screen.getByRole('button', { name: /Restart ADB Server/i });
    fireEvent.click(restartBtn);

    // Confirmation dialog appears
    await waitFor(() => {
      expect(screen.getByText(/Confirm ADB Server Restart/i)).toBeInTheDocument();
      expect(screen.getByText(/Restarting the local ADB server daemon will terminate all active ADB connections/i)).toBeInTheDocument();
    });

    // Click confirm in the dialog
    const confirmButtons = screen.getAllByRole('button', { name: /Restart ADB Server/i });
    const modalConfirmBtn = confirmButtons[confirmButtons.length - 1];
    fireEvent.click(modalConfirmBtn);

    await waitFor(() => {
      expect(restartSpy).toHaveBeenCalledTimes(1);
    });
  });
});

describe('WirelessConnectModal', () => {
  const bridge = getBridge();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('validates 6-digit pairing code format', async () => {
    render(
      <WirelessConnectModal
        open={true}
        onOpenChange={() => {}}
      />
    );

    expect(screen.getByText(/Wireless Debugging \(Android 11\+\)/i)).toBeInTheDocument();

    const hostInput = screen.getByPlaceholderText(/192.168.1.50/i);
    const portInput = screen.getByPlaceholderText(/37123/i);
    const codeInput = screen.getByPlaceholderText(/123456/i);
    const pairBtn = screen.getByRole('button', { name: /Pair Device/i });

    // Fill invalid code (< 6 digits)
    fireEvent.change(hostInput, { target: { value: '192.168.1.100' } });
    fireEvent.change(portInput, { target: { value: '37123' } });
    fireEvent.change(codeInput, { target: { value: '123' } });
    fireEvent.click(pairBtn);

    await waitFor(() => {
      expect(screen.getByText(/Pairing code must be exactly 6 digits/i)).toBeInTheDocument();
    });
  });

  it('successfully pairs device and switches to connect tab', async () => {
    const pairSpy = vi.spyOn(bridge.device, 'wirelessPair').mockResolvedValue(undefined);

    render(
      <WirelessConnectModal
        open={true}
        onOpenChange={() => {}}
      />
    );

    const hostInput = screen.getByPlaceholderText(/192.168.1.50/i);
    const portInput = screen.getByPlaceholderText(/37123/i);
    const codeInput = screen.getByPlaceholderText(/123456/i);
    const pairBtn = screen.getByRole('button', { name: /Pair Device/i });

    fireEvent.change(hostInput, { target: { value: '192.168.1.100' } });
    fireEvent.change(portInput, { target: { value: '37123' } });
    fireEvent.change(codeInput, { target: { value: '654321' } });
    fireEvent.click(pairBtn);

    await waitFor(() => {
      expect(pairSpy).toHaveBeenCalledWith('192.168.1.100', 37123, '654321');
      expect(screen.getByText(/Device paired successfully!/i)).toBeInTheDocument();
    });
  });

  it('connects to paired wireless device endpoint', async () => {
    const connectSpy = vi.spyOn(bridge.device, 'wirelessConnect').mockResolvedValue('192.168.1.100:5555');
    const onConnected = vi.fn();

    render(
      <WirelessConnectModal
        open={true}
        onOpenChange={() => {}}
        onConnected={onConnected}
      />
    );

    // Switch to Connect tab
    const connectTab = screen.getByRole('tab', { name: /Connect Paired Device/i });
    fireEvent.pointerDown(connectTab);
    fireEvent.keyDown(connectTab, { key: 'Enter', code: 'Enter' });

    const connectBtn = await screen.findByRole('button', { name: /Connect Session/i });
    const hostInputs = screen.getAllByPlaceholderText(/192.168.1.50/i);
    const activeHostInput = hostInputs[hostInputs.length - 1];
    fireEvent.change(activeHostInput, { target: { value: '192.168.1.100' } });
    fireEvent.click(connectBtn);

    await waitFor(() => {
      expect(connectSpy).toHaveBeenCalledWith('192.168.1.100', 5555);
      expect(onConnected).toHaveBeenCalledWith('192.168.1.100:5555');
    });
  });
});
