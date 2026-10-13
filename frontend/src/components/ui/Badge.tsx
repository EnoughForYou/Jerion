import React from 'react';
import {
  AlertOctagon,
  AlertTriangle,
  AlertCircle,
  Info,
  HelpCircle,
  CheckCircle2,
} from 'lucide-react';
import type { Severity } from '../../lib/bridge/schema';

export type BadgeSeverity = Severity | 'ok';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  severity?: BadgeSeverity;
  count?: number;
  showIcon?: boolean;
}

export const Badge: React.FC<BadgeProps> = ({
  severity = 'info',
  count,
  showIcon = true,
  children,
  className = '',
  ...props
}) => {
  const config: Record<
    BadgeSeverity,
    { label: string; icon: React.ReactNode; style: string }
  > = {
    critical: {
      label: 'Critical',
      icon: <AlertOctagon className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-sev-critical text-white border-sev-critical shadow-xs font-semibold',
    },
    high: {
      label: 'High',
      icon: <AlertTriangle className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-sev-high/15 text-sev-high border-sev-high/30',
    },
    medium: {
      label: 'Medium',
      icon: <AlertCircle className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-sev-medium/15 text-sev-medium border-sev-medium/30',
    },
    low: {
      label: 'Low',
      icon: <Info className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-sev-low/15 text-sev-low border-sev-low/30',
    },
    info: {
      label: 'Info',
      icon: <HelpCircle className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-sev-info/15 text-sev-info border-sev-info/30',
    },
    ok: {
      label: 'Passed',
      icon: <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />,
      style: 'bg-ok/15 text-ok border-ok/30',
    },
  };

  const item = config[severity];

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 text-xs rounded-sm border font-medium select-none ${item.style} ${className}`}
      {...props}
    >
      {showIcon ? item.icon : null}
      <span>{children ?? item.label}</span>
      {typeof count === 'number' ? (
        <span className="ml-0.5 px-1 rounded-sm bg-black/15 text-[10px] font-mono leading-none py-0.5">
          {count}
        </span>
      ) : null}
    </span>
  );
};
