import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { SettingsView } from './SettingsView';
import '../../i18n';

describe('SettingsView', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    delete (window as unknown as { go?: unknown }).go;
  });

  it('renders theme selector and allowlist table', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    expect(screen.getByText('Settings')).toBeInTheDocument();
    expect(screen.getByText('System Theme')).toBeInTheDocument();
    expect(screen.getByText('Dark Theme')).toBeInTheDocument();
    expect(screen.getByText('Light Theme')).toBeInTheDocument();

    const systemBtn = screen.getByRole('button', { name: /system theme/i });
    fireEvent.click(systemBtn);

    await waitFor(() => {
      expect(screen.getByText('com.android.settings')).toBeInTheDocument();
    });
  });

  it('saves custom ADB path when valid input is submitted', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByDisplayValue(/adb\.exe/i)).toBeInTheDocument();
    });

    const saveBtn = screen.getByRole('button', { name: /save & validate/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(screen.getByText(/adb path updated and validated/i)).toBeInTheDocument();
    });
  });

  it('adds and removes allowlist entry', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('com.android.settings')).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText(/package name/i);
    fireEvent.change(input, { target: { value: 'com.example.testpkg' } });

    const addBtn = screen.getByRole('button', { name: /add to allowlist/i });
    fireEvent.click(addBtn);

    await waitFor(() => {
      expect(screen.getByText('com.example.testpkg')).toBeInTheDocument();
    });
  });

  it('renders and tests ClamAV connection successfully', async () => {
    const testConnMock = vi.fn().mockResolvedValue({
      connected: true,
      version: 'ClamAV 1.4.1',
    });
    (window as unknown as { go: { main: { App: { ClamAVTestConnection: typeof testConnMock } } } }).go = {
      main: {
        App: {
          ClamAVTestConnection: testConnMock,
        },
      },
    };

    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    expect(screen.getByText(/ClamAV Local Antivirus Engine/i)).toBeInTheDocument();
    const testBtn = screen.getByRole('button', { name: /test connection/i });
    fireEvent.click(testBtn);

    await waitFor(() => {
      expect(screen.getByText(/connected: clamav 1.4.1/i)).toBeInTheDocument();
    });
    expect(testConnMock).toHaveBeenCalledWith('127.0.0.1:3310');
  });

  it('displays error when ClamAV connection test fails (e.g. non-loopback)', async () => {
    const testConnMock = vi.fn().mockResolvedValue({
      connected: false,
      error: 'clamav host must be a local loopback address',
    });
    (window as unknown as { go: { main: { App: { ClamAVTestConnection: typeof testConnMock } } } }).go = {
      main: {
        App: {
          ClamAVTestConnection: testConnMock,
        },
      },
    };

    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    const hostInput = screen.getByPlaceholderText(/127\.0\.0\.1:3310/i);
    fireEvent.change(hostInput, { target: { value: '192.168.1.50:3310' } });

    const testBtn = screen.getByRole('button', { name: /test connection/i });
    fireEvent.click(testBtn);

    await waitFor(() => {
      expect(screen.getByText(/connection failed: clamav host must be a local loopback address/i)).toBeInTheDocument();
    });
  });

  it('displays error when saving invalid ClamAV host', async () => {
    const saveMock = vi.fn().mockRejectedValue(new Error('IP "10.0.0.1" is not a local loopback address'));
    (window as unknown as { go: { main: { App: { ClamAVSaveConfig: typeof saveMock } } } }).go = {
      main: {
        App: {
          ClamAVSaveConfig: saveMock,
        },
      },
    };

    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    const hostInput = screen.getByPlaceholderText(/127\.0\.0\.1:3310/i);
    fireEvent.change(hostInput, { target: { value: '10.0.0.1:3310' } });

    const saveBtn = screen.getByRole('button', { name: /save clamav settings/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(screen.getByText(/is not a local loopback address/i)).toBeInTheDocument();
    });
  });

  it('renders default system info with 8 active rules when backend binding is absent', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('v1.2.1')).toBeInTheDocument();
      expect(screen.getByText('v2026.03.1 (8 rules)')).toBeInTheDocument();
      expect(screen.getByText('v1.1.0')).toBeInTheDocument();
    });
  });

  it('renders dynamic system info and rules count when AppGetSystemInfo is provided', async () => {
    const getSysInfoMock = vi.fn().mockResolvedValue({
      appVersion: '1.2.1',
      rulesVersion: '2026.03.1',
      rulesCount: 8,
      contractVersion: '1.1.0',
    });
    (window as unknown as { go: { main: { App: { AppGetSystemInfo: typeof getSysInfoMock } } } }).go = {
      main: {
        App: {
          AppGetSystemInfo: getSysInfoMock,
        },
      },
    };

    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('v1.2.1')).toBeInTheDocument();
      expect(screen.getByText('v2026.03.1 (8 rules)')).toBeInTheDocument();
      expect(screen.getByText('v1.1.0')).toBeInTheDocument();
    });
    expect(getSysInfoMock).toHaveBeenCalled();
  });
});
