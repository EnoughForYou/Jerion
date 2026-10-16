import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { StaticFallback } from './StaticFallback';
import { DeviceScene } from './DeviceScene';
import { isWebGLAvailable, prefersReducedMotion } from './webgl';

describe('Three.js DeviceScene & Fallback', () => {
  describe('StaticFallback', () => {
    it('renders SVG rings and ready status when idle without scan', () => {
      render(<StaticFallback status="idle" findingsCount={0} />);

      expect(screen.getByTestId('static-fallback')).toBeInTheDocument();
      expect(screen.getByText('Ready to Scan')).toBeInTheDocument();
      expect(screen.getByText('No scan performed yet')).toBeInTheDocument();
    });

    it('renders clean verdict when scan is completed with 0 findings', () => {
      render(<StaticFallback status="completed" findingsCount={0} />);

      expect(screen.getByText('No findings in the last scan')).toBeInTheDocument();
      expect(screen.getByText('Zero anomalies identified')).toBeInTheDocument();
    });

    it('renders threat review status when findings exist', () => {
      render(<StaticFallback status="completed" findingsCount={3} />);

      expect(screen.getByText('Threat Review Required')).toBeInTheDocument();
      expect(screen.getByText('3 potential threats identified')).toBeInTheDocument();
    });

    it('renders scanner active status', () => {
      render(<StaticFallback status="scanning" />);

      expect(screen.getByText('Scanner Active')).toBeInTheDocument();
      expect(screen.getByText('Analyzing applications and permissions')).toBeInTheDocument();
    });

    it('renders canonicalState correctly for no-device, unauthorized, and offline', () => {
      const { rerender } = render(<StaticFallback canonicalState="no-device" />);
      expect(screen.getByText('No Device Connected')).toBeInTheDocument();

      rerender(<StaticFallback canonicalState="unauthorized" />);
      expect(screen.getByText('Authorization Required')).toBeInTheDocument();

      rerender(<StaticFallback canonicalState="offline" />);
      expect(screen.getByText('Device Offline')).toBeInTheDocument();
    });
  });

  describe('DeviceScene WebGL Detection', () => {
    it('detects WebGL availability and motion preference', () => {
      expect(typeof isWebGLAvailable()).toBe('boolean');
      expect(typeof prefersReducedMotion()).toBe('boolean');
    });

    it('gracefully falls back to StaticFallback in environment without WebGL', () => {
      render(<DeviceScene status="idle" findingsCount={0} />);

      // In jsdom without WebGL, StaticFallback must be rendered
      expect(screen.getByTestId('static-fallback')).toBeInTheDocument();
      expect(screen.getByText('Ready to Scan')).toBeInTheDocument();
    });
  });
});
