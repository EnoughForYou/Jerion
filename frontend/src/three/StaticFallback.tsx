import React from 'react';
import { Smartphone, ShieldAlert, CheckCircle2, Clock, AlertTriangle } from 'lucide-react';
import type { CanonicalSecurityState } from '../app/deviceSecurityState';

export interface StaticFallbackProps {
  status?: 'idle' | 'scanning' | 'completed';
  findingsCount?: number;
  canonicalState?: CanonicalSecurityState;
  className?: string;
}

export const StaticFallback: React.FC<StaticFallbackProps> = ({
  status = 'idle',
  findingsCount = 0,
  canonicalState,
  className = '',
}) => {
  const safeCount = Math.max(0, Number(findingsCount) || 0);

  // Derive visual indicators strictly from canonicalState if provided, otherwise fallback to status
  let title: string;
  let subtitle: string;
  let icon: React.ReactNode;

  const effectiveState =
    canonicalState ||
    (status === 'scanning'
      ? 'scanning'
      : status === 'completed'
      ? safeCount > 0
        ? 'scanned-findings'
        : 'scanned-clean'
      : 'never-scanned');

  switch (effectiveState) {
    case 'no-device':
      title = 'No Device Connected';
      subtitle = 'Connect Android phone via USB';
      icon = <Smartphone className="w-3 h-3 text-text-secondary" />;
      break;
    case 'unauthorized':
      title = 'Authorization Required';
      subtitle = 'Allow USB debugging on phone screen';
      icon = <AlertTriangle className="w-3 h-3 text-sev-high" />;
      break;
    case 'offline':
      title = 'Device Offline';
      subtitle = 'Connection inactive in ADB daemon';
      icon = <AlertTriangle className="w-3 h-3 text-sev-medium" />;
      break;
    case 'scanning':
      title = 'Scanner Active';
      subtitle = 'Analyzing applications and permissions';
      icon = <Smartphone className="w-3 h-3 text-accent animate-pulse" />;
      break;
    case 'partial-failed':
      title = 'Scan Incomplete';
      subtitle = 'Analysis was partial or interrupted';
      icon = <AlertTriangle className="w-3 h-3 text-warning" />;
      break;
    case 'scanned-findings':
      title = 'Threat Review Required';
      subtitle = `${safeCount} potential threats identified`;
      icon = <ShieldAlert className="w-3 h-3 text-sev-high" />;
      break;
    case 'scanned-clean':
      title = 'No findings in the last scan';
      subtitle = 'Zero anomalies identified';
      icon = <CheckCircle2 className="w-3 h-3 text-ok" />;
      break;
    case 'never-scanned':
    default:
      title = 'Ready to Scan';
      subtitle = 'No scan performed yet';
      icon = <Clock className="w-3 h-3 text-text-secondary" />;
      break;
  }

  return (
    <div
      data-testid="static-fallback"
      className={`relative w-full h-[220px] flex items-center justify-center select-none overflow-hidden ${className}`}
    >
      {/* Background SVG Rings */}
      <svg
        className="w-48 h-48 opacity-30 text-accent"
        viewBox="0 0 200 200"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle
          cx="100"
          cy="100"
          r="80"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeDasharray="4 4"
        />
        <circle
          cx="100"
          cy="100"
          r="60"
          stroke="currentColor"
          strokeWidth="1"
        />
        <circle
          cx="100"
          cy="100"
          r="40"
          stroke="currentColor"
          strokeWidth="1"
          strokeDasharray="2 4"
        />
      </svg>

      {/* Center Phone Chassis Wireframe Representation */}
      <div className="absolute flex flex-col items-center justify-center p-4 rounded-lg bg-surface/80 border border-separator-subtle backdrop-blur-xs shadow-md">
        <Smartphone className="w-8 h-8 text-accent mb-1.5" />
        <span className="text-[11px] font-semibold text-text-primary capitalize">
          {title}
        </span>
        <div className="flex items-center gap-1.5 mt-1 text-[10px] text-text-secondary">
          {icon}
          <span>{subtitle}</span>
        </div>
      </div>
    </div>
  );
};
