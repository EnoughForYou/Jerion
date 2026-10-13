import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { AppsView } from './AppsView';
import { AppDetailDrawer } from './AppDetailDrawer';
import { MOCK_APPS } from '../../lib/bridge/mock';
import '../../i18n';

describe('AppsView & AppDetailDrawer', () => {
  it('renders package inventory table with app items', async () => {
    render(
      <MemoryRouter>
        <AppsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('com.example.mock.adware')).toBeInTheDocument();
      expect(screen.getByText('Flashlight HD (mock)')).toBeInTheDocument();
    });
  });

  it('filters apps when typing in the search box', async () => {
    render(
      <MemoryRouter>
        <AppsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('com.example.mock.adware')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText(/search by package name/i);
    fireEvent.change(searchInput, { target: { value: 'cleaner' } });

    await waitFor(() => {
      expect(screen.getByText('com.example.mock.popup')).toBeInTheDocument();
      expect(screen.queryByText('Flashlight HD (mock)')).not.toBeInTheDocument();
    });
  });

  it('filters by category pills (e.g. system packages)', async () => {
    render(
      <MemoryRouter>
        <AppsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('com.example.mock.adware')).toBeInTheDocument();
    });

    const systemPill = screen.getByRole('button', { name: /^system$/i });
    fireEvent.click(systemPill);

    await waitFor(() => {
      expect(screen.getByText('com.android.settings')).toBeInTheDocument();
      expect(screen.queryByText('com.example.mock.adware')).not.toBeInTheDocument();
    });
  });

  it('renders AppDetailDrawer with permissions, signatures, and triggers APK extraction', async () => {
    const app = MOCK_APPS[0];
    render(
      <MemoryRouter>
        <AppDetailDrawer app={app} onClose={() => {}} />
      </MemoryRouter>
    );

    expect(screen.getByText(app.label)).toBeInTheDocument();
    expect(screen.getByText(app.package)).toBeInTheDocument();
    expect(screen.getByText(`API ${app.targetSdk} (min ${app.minSdk})`)).toBeInTheDocument();

    const pullBtn = screen.getByRole('button', { name: /pull apk/i });
    fireEvent.click(pullBtn);

    await waitFor(() => {
      expect(screen.getByText(/extracted to:/i)).toBeInTheDocument();
    });
  });
});
