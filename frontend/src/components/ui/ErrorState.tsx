import React from 'react';
import { AlertCircle, RotateCcw } from 'lucide-react';
import { Button } from './Button';

export interface ErrorStateProps {
  code?: string;
  title: string;
  remedy?: string;
  onRetry?: () => void;
  action?: React.ReactNode;
  className?: string;
}

export const ErrorState: React.FC<ErrorStateProps> = ({
  code,
  title,
  remedy,
  onRetry,
  action,
  className = '',
}) => {
  return (
    <div
      className={`flex flex-col items-center justify-center p-6 text-center border border-sev-critical/30 rounded-lg bg-sev-critical/5 max-w-md mx-auto ${className}`}
    >
      <div className="w-10 h-10 rounded-full bg-sev-critical/15 text-sev-critical flex items-center justify-center mb-3">
        <AlertCircle className="w-5 h-5" />
      </div>
      {code ? (
        <span className="text-[10px] font-mono uppercase tracking-wider text-sev-critical mb-1 font-semibold">
          {code}
        </span>
      ) : null}
      <h4 className="text-md font-semibold text-text-primary mb-1">{title}</h4>
      {remedy ? (
        <p className="text-xs text-text-secondary leading-relaxed mb-4 max-w-xs">{remedy}</p>
      ) : null}
      <div className="flex items-center gap-2">
        {onRetry ? (
          <Button variant="secondary" size="sm" onClick={onRetry}>
            <RotateCcw className="w-3.5 h-3.5" />
            Retry
          </Button>
        ) : null}
        {action}
      </div>
    </div>
  );
};
