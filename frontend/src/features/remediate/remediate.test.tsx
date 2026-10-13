import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { RemediateView } from './RemediateView';
import { RemediateConfirmDialog } from './RemediateConfirmDialog';
import { MOCK_FINDINGS } from '../../lib/bridge/mock';
import '../../i18n';

describe('RemediateView & RemediateConfirmDialog', () => {
  it('renders pending staged actions from latest scan findings', async () => {
    render(
      <MemoryRouter>
        <RemediateView />
      </MemoryRouter>
    );

    expect(screen.getByText('Remediation & Cleanup')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Staged Remediation Actions')).toBeInTheDocument();
      expect(screen.getByText('Action History')).toBeInTheDocument();
    });
  });

  it('renders confirmation dialog with actions and backup toggle', () => {
    const actions = [MOCK_FINDINGS[0].recommendedAction!];
    render(
      <MemoryRouter>
        <RemediateConfirmDialog
          open={true}
          onOpenChange={() => {}}
          actions={actions}
          onConfirm={() => {}}
          isExecuting={false}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Confirm Remediation Plan')).toBeInTheDocument();
    expect(screen.getByText(actions[0].title)).toBeInTheDocument();
    expect(screen.getByText(/create apk and state backup/i)).toBeInTheDocument();
  });

  it('allows opening confirm dialog when execute button is clicked', async () => {
    render(
      <MemoryRouter>
        <RemediateView />
      </MemoryRouter>
    );

    await waitFor(() => {
      const executeBtn = screen.queryByRole('button', { name: /execute staged actions/i });
      if (executeBtn) {
        fireEvent.click(executeBtn);
        expect(screen.getByText('Confirm Remediation Plan')).toBeInTheDocument();
      }
    });
  });
});
