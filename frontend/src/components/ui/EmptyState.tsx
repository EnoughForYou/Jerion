import React from 'react';
import { Inbox } from 'lucide-react';

export interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description: string;
  action?: React.ReactNode;
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
  className = '',
}) => {
  return (
    <div
      className={`flex flex-col items-center justify-center p-8 text-center border border-dashed border-separator/40 rounded-lg bg-surface/30 max-w-md mx-auto ${className}`}
    >
      <div className="w-10 h-10 rounded-full bg-raised flex items-center justify-center text-text-secondary mb-3">
        {icon ?? <Inbox className="w-5 h-5" />}
      </div>
      <h4 className="text-md font-semibold text-text-primary mb-1">{title}</h4>
      <p className="text-xs text-text-secondary leading-relaxed mb-4 max-w-xs">{description}</p>
      {action ? <div>{action}</div> : null}
    </div>
  );
};
