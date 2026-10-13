import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { StaticFallback } from './StaticFallback';
import { DeviceScene } from './DeviceScene';
import { isWebGLAvailable, prefersReducedMotion } from './webgl';

describe('Three.js DeviceScene & Fallback', () => {
  describe('StaticFallback', () => {
    it('renders SVG rings and device protected status', () => {
      render(<StaticFallback status="idle" findingsCount={0} />);

      expect(screen.getByTestId('static-fallback')).toBeInTheDocument();
      expect(screen.getByText('Device Protected')).toBeInTheDocument();
      expect(screen.getByText('No critical threats')).toBeInTheDocument();
    });

    it('renders threat count when findings exist', () => {
      render(<StaticFallback status="completed" findingsCount={3} />);

      expect(screen.getByText('3 threats identified')).toBeInTheDocument();
    });

    it('renders scanner active status', () => {
      render(<StaticFallback status="scanning" />);

      expect(screen.getByText('Scanner Active')).toBeInTheDocument();
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
      expect(screen.getByText('Device Protected')).toBeInTheDocument();
    });
  });
});
