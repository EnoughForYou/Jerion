import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { ReportsView } from './ReportsView';
import '../../i18n';

describe('ReportsView', () => {
  it('renders historical threat scans list', async () => {
    render(
      <MemoryRouter>
        <ReportsView />
      </MemoryRouter>
    );

    expect(screen.getByText('Scan Reports')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('scan-mock-seed-01')).toBeInTheDocument();
      expect(screen.getByText('Export HTML')).toBeInTheDocument();
      expect(screen.getByText('Export JSON')).toBeInTheDocument();
    });
  });

  it('triggers report export when export button is clicked', async () => {
    render(
      <MemoryRouter>
        <ReportsView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Export JSON')).toBeInTheDocument();
    });

    const exportBtn = screen.getByText('Export JSON');
    fireEvent.click(exportBtn);

    await waitFor(() => {
      expect(screen.getByText(/report exported successfully/i)).toBeInTheDocument();
    });
  });
});
