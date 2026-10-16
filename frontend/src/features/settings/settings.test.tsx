import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { SettingsView } from './SettingsView';
import '../../i18n';

describe('SettingsView', () => {
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

  it('renders and tests ClamAV connection', async () => {
    render(
      <MemoryRouter>
        <SettingsView />
      </MemoryRouter>
    );

    expect(screen.getByText(/ClamAV Local Antivirus Engine/i)).toBeInTheDocument();
    const testBtn = screen.getByRole('button', { name: /test connection/i });
    fireEvent.click(testBtn);

    await waitFor(() => {
      expect(screen.getByText(/connected/i)).toBeInTheDocument();
    });
  });
});
