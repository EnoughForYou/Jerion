import React from 'react';
import { Smartphone, ShieldAlert, CheckCircle2 } from 'lucide-react';

export interface StaticFallbackProps {
  status?: 'idle' | 'scanning' | 'completed';
  findingsCount?: number;
  className?: string;
}

export const StaticFallback: React.FC<StaticFallbackProps> = ({
  status = 'idle',
  findingsCount = 0,
  className = '',
}) => {
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
      <div className="absolute flex flex-col items-center justify-center p-4 rounded-lg bg-surface/80 border border-separator/40 backdrop-blur-xs shadow-md">
        <Smartphone className="w-8 h-8 text-accent mb-1.5" />
        <span className="text-[11px] font-semibold text-text-primary capitalize">
          {status === 'scanning' ? 'Scanner Active' : 'Device Protected'}
        </span>
        <div className="flex items-center gap-1.5 mt-1 text-[10px] text-text-secondary">
          {findingsCount > 0 ? (
            <>
              <ShieldAlert className="w-3 h-3 text-sev-high" />
              <span>{findingsCount} threats identified</span>
            </>
          ) : (
            <>
              <CheckCircle2 className="w-3 h-3 text-ok" />
              <span>No critical threats</span>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
