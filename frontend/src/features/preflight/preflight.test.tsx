import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import { PreflightCheckItem } from './PreflightCheckItem';
import { PreflightView } from './PreflightView';
import type { PreflightCheck } from '../../lib/bridge/schema';
import '../../i18n';

describe('Preflight components', () => {
  describe('PreflightCheckItem', () => {
    it('renders passed check cleanly', () => {
      const check: PreflightCheck = {
        id: 'adb_binary',
        title: 'ADB Binary',
        status: 'ok',
        detail: 'Android Debug Bridge located at C:\\platform-tools\\adb.exe',
      };

      render(<PreflightCheckItem check={check} />);

      expect(screen.getByText('ADB Binary')).toBeInTheDocument();
      expect(screen.getByText(/Android Debug Bridge located/)).toBeInTheDocument();
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('renders missing check with remedy and action buttons', () => {
      const handleDownload = vi.fn();
      const handleChoosePath = vi.fn();
      const handleRetry = vi.fn();

      const check: PreflightCheck = {
        id: 'adb_binary',
        title: 'ADB Binary',
        status: 'missing',
        detail: 'Binary missing on system PATH',
        remedy: 'Please download official platform-tools.',
        sourceUrl: 'https://dl.google.com/android/repository/platform-tools-latest-windows.zip',
      };

      render(
        <PreflightCheckItem
          check={check}
          onDownload={handleDownload}
          onChoosePath={handleChoosePath}
          onRetry={handleRetry}
        />
      );

      expect(screen.getByText('ADB Binary')).toBeInTheDocument();
      expect(screen.getByText('Please download official platform-tools.')).toBeInTheDocument();

      const downloadBtn = screen.getByRole('button', { name: /download/i });
      fireEvent.click(downloadBtn);
      expect(handleDownload).toHaveBeenCalledWith('adb_binary');

      const chooseBtn = screen.getByRole('button', { name: /choose existing/i });
      fireEvent.click(chooseBtn);
      expect(handleChoosePath).toHaveBeenCalledWith('adb_binary');

      const retryBtn = screen.getByRole('button', { name: /retry/i });
      fireEvent.click(retryBtn);
      expect(handleRetry).toHaveBeenCalledWith('adb_binary');
    });

    it('renders download progress bar when status is downloading', () => {
      const check: PreflightCheck = {
        id: 'adb_binary',
        title: 'ADB Binary',
        status: 'downloading',
        detail: 'Downloading package...',
        downloadProgress: 65,
      };

      render(<PreflightCheckItem check={check} />);

      expect(screen.getByText('65%')).toBeInTheDocument();
    });
  });

  describe('PreflightView', () => {
    it('runs checks and completes automatically with mock data', async () => {
      const onComplete = vi.fn();
      render(<PreflightView onComplete={onComplete} />);

      expect(screen.getByText('System Preflight')).toBeInTheDocument();

      await waitFor(() => {
        expect(onComplete).toHaveBeenCalled();
      });
    });
  });
});
