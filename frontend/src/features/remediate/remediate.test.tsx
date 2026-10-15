import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { RemediateView } from './RemediateView';
import { RemediateConfirmDialog } from './RemediateConfirmDialog';
import { MOCK_FINDINGS } from '../../lib/bridge/mock';
import { getBridge } from '../../lib/bridge';
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

  it('renders execution error and audit persistence failure in history log', async () => {
    const bridge = getBridge();
    // Insert a failed action entry with audit persistence failure into mock history
    const failedLog = {
      id: 'log-audit-fail-1',
      timestamp: '2026-10-15T12:00:00Z',
      action: {
        id: 'act-failed-1',
        type: 'disable' as const,
        package: 'com.malware.failtest',
        title: 'Disable Failing Package',
        impact: 'Freeze background services',
        undoable: true,
        isBlocked: false,
      },
      status: 'failed' as const,
      error: 'Permission denied by package manager',
      auditSaved: false,
      auditError: 'SQLite database is locked',
    };

    // Override getHistory to return the failed log entry
    const originalGetHistory = bridge.remediate.getHistory;
    bridge.remediate.getHistory = async () => [failedLog];

    render(
      <MemoryRouter>
        <RemediateView />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Disable Failing Package')).toBeInTheDocument();
      // Execution status badge: FAILED
      expect(screen.getByText('FAILED')).toBeInTheDocument();
      // Audit persistence badge: Audit Failed
      expect(screen.getByText('Audit Failed')).toBeInTheDocument();
      // Action error message
      expect(screen.getByText(/Permission denied by package manager/)).toBeInTheDocument();
      // Audit error message
      expect(screen.getByText(/SQLite database is locked/)).toBeInTheDocument();
    });

    // Restore original mock method
    bridge.remediate.getHistory = originalGetHistory;
  });

  it('displays accurate execution feedback based on status === success and warns on audit failure', async () => {
    const bridge = getBridge();
    const originalExecute = bridge.remediate.execute;

    // Simulate partial failure with audit failure
    bridge.remediate.execute = async () => [
      {
        id: 'log-res-1',
        timestamp: '2026-10-15T12:01:00Z',
        action: MOCK_FINDINGS[0].recommendedAction!,
        status: 'success',
        auditSaved: false,
        auditError: 'Write failed',
      },
      {
        id: 'log-res-2',
        timestamp: '2026-10-15T12:01:01Z',
        action: MOCK_FINDINGS[1].recommendedAction!,
        status: 'failed',
        error: 'Command timed out',
        auditSaved: true,
      },
    ];

    render(
      <MemoryRouter>
        <RemediateView />
      </MemoryRouter>
    );

    const executeBtn = await screen.findByRole('button', { name: /execute staged actions/i });
    fireEvent.click(executeBtn);

    // In the dialog, click proceed with execution
    const proceedBtn = await screen.findByRole('button', { name: /proceed with execution/i });
    fireEvent.click(proceedBtn);

    // Feedback should show 1 success, 1 failed, and audit warning
    await waitFor(() => {
      expect(screen.getByText(/Executed 1 of 2 actions successfully. 1 action\(s\) failed./i)).toBeInTheDocument();
      expect(screen.getByText(/Warning: Local audit log could not be saved for 1 action\(s\)./i)).toBeInTheDocument();
    });

    bridge.remediate.execute = originalExecute;
  });
});
